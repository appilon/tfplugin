// +build exporter

package instrumentation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"regexp"
	"runtime"
	"strings"

	"github.com/appilon/tfdev/trav"
)

func stripImportPath(fnName string) string {
	return fnName[strings.LastIndexByte(fnName, '.')+1:]
}

func CaptureHelper(args ...interface{}) {
	skip := 1
	// get helper function info
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		log.Print("Could not get Helper Function info")
		return
	}
	helper := stripImportPath(runtime.FuncForPC(pc).Name())
	helperType := "ValidateFunc"
	if helper == "EnvDefaultFunc" || helper == "MultiEnvDefaultFunc" {
		helperType = "DefaultFunc"
	}

	// walk callstack/ast to find schema definition
	var fset *token.FileSet
	var nodePath trav.Path
	var calleeNodeIndex int
	callee := helper
	for {
		skip++
		pc, filename, line, ok := runtime.Caller(skip)
		if !ok {
			log.Print("Could not get Helper Function Caller info")
			return
		}
		fset = token.NewFileSet()
		file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			log.Printf("Error parsing source file %s", err)
			return
		}
		// find call, match line numbers
		trav.TraverseNode(file, func(path trav.Path) bool {
			ident, ok := path.Current().(*ast.Ident)
			if ok && ident.Name == callee && fset.Position(ident.Pos()).Line == line {
				nodePath = path.Copy()
				return false
			}
			return true
		})
		// setup next callee search
		callee = stripImportPath(runtime.FuncForPC(pc).Name())
		// ensure this was the schema assignment of the helper
		if calleeNodeIndex = findHelperAssignmentIndex(nodePath, helperType); calleeNodeIndex >= 0 {
			break
		}
	}
	// this is a false assumption but it is our most general case
	// the aws provider more or less is complete like this except for a few unhandled patterns
	// one being s["instance_interruption_behaviour"] = &schema.Schema...
	// resource_aws_spot_instance_request.go:82
	attr, kveIndex := findKVEKeys(nodePath, calleeNodeIndex-1, false, true)
	if attr == "" {
		log.Printf("Could not find attribute name in AST for %s:%s", helperType, helper)
		return
	}

	var resource string
	var lastKvePath trav.Path
	searchStart := kveIndex - 1
	lastKve := -1
	for {
		r, kveIndex := findKVEKeys(nodePath, searchStart, true, false)
		if r != "" {
			resource = r + resource
			lastKve = kveIndex
			lastKvePath = nodePath.Copy()
		}
		// we are at the top of the callstack
		if callee == "Provider" {
			break
		}
		skip++
		pc, filename, line, ok := runtime.Caller(skip)
		if !ok {
			log.Print("Could not walk the callstack")
			return
		}
		fset = token.NewFileSet()
		file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			log.Printf("Error parsing source file %s", err)
			return
		}
		// find call, match line numbers
		trav.TraverseNode(file, func(path trav.Path) bool {
			ident, ok := path.Current().(*ast.Ident)
			if ok && ident.Name == callee && fset.Position(ident.Pos()).Line == line {
				nodePath = path.Copy()
				return false
			}
			return true
		})
		// setup next callee search
		callee = stripImportPath(runtime.FuncForPC(pc).Name())
		searchStart = len(nodePath) - 1
	}
	// should have index on tree close to DataSourceMap/ResourcesMap assignment
	var resourceType string
	if resource == "" {
		resourceType = "[provider]"
	} else {
		resourceType = findResourceType(lastKvePath, lastKve-1)
	}
	storeHelperInfo(resourceType+resource+attr, helperType, helper, args)
}

