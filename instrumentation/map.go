package instrumentation

import "regexp"

type MultiEnvDefaultFuncArgs struct {
	KS []string
	DV interface{}
}

type EnvDefaultFuncArgs struct {
	K  string
	DV interface{}
}

type ValidateRFC3339TimeStringArgs struct {
	V interface{}
	K string
}

type ValidateRegexpArgs struct {
	V interface{}
	K string
}

type ValidateListUniqueStringsArgs struct {
	V interface{}
	K string
}

type ValidateJsonStringArgs struct {
	V interface{}
	K string
}

type CIDRNetworkArgs struct {
	Min int
	Max int
}

type NoZeroValuesArgs struct {
	I interface{}
	K string
}

type StringMatchArgs struct {
	R       *regexp.Regexp
	Message string
}

type StringLenBetweenArgs struct {
	Min int
	Max int
}

type StringInSliceArgs struct {
	Valid      []string
	IgnoreCase bool
}

type IntAtMostArgs struct {
	Max int
}

type IntAtLeastArgs struct {
	Min int
}

type IntBetweenArgs struct {
	Min int
	Max int
}

type HelperInfo struct {
	Type   string
	Helper string
	Args   interface{}
}

var HelperInfoMap map[string]*HelperInfo

func init() {
	HelperInfoMap = make(map[string]*HelperInfo)
}
