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

func Find(wd string, existFile ExistFile) (string, bool, error) {
	names := []string{".github-comment.yml", ".github-comment.yaml"}
	for {
		for _, name := range names {
			p := filepath.Join(wd, name)
			if existFile(p) {
				return p, true, nil
			}
		}
		if wd == "/" || wd == "" {
			return "", false, nil
		}
		wd = filepath.Dir(wd)
	}
}

func Read(p string, cfg *Config) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(cfg)
}
