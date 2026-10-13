package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Reader struct {
	ExistFile ExistFile
}

type ExistFile func(string) bool

func (r *Reader) FindAndRead(cfgPath, wd string) (*Config, error) {
	cfg := &Config{}
	if cfgPath == "" {
		p, b := r.find(wd)
		if !b {
			setDefaultConditions(cfg)
			return cfg, nil
		}
		cfgPath = p
	}
	cfg, err := r.read(cfgPath)
	if err != nil {
		return nil, err
	}
	setDefaultConditions(cfg)
	return cfg, nil
}

// setDefaultConditions sets the default condition of the hide and delete
// commands if they aren't set.
func setDefaultConditions(cfg *Config) {
	if cfg.Hide == nil {
		cfg.Hide = map[string]string{}
	}
	if _, ok := cfg.Hide["default"]; !ok {
		cfg.Hide["default"] = defaultHideCondition
	}
	if cfg.Delete == nil {
		cfg.Delete = map[string]string{}
	}
	if _, ok := cfg.Delete["default"]; !ok {
		cfg.Delete["default"] = defaultDeleteCondition
	}
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
