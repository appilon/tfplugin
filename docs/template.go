package docs

import (
	"os"
	"text/template"

	"github.com/appilon/tfdev/schema"
)

var tmpl = `
## Argument Reference

The following arguments are supported:

{{range .Arguments}}
* {{.Name}} - ({{.RequiredOrOptional}}) {{.Description}}
              Default: {{.Default}}
{{end}}

## Attributes Reference

The following attributes are exported:

{{range .Attributes}}
* {{.Name}} - {{.Description}}
{{end}}
`

type AttributeDoc struct {
	Name        string
	Description string
}

type ArgumentDoc struct {
	Name               string
	Description        string
	RequiredOrOptional string
	Default            interface{}
}

type ResourceDoc struct {
	Arguments  []*ArgumentDoc
	Attributes []*AttributeDoc
}

func PrintDoc(resource *schema.ResourceDump) error {
	t, err := template.New("doc").Parse(tmpl)
	if err != nil {
		return err
	}

	data := &ResourceDoc{}
	for name, s := range resource.Schema {
		if s.Computed {
			data.Attributes = append(data.Attributes, &AttributeDoc{
				Name:        name,
				Description: s.Description,
			})
		} else {
			reqOrOpt := "Required"
			if s.Optional {
				reqOrOpt = "Optional"
			}
			data.Arguments = append(data.Arguments, &ArgumentDoc{
				Name:               name,
				Description:        s.Description,
				RequiredOrOptional: reqOrOpt,
				Default:            s.Default,
			})
		}
	}

	t.Execute(os.Stdout, data)
	return nil
}
