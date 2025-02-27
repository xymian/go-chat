package tracer

import (
	"fmt"
	"io"
	"os"
)

type Tracer interface {
	Trace(a...interface{})
}

type eventTracer struct {
	out io.Writer
}

func (t *eventTracer) Trace(a...interface{}) {
	fmt.Fprintln(t.out, a...)
}

func New() Tracer {
	return &eventTracer{
		out: os.Stdout,
	}
}