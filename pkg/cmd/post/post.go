package post

import (
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/github-comment/pkg/cmd/util"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/controller/post"
	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/log"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// parsePostOptions parses the command line arguments of the subcommand "post".
func parsePostOptions(opts *option.PostOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.StdinTemplate = c.Bool("stdin-template")
	opts.LogLevel = c.String("log-level")
	opts.UpdateCondition = c.String("update-condition")
	vars, err := util.ParseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err //nolint:wrapcheck
	}
	varFiles, err := util.ParseVarFilesFlag(c.StringSlice("var-file"))
	if err != nil {
		return err //nolint:wrapcheck
	}
	for k, v := range varFiles {
		vars[k] = v
	}
	opts.Vars = vars
	return nil
}

// postAction is an entrypoint of the subcommand "post".
func (runner *Runner) postAction(c *cli.Context) error {
	if a := runner.osEnv.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}
	opts := &option.PostOptions{}
	if err := parsePostOptions(opts, c); err != nil {
		return err
	}

	log.SetLevel(opts.LogLevel, runner.logE)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	cfgReader := config.Reader{
		ExistFile: util.ExistFile,
	}

	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken

	var pt domain.Platform = platform.Get(util.GetPlatformParam(cfg.Complement))

	gh, err := util.GetGitHub(c.Context, &opts.Options, cfg)
	if err != nil {
		return fmt.Errorf("initialize commenter: %w", err)
	}

	ctrl := &post.Controller{
		Wd:     wd,
		Getenv: runner.osEnv.Getenv,
		HasStdin: func() bool {
			return !term.IsTerminal(0)
		},
		Stdin:  runner.stdio.Stdin,
		Stderr: runner.stdio.Stderr,
		GitHub: gh,
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Platform: pt,
		Config:   cfg,
		Expr:     &expr.Expr{},
	}
	return ctrl.Post(c.Context, opts) //nolint:wrapcheck
}