func findResourceType(path trav.Path, start int) string {
	resourceType := "[error finding type]"
	// walk up ast to find KeyValueExpression
	for i := start; i >= 0; i-- {
		kve, ok := path[i].(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// ensure it is ResourcesMap or DataSourcesMap
		key, ok := kve.Key.(*ast.Ident)
		if ok {
			if key.Name == "ResourcesMap" {
				resourceType = "[resource]"
			} else if key.Name == "DataSourcesMap" {
				resourceType = "[datasource]"
			}
		}
	}
	return resourceType
}

func findHelperAssignmentIndex(path trav.Path, helperType string) int {
	index := -1
	// walk up ast to find KeyValueExpression
	for i := len(path) - 2; i >= 0; i-- {
		kve, ok := path[i].(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// ensure it is helper assignment (in schema definition)
		ident, ok := kve.Key.(*ast.Ident)
		if ok && ident.Name == helperType {
			index = i
			break
		}
	}
	return index
}

func findKVEKeys(path trav.Path, start int, allowElem, exitEarly bool) (string, int) {
	var key string
	index := -1
	for i := start; i >= 0; i-- {
		kve, ok := path[i].(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if allowElem {
			ident, ok := kve.Key.(*ast.Ident)
			if ok && ident.Name == "Elem" {
				key = ".elem" + key
				index = i
				if exitEarly {
					break
				} else {
					continue
				}
			}
		}
		bl, ok := kve.Key.(*ast.BasicLit)
		if !ok {
			continue
		}
		if bl.Kind != token.STRING {
			continue
		}
		key = "." + strings.Trim(bl.Value, `"`) + key
		index = i
		if exitEarly {
			break
		}
	}
	return key, index
}

func storeHelperInfo(key, helperType, helper string, args []interface{}) {
	key += "." + helperType
	HelperInfoMap[key] = &HelperInfo{
		Type:   helperType,
		Helper: helper,
	}

	switch helper {
	case "EnvDefaultFunc":
		HelperInfoMap[key].Args = &EnvDefaultFuncArgs{
			K:  args[0].(string),
			DV: args[1],
		}
	case "MultiEnvDefaultFunc":
		HelperInfoMap[key].Args = &MultiEnvDefaultFuncArgs{
			KS: args[0].([]string),
			DV: args[1],
		}
	case "IntBetween":
		HelperInfoMap[key].Args = &IntBetweenArgs{
			Min: args[0].(int),
			Max: args[1].(int),
		}
	case "IntAtLeast":
		HelperInfoMap[key].Args = &IntAtLeastArgs{
			Min: args[0].(int),
		}
	case "IntAtMost":
		HelperInfoMap[key].Args = &IntAtMostArgs{
			Max: args[0].(int),
		}
	case "StringInSlice":
		HelperInfoMap[key].Args = &StringInSliceArgs{
			Valid:      args[0].([]string),
			IgnoreCase: args[1].(bool),
		}
	case "StringLenBetween":
		HelperInfoMap[key].Args = &StringLenBetweenArgs{
			Min: args[0].(int),
			Max: args[1].(int),
		}
	case "StringMatch":
		HelperInfoMap[key].Args = &StringMatchArgs{
			R:       args[0].(*regexp.Regexp),
			Message: args[1].(string),
		}
	case "NoZeroValues":
		HelperInfoMap[key].Args = &NoZeroValuesArgs{
			I: args[0],
			K: args[1].(string),
		}
	case "CIDRNetwork":
		HelperInfoMap[key].Args = &CIDRNetworkArgs{
			Min: args[0].(int),
			Max: args[1].(int),
		}
	case "SingleIP":
	case "IPRange":
	case "ValidateJsonString":
		HelperInfoMap[key].Args = &ValidateJsonStringArgs{
			V: args[0],
			K: args[1].(string),
		}
	case "ValidateListUniqueStrings":
		HelperInfoMap[key].Args = &ValidateListUniqueStringsArgs{
			V: args[0],
			K: args[1].(string),
		}
	case "ValidateRegexp":
		HelperInfoMap[key].Args = &ValidateRegexpArgs{
			V: args[0],
			K: args[1].(string),
		}
	case "ValidateRFC3339TimeString":
		HelperInfoMap[key].Args = &ValidateRFC3339TimeStringArgs{
			V: args[0],
			K: args[1].(string),
		}
	}
}
