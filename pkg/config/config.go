package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Base Base
	Post map[string]string
	Exec map[string][]ExecConfig
}

type Base struct {
	Org  string
	Repo string
}

type ExecConfig struct {
	When        string
	Template    string
	DontComment bool `yaml:"dont_comment"`
}

type ExistFile func(string) bool

type Reader struct {
	ExistFile ExistFile
}

func (reader Reader) find(wd string) (string, bool, error) {
	names := []string{".github-comment.yml", ".github-comment.yaml"}
	for {
		for _, name := range names {
			p := filepath.Join(wd, name)
			if reader.ExistFile(p) {
				return p, true, nil
			}
		}
		if wd == "/" || wd == "" {
			return "", false, nil
		}
		wd = filepath.Dir(wd)
	}
}

func (reader Reader) read(p string) (Config, error) {
	cfg := Config{}
	f, err := os.Open(p)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (reader Reader) FindAndRead(cfgPath, wd string) (Config, error) {
	cfg := Config{}
	if cfgPath == "" {
		p, b, err := reader.find(wd)
		if err != nil {
			return cfg, err
		}
		if !b {
			return cfg, errors.New("configuration file isn't found")
		}
		cfgPath = p
	}
	return reader.read(cfgPath)
}
