package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Base               *Base
	GHEBaseURL         string `yaml:"ghe_base_url"`
	GHEGraphQLEndpoint string `yaml:"ghe_graphql_endpoint"`
	Vars               map[string]interface{}
	Templates          map[string]string
	Post               map[string]*PostConfig
	Exec               map[string][]*ExecConfig
	Hide               map[string]string
	SkipNoToken        bool `yaml:"skip_no_token"`
	Silent             bool
}

type Base struct {
	Org  string
	Repo string
}

type PostConfig struct {
	Template           string
	TemplateForTooLong string
	EmbeddedVarNames   []string
	// UpdateCondition Update the comment that matches with the condition.
	// If multiple comments match, the latest comment is updated
	// If no comment matches, aa new comment is created
	UpdateCondition string
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
	When               string
	Template           string
	TemplateForTooLong string   `yaml:"template_for_too_long"`
	DontComment        bool     `yaml:"dont_comment"`
	EmbeddedVarNames   []string `yaml:"embedded_var_names"`
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
