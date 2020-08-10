package fsys

import (
	"io/ioutil"
	"os"
)

type Fsys struct{}

func (fsys Fsys) Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (fsys Fsys) Write(path string, content []byte) error {
	return ioutil.WriteFile(path, content, 0o755) //nolint:gosec
}
