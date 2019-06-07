package pipes

import (
	"io"
	"os"
)

type standard struct {
	Observer

	err *throughProcessor
	out *throughProcessor
}

var _ Piper = (*standard)(nil)

// NewStandard returns a Piper which sends data to StdOut/StdErr.
func NewStandard() Piper {
	p := &standard{
		Observer: NewObservation(),
	}
	p.err = &throughProcessor{
		parent: p,
		out:    os.Stderr,
	}
	p.out = &throughProcessor{
		parent: p,
		out:    os.Stdout,
	}
	return p
}

func (p *standard) Err() io.Writer {
	return p.err
}

func (p *standard) Out() io.Writer {
	return p.out
}
