package cmd

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func parseVarEnvs() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		if a, ok := strings.CutPrefix(k, "GH_COMMENT_VAR_"); ok {
			m[a] = v
		}
	}
	return m
}

func parseVarsFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2) //nolint:mnd
		if len(a) < 2 {                //nolint:mnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		vars[a[0]] = a[1]
	}
	return vars, nil
}

func parseVarFilesFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2) //nolint:mnd
		if len(a) < 2 {                //nolint:mnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		name := a[0]
		filePath := a[1]
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read the value of the variable %s from the file %s: %w", name, filePath, err)
		}
		vars[name] = string(b)
	}
	return vars, nil
}

func parseVars(c *cli.Command) (map[string]string, error) {
	vars := parseVarEnvs()
	flagVars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return nil, err
	}
	maps.Copy(vars, flagVars)
	varFiles, err := parseVarFilesFlag(c.StringSlice("var-file"))
	if err != nil {
		return nil, err
	}
	maps.Copy(vars, varFiles)
	return vars, nil
}
