package tracer

import (
	"fmt"
	"io"
)

type Tracer interface {
	Trace(a...interface{})
}

type EventTracer struct {
	Out io.Writer
}

func (t *EventTracer) Trace(a...interface{}) {
	fmt.Fprint(t.Out, a...)
	fmt.Fprintln(t.Out)
}