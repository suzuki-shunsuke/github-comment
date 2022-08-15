package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/execute"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/urfave/cli/v2"
)

func parseExecOptions(opts *option.ExecOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.Args = c.Args().Slice()
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.LogLevel = c.String("log-level")

	vars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err
	}
	varFiles, err := parseVarFilesFlag(c.StringSlice("var-file"))
	if err != nil {
		return err
	}
	for k, v := range varFiles {
		vars[k] = v
	}
	opts.Vars = vars

	return nil
}

func existFile(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (runner *Runner) execAction(c *cli.Context) error {
	opts := &option.ExecOptions{}
	if err := parseExecOptions(opts, c); err != nil {
		return err
	}
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		opts.SkipComment = skipComment
	}
	setLogLevel(opts.LogLevel)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}
	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken
	opts.Silent = opts.Silent || cfg.Silent

	var pt api.Platform = platform.Get(getPlatformParam(cfg.Complement))

	gh, err := getGitHub(c.Context, &opts.Options, cfg)
	if err != nil {
		return fmt.Errorf("initialize commenter: %w", err)
	}

	ctrl := api.ExecController{
		Wd:     wd,
		Getenv: os.Getenv,
		Stdin:  runner.Stdin,
		Stdout: runner.Stdout,
		Stderr: runner.Stderr,
		GitHub: gh,
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Executor: &execute.Executor{
			Stdout: runner.Stdout,
			Stderr: runner.Stderr,
			Env:    os.Environ(),
		},
		Expr:     &expr.Expr{},
		Platform: pt,
		Config:   cfg,
	}
	return ctrl.Exec(c.Context, opts) //nolint:wrapcheck
}
