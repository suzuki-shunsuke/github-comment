package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/github"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/template"
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

	vars, err := parseVars(c)
	if err != nil {
		return err
	}
	opts.Vars = vars

	return nil
}

func getGitHub(ctx context.Context, opts *option.Options, cfg *config.Config) (api.GitHub, error) {
	if opts.DryRun {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}
	if opts.SkipNoToken && opts.Token == "" {
		return &github.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}, nil
	}

	return github.New(ctx, &github.ParamNew{ //nolint:wrapcheck
		Token:              opts.Token,
		GHEBaseURL:         cfg.GHEBaseURL,
		GHEGraphQLEndpoint: cfg.GHEGraphQLEndpoint,
	})
}

func setLogLevel(logLevel string) {
	if logLevel == "" {
		return
	}
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"log_level": logLevel,
		}).WithError(err).Error("the log level is invalid")
	}
	logrus.SetLevel(lvl)
}

// postAction is an entrypoint of the subcommand "post".
func (runner *Runner) postAction(c *cli.Context) error {
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
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

	var pt api.Platform = platform.Get()

	gh, err := getGitHub(c.Context, &opts.Options, cfg)
	if err != nil {
		return fmt.Errorf("initialize commenter: %w", err)
	}

	ctrl := api.PostController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !term.IsTerminal(0)
		},
		Stdin:  runner.Stdin,
		Stderr: runner.Stderr,
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
