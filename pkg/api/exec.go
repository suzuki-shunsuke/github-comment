package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/antonmedv/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

type Env struct {
	Stdout         string
	Stderr         string
	CombinedOutput string
	Command        string
	ExitCode       int
	Env            func(string) string
}

func Exec(
	ctx context.Context, wd string, opts *option.ExecOptions,
	getEnv func(string) string, existFile func(string) bool,
	readConfig func(string, *config.Config) error,
) error {
	if option.IsCircleCI(getEnv) {
		if err := option.ComplementExec(opts, getEnv); err != nil {
			return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if err := option.ValidateExec(opts); err != nil {
		return err
	}

	cfg := &config.Config{}
	if opts.ConfigPath == "" {
		p, b, err := config.Find(wd, existFile)
		if err != nil {
			return err
		}
		if !b {
			return errors.New("configuration file isn't found")
		}
		opts.ConfigPath = p
	}

	if err := readConfig(opts.ConfigPath, cfg); err != nil {
		return err
	}

	execConfigs, ok := cfg.Exec[opts.TemplateKey]
	if !ok {
		return errors.New("template isn't found: " + opts.TemplateKey)
	}

	cmd := exec.Command(opts.Args[0], opts.Args[1:]...)
	cmd.Stdin = os.Stdin
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	combinedOutput := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(os.Stdout, stdout, combinedOutput)
	cmd.Stderr = io.MultiWriter(os.Stderr, stderr, combinedOutput)
	cmd.Env = os.Environ()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)

	runner := timeout.NewRunner(0)
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	sentSignals := map[os.Signal]struct{}{}
	exitChan := make(chan error, 1)

	go func() {
		exitChan <- runner.Run(c, cmd)
	}()

	for {
		select {
		case err := <-exitChan:
			if err != nil {
				return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
			}
			return execPost(c, opts, execConfigs, &Env{
				ExitCode:       cmd.ProcessState.ExitCode(),
				Command:        cmd.String(),
				Stdout:         stdout.String(),
				Stderr:         stderr.String(),
				CombinedOutput: combinedOutput.String(),
			})
		case sig := <-signalChan:
			if _, ok := sentSignals[sig]; ok {
				continue
			}
			sentSignals[sig] = struct{}{}
			runner.SendSignal(sig.(syscall.Signal))
		}
	}
}

func execPostConfig(
	ctx context.Context, opts *option.ExecOptions, execConfig *config.ExecConfig, env *Env,
) (bool, error) {
	e := expr.Env(env)
	prog, err := expr.Compile(execConfig.When, e, expr.AsBool())
	if err != nil {
		return false, err
	}
	output, err := expr.Run(prog, env)
	if err != nil {
		return false, err
	}
	if f, ok := output.(bool); ok && f {
		if execConfig.DontComment {
			return true, nil
		}
		tmpl, err := template.New("comment").Funcs(template.FuncMap{
			"Env": os.Getenv,
		}).Parse(execConfig.Template)
		if err != nil {
			return true, err
		}
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, env); err != nil {
			return true, err
		}
		client := &http.Client{}
		cmt := &comment.Comment{
			PRNumber: opts.PRNumber,
			Org:      opts.Org,
			Repo:     opts.Repo,
			Body:     buf.String(),
			SHA1:     opts.SHA1,
		}
		if err := comment.Create(ctx, client, opts.Token, cmt); err != nil {
			return true, fmt.Errorf("failed to create an issue comment: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func execPost(ctx context.Context, opts *option.ExecOptions, execConfigs []config.ExecConfig, env *Env) error {
	for _, execConfig := range execConfigs {
		f, err := execPostConfig(ctx, opts, &execConfig, env)
		if err != nil {
			return err
		}
		if f {
			return nil
		}
	}
	return nil
}
