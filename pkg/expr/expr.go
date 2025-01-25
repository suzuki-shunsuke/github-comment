package expr

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Expr struct{}

func (*Expr) Match(expression string, params any) (bool, error) {
	prog, err := expr.Compile(expression, expr.Env(params), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("compile an expression: %s: %w", expression, err)
	}
	output, err := expr.Run(prog, params)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression with params: %s: %w", expression, err)
	}
	if f, ok := output.(bool); !ok || !f {
		return false, nil
	}
	return true, nil
}

type Program interface {
	Run(params any) (bool, error)
}

func (*Expr) Compile(expression string) (Program, error) {
	prog := Prog{}
	prg, err := expr.Compile(expression, expr.AsBool())
	if err != nil {
		return &prog, fmt.Errorf("compile an expression: "+expression+": %w", err)
	}
	prog.prg = prg
	return &prog, nil
}

type Prog struct {
	prg *vm.Program
}

func (p *Prog) Run(params any) (bool, error) {
	output, err := expr.Run(p.prg, params)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression with params: %w", err)
	}
	if f, ok := output.(bool); !ok || !f {
		return false, nil
	}
	return true, nil
}
