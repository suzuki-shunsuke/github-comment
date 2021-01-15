package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Base        Base
	Vars        map[string]interface{}
	Templates   map[string]string
	Post        map[string]PostConfig
	Exec        map[string][]ExecConfig
	SkipNoToken bool `yaml:"skip_no_token"`
	Silent      bool
}

type Base struct {
	Org  string
	Repo string
}

type PostConfig struct {
	Template           string
	TemplateForTooLong string `yaml:"template_for_too_long"`
	HideOldComment     string `yaml:"hide_old_comment"`
}

func (pc *PostConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
		if tpl, ok := m["hide_old_comment"]; ok {
			t, ok := tpl.(string)
			if !ok {
				return fmt.Errorf("invalid config. hide_old_comment should be string: %+v", tpl)
			}
			pc.HideOldComment = t
		}
		return nil
	}
	return fmt.Errorf("invalid config. post config should be string or map[string]intterface{}: %+v", val)
}

type ExecConfig struct {
	When               string
	Template           string
	TemplateForTooLong string `yaml:"template_for_too_long"`
	HideOldComment     string `yaml:"hide_old_comment"`
	DontComment        bool   `yaml:"dont_comment"`
}

type ExistFile func(string) bool

type Reader struct {
	ExistFile ExistFile
}

func (reader Reader) find(wd string) (string, bool) {
	names := []string{".github-comment.yml", ".github-comment.yaml"}
	for {
		for _, name := range names {
			p := filepath.Join(wd, name)
			if reader.ExistFile(p) {
				return p, true
			}
		}
		if wd == "/" || wd == "" {
			return "", false
		}
		wd = filepath.Dir(wd)
	}
}

func (reader Reader) read(p string) (Config, error) {
	cfg := Config{}
	f, err := os.Open(p)
	if err != nil {
		return cfg, fmt.Errorf("open a configuration file "+p+": %w", err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode a configuration file as YAML: %w", err)
	}
	return cfg, nil
}

func (reader Reader) FindAndRead(cfgPath, wd string) (Config, error) {
	cfg := Config{}
	if cfgPath == "" {
		p, b := reader.find(wd)
		if !b {
			return cfg, nil
		}
		cfgPath = p
	}
	return reader.read(cfgPath)
}
