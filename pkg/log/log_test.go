package log_test

import (
	"testing"

	"github.com/suzuki-shunsuke/github-comment/pkg/log"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if logE := log.New("v3.0.0"); logE == nil {
		t.Fatal("logE must not be nil")
	}
}

func TestSetLevel(t *testing.T) {
	t.Parallel()
	logE := log.New("v3.0.0")
	log.SetLevel("debug", logE)
}
