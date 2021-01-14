package expr

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expr struct{}

func (Expr) Match(expression string, params interface{}) (bool, error) {
	prog, err := expr.Compile(expression, expr.Env(params), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("compile an expression: "+expression+": %w", err)
	}
	output, err := expr.Run(prog, params)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression with params: "+expression+": %w", err)
	}
	if f, ok := output.(bool); !ok || !f {
		return false, nil
	}
	return true, nil
}

type Program interface {
	Run(params interface{}) (bool, error)
}

func (Expr) Compile(expression string) (Program, error) {
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

func (prog *Prog) Run(params interface{}) (bool, error) {
	output, err := expr.Run(prog.prg, params)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression with params: %w", err)
	}
	if f, ok := output.(bool); !ok || !f {
		return false, nil
	}
	return true, nil
}
