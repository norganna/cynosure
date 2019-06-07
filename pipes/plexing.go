package pipes

import "io"

type plexing struct {
	Observer

	plexes []Piper

	out *plexWriter
	err *plexWriter
}

var _ Piper = (*plexing)(nil)

// NewPlexing returns a pipe that can have multiple sub-pipes, each of which can process data.
func NewPlexing(plexes []Piper) Piper {
	p := &plexing{
		Observer: NewObservation(),

		plexes: plexes,
	}

	p.out = &plexWriter{
		parent: p,
	}
	p.err = &plexWriter{
		parent: p,
	}

	return p
}

func (p *plexing) Err() io.Writer {
	return p.err
}

func (p *plexing) Out() io.Writer {
	return p.out
}

type plexWriter struct {
	parent *plexing
}

func (p *plexWriter) Write(b []byte) (n int, err error) {
	if w, name, line := findListenData(p.parent.Watches(), b); name != "" {
		p.parent.Observation(w, name, line)
	}

	out := p == p.parent.Out()
	for _, pp := range p.parent.plexes {
		if out {
			n, err = pp.Out().Write(b)
		} else {
			n, err = pp.Err().Write(b)
		}
		if err != nil {
			return n, err
		}
	}
	return len(b), nil
}
