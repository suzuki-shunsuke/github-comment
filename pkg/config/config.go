package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Base               *Base                    `json:"base,omitempty"`
	GHEBaseURL         string                   `json:"ghe_base_url,omitempty" yaml:"ghe_base_url"`
	GHEGraphQLEndpoint string                   `json:"ghe_graphql_endpoint,omitempty" yaml:"ghe_graphql_endpoint"`
	Vars               map[string]interface{}   `json:"vars,omitempty"`
	Templates          map[string]string        `json:"templates,omitempty"`
	Post               map[string]*PostConfig   `json:"post,omitempty"`
	Exec               map[string][]*ExecConfig `json:"exec,omitempty"`
	Hide               map[string]string        `json:"hide,omitempty"`
	SkipNoToken        bool                     `json:"skip_no_token,omitempty" yaml:"skip_no_token"`
	Silent             bool                     `json:"silent,omitempty"`
}

type Base struct {
	Org  string `json:"org,omitempty"`
	Repo string `json:"repo,omitempty"`
}

type PostConfig struct { //nolint:recvcheck
	Template           string   `json:"template"`
	TemplateForTooLong string   `json:"template_for_too_long,omitempty"`
	EmbeddedVarNames   []string `json:"embedded_var_names,omitempty"`
	// UpdateCondition Update the comment that matches with the condition.
	// If multiple comments match, the latest comment is updated.
	// If no comment matches, a new comment is created.
	UpdateCondition string `json:"update,omitempty"`
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

func (pc *PostConfig) UnmarshalYAML(unmarshal func(interface{}) error) error { //nolint:cyclop
	var val interface{}
	if err := unmarshal(&val); err != nil {
		return err
	}
	if s, ok := val.(string); ok {
		pc.Template = s
		return nil
	}
	if m, ok := val.(map[interface{}]interface{}); ok { //nolint:nestif
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
			t, ok := tpl.([]interface{})
			if !ok {
				return fmt.Errorf("invalid config. embedded_var_names should be []interface{}: %+v", tpl)
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
	When               string   `json:"when"`
	Template           string   `json:"template,omitempty"`
	TemplateForTooLong string   `json:"template_for_too_long,omitempty" yaml:"template_for_too_long"`
	DontComment        bool     `json:"dont_comment,omitempty" yaml:"dont_comment"`
	EmbeddedVarNames   []string `json:"embedded_var_names,omitempty" yaml:"embedded_var_names"`
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

type ExistFile func(string) bool

type Reader struct {
	ExistFile ExistFile
}

func (r *Reader) find(wd string) (string, bool) {
	names := []string{"github-comment.yaml", "github-comment.yml", ".github-comment.yml", ".github-comment.yaml"}
	for {
		for _, name := range names {
			p := filepath.Join(wd, name)
			if r.ExistFile(p) {
				return p, true
			}
		}
		if wd == "/" || wd == "" {
			return "", false
		}
		wd = filepath.Dir(wd)
	}
}

func (r *Reader) read(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("open a configuration file "+p+": %w", err)
	}
	defer f.Close()
	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode a configuration file as YAML: %w", err)
	}
	return cfg, nil
}

const defaultHideCondition = "Comment.HasMeta && Comment.Meta.SHA1 != Commit.SHA1"

func (r *Reader) FindAndRead(cfgPath, wd string) (*Config, error) {
	cfg := &Config{
		Hide: map[string]string{
			"default": defaultHideCondition,
		},
	}
	if cfgPath == "" {
		p, b := r.find(wd)
		if !b {
			return cfg, nil
		}
		cfgPath = p
	}
	cfg, err := r.read(cfgPath) //nolint:ifshort
	if err != nil {
		return nil, err
	}
	if cfg.Hide == nil {
		cfg.Hide = map[string]string{
			"default": defaultHideCondition,
		}
		return cfg, nil
	}
	if _, ok := cfg.Hide["default"]; ok {
		return cfg, nil
	}
	cfg.Hide["default"] = defaultHideCondition
	return cfg, nil
}
