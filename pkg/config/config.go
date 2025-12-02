package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type Config struct {
	Base               *Base                    `json:"base,omitempty" jsonschema:"description=Repository where to post comments"`
	GHEBaseURL         string                   `json:"ghe_base_url,omitempty" yaml:"ghe_base_url" jsonschema:"description=GitHub Enterprise Base URL"`
	GHEGraphQLEndpoint string                   `json:"ghe_graphql_endpoint,omitempty" yaml:"ghe_graphql_endpoint" jsonschema:"description=GitHub Enterprise GraphQL Endpoint"`
	Vars               map[string]any           `json:"vars,omitempty" jsonschema:"description=variables to pass to templates"`
	Templates          map[string]string        `json:"templates,omitempty" jsonschema:"description=templates"`
	Post               map[string]*PostConfig   `json:"post,omitempty" jsonschema:"description=configuration for github-comment post command"`
	Exec               map[string][]*ExecConfig `json:"exec,omitempty" jsonschema:"description=configuration for github-comment exec command"`
	Hide               map[string]string        `json:"hide,omitempty" jsonschema:"description=configuration for github-comment hide command"`
	SkipNoToken        bool                     `json:"skip_no_token,omitempty" yaml:"skip_no_token" jsonschema:"description=Skip to post comments if no GitHub access token is passed"`
	Silent             bool                     `json:"silent,omitempty"`
}

type Base struct {
	Org  string `json:"org,omitempty" jsonschema:"description=GitHub organization name"`
	Repo string `json:"repo,omitempty" jsonschema:"description=GitHub repository name"`
}

type PostConfig struct { //nolint:recvcheck
	Template           string   `json:"template" jsonschema:"description=Comment template"`
	TemplateForTooLong string   `json:"template_for_too_long,omitempty"`
	EmbeddedVarNames   []string `json:"embedded_var_names,omitempty" jsonschema:"description=Embedded variable names"`
	// UpdateCondition Update the comment that matches with the condition.
	// If multiple comments match, the latest comment is updated.
	// If no comment matches, a new comment is created.
	UpdateCondition string `json:"update,omitempty" jsonschema:"description=Update comments that matches with the condition"`
}

type postConfigForJS PostConfig

func (PostConfig) JSONSchema() *jsonschema.Schema {
	a := jsonschema.Reflect(&postConfigForJS{}).Definitions["postConfigForJS"]
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:       "string",
				Deprecated: true,
			},
			a,
		},
	}
}

func (pc *PostConfig) UnmarshalYAML(unmarshal func(any) error) error { //nolint:cyclop
	var val any
	if err := unmarshal(&val); err != nil {
		return err
	}
	if s, ok := val.(string); ok {
		pc.Template = s
		return nil
	}
	if m, ok := val.(map[any]any); ok { //nolint:nestif
		if tpl, ok := m["template"]; ok {
			t, ok := tpl.(string)
			if !ok {
				return fmt.Errorf("invalid config. template should be string: %+v", tpl)
			}
			pc.Template = t
		}
		if tpl, ok := m["template_for_too_long"]; ok {
			t, ok := tpl.(string)
			if !ok {
				return fmt.Errorf("invalid config. template_for_too_long should be string: %+v", tpl)
			}
			pc.TemplateForTooLong = t
		}
		if tpl, ok := m["embedded_var_names"]; ok {
			t, ok := tpl.([]any)
			if !ok {
				return fmt.Errorf("invalid config. embedded_var_names should be []any: %+v", tpl)
			}
			names := make([]string, len(t))
			for i, name := range t {
				s, ok := name.(string)
				if !ok {
					return fmt.Errorf("invalid config. embedded_var_names[%d] should be string: %+v", i, name)
				}
				names[i] = s
			}
			pc.EmbeddedVarNames = names
		}
		if tpl, ok := m["update"]; ok {
			t, ok := tpl.(string)
			if !ok {
				return fmt.Errorf("invalid config. update should be string: %+v", tpl)
			}
			pc.UpdateCondition = t
		}
		return nil
	}
	return fmt.Errorf("invalid config. post config should be string or map[string]intterface{}: %+v", val)
}

type ExecConfig struct {
	When               string   `json:"when" jsonschema:"description=Condition that this setting is chosen"`
	Template           string   `json:"template,omitempty" jsonschema:"description=Comment template"`
	TemplateForTooLong string   `json:"template_for_too_long,omitempty" yaml:"template_for_too_long"`
	DontComment        bool     `json:"dont_comment,omitempty" yaml:"dont_comment" jsonschema:"description=Don't post a comment"`
	EmbeddedVarNames   []string `json:"embedded_var_names,omitempty" yaml:"embedded_var_names" jsonschema:"description=Embedded variable names"`
	// UpdateCondition Update the comment that matches with the condition.
	// If multiple comments match, the latest comment is updated.
	// If no comment matches, a new comment is created.
	UpdateCondition string `json:"update,omitempty" yaml:"update" jsonschema:"description=Update comments that matches with the condition"`
}

func (ec ExecConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Properties.Set("when", &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			{
				Type: "boolean",
			},
		},
	})
}

const defaultHideCondition = "Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1"
