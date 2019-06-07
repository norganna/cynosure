package pipes

import (
	"bytes"
	"io"
	"sync"
)

type processor struct {
	sync.Mutex

	parent Piper

	class string
	buf   []byte
	cb    func(class, message string)
}

var _ io.Writer = (*processor)(nil)

func (l *processor) Write(b []byte) (n int, err error) {
	l.Lock()
	defer l.Unlock()

	if w, name, line := findListenData(l.parent.Watches(), b); name != "" {
		l.parent.Observation(w, name, line)
	}

	l.buf = append(l.buf, b...)
	lines := bytes.Split(l.buf, []byte{10})
	n = len(lines)
	last := lines[n-1]
	if len(last) > MaxLineLength {
		l.buf = append([]byte("…"), last[MaxLineLength:]...)
		lines[n-1] = append(last[:MaxLineLength], []byte("…")...)
	} else {
		l.buf = last
		lines = lines[:n-1]
	}
	for _, line := range lines {
		l.cb(l.class, string(line))
	}
	return len(b), nil
}

type throughProcessor struct {
	parent Piper
	out    io.Writer
}

var _ io.Writer = (*throughProcessor)(nil)

func (t *throughProcessor) Write(b []byte) (n int, err error) {
	if w, name, line := findListenData(t.parent.Watches(), b); name != "" {
		t.parent.Observation(w, name, line)
	}

	return t.out.Write(b)
}
