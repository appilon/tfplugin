package schema

import "github.com/hashicorp/terraform/helper/schema"

// Copypaste of schema.Provider but removes functions or anything else
// that will fail to serialize
type ProviderDump struct {
	Schema         map[string]*SchemaDump
	ResourcesMap   map[string]*ResourceDump
	DataSourcesMap map[string]*ResourceDump
}

// Converts terraform/helper/schema.Provider to ProviderDump
func NewProviderDump(i *schema.Provider) *ProviderDump {
	p := &ProviderDump{}
	p.Schema = make(map[string]*SchemaDump)
	for key, value := range i.Schema {
		p.Schema[key] = NewSchemaDump(value)
	}
	p.DataSourcesMap = make(map[string]*ResourceDump)
	for key, value := range i.DataSourcesMap {
		p.DataSourcesMap[key] = NewResourceDump(value)
	}
	p.ResourcesMap = make(map[string]*ResourceDump)
	for key, value := range i.ResourcesMap {
		p.ResourcesMap[key] = NewResourceDump(value)
	}
	return p
}

// Copypaste of schema.Resource but removes functions or anything else
// that will fail to serialize
type ResourceDump struct {
	Schema             map[string]*SchemaDump
	SchemaVersion      int
	DeprecationMessage string
	Timeouts           *schema.ResourceTimeout
}

// Converts terraform/helper/schema.Resource to ResourceDump
func NewResourceDump(i *schema.Resource) *ResourceDump {
	r := &ResourceDump{}
	r.SchemaVersion = i.SchemaVersion
	r.DeprecationMessage = i.DeprecationMessage
	r.Timeouts = i.Timeouts
	r.Schema = make(map[string]*SchemaDump)
	for key, value := range i.Schema {
		r.Schema[key] = NewSchemaDump(value)
	}
	return r
}

// Copypaste of schema.Schema but removes functions or anything else
// that will fail to serialize
type SchemaDump struct {
	Type          schema.ValueType
	Optional      bool
	Required      bool
	Default       interface{}
	Description   string
	InputDefault  string
	Computed      bool
	ForceNew      bool
	Elem          interface{}
	MaxItems      int
	MinItems      int
	PromoteSingle bool
	ComputedWhen  []string
	ConflictsWith []string
	Deprecated    string
	Removed       string
	Sensitive     bool
}

// Converts terraform/helper/schema.Schema to SchemaDump
func NewSchemaDump(i *schema.Schema) *SchemaDump {
	s := &SchemaDump{}
	s.Type = i.Type
	s.Optional = i.Optional
	s.Required = i.Required
	s.Default = i.Default
	s.Description = i.Description
	s.InputDefault = i.InputDefault
	s.Computed = i.Computed
	s.ForceNew = i.ForceNew
	s.MaxItems = i.MaxItems
	s.MinItems = i.MinItems
	s.PromoteSingle = i.PromoteSingle
	s.ComputedWhen = i.ComputedWhen
	s.ConflictsWith = i.ConflictsWith
	s.Deprecated = i.Deprecated
	s.Removed = i.Removed
	s.Sensitive = i.Sensitive
	if i.Elem != nil {
		if nestedSchema, ok := i.Elem.(*schema.Schema); ok {
			s.Elem = NewSchemaDump(nestedSchema)
		} else if nestedResource, ok := i.Elem.(*schema.Resource); ok {
			s.Elem = NewResourceDump(nestedResource)
		}
	}
	return s
}
