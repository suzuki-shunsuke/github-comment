package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Post map[string]string
	Exec map[string][]ExecConfig
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

func (reader Reader) Find(wd string) (string, bool, error) {
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

func (reader Reader) Read(p string, cfg *Config) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(cfg)
}
