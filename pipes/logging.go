package pipes

import (
	"io"
	"sort"
	"sync"
	"time"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/proto/cynosure"
)

type logging struct {
	Observer
	sync.RWMutex

	count int64
	lines []Liner

	parser func(*line)

	err *processor
	out *processor
}

var _ Piper = (*logging)(nil)
var _ Logger = (*logging)(nil)

// NewLogging returns a Logger which tracks log lines in memory.
func NewLogging() Logger {
	p := &logging{
		Observer: NewObservation(),
	}

	p.err = &processor{
		cb:     p.Add,
		class:  "err",
		parent: p,
	}
	p.out = &processor{
		cb:     p.Add,
		class:  "out",
		parent: p,
	}

	return p
}

func (p *logging) Add(class, msg string) {
	p.Lock()
	defer p.Unlock()

	p.count++
	line := &line{
		line:    p.count,
		class:   class,
		ts:      time.Now(),
		message: msg,
	}
	if p.parser != nil {
		p.parser(line)
	}
	p.lines = append(p.lines, line)
}

func (p *logging) Count() int64 {
	return p.count
}

func (p *logging) Err() io.Writer {
	return p.err
}

func (p *logging) Head(n int) (lines []Liner, count int64) {
	p.RLock()
	defer p.RUnlock()

	lines = p.lines
	if n < len(lines) {
		lines = lines[:n]
	}
	return lines, p.count
}

func (p *logging) Out() io.Writer {
	return p.out
}
func (p *logging) Since(t time.Time) (lines []Liner, count int64) {
	p.RLock()
	defer p.RUnlock()

	pos := sort.Search(len(p.lines), func(i int) bool {
		return p.lines[i].Time().After(t)
	})
	return p.lines[pos:], p.count
}

func (p *logging) Tail(n int) (lines []Liner, count int64) {
	p.RLock()
	defer p.RUnlock()

	lines = p.lines
	if n < len(lines) {
		lines = lines[len(lines)-n:]
	}
	return lines, p.count
}

type line struct {
	line    int64
	class   string
	ts      time.Time
	message string
	fields  common.ObjMap
}

var _ Liner = (*line)(nil)

func (l *line) Class() string {
	return l.class
}

func (l *line) Fields() common.ObjMap {
	return l.fields
}

func (l *line) Message() string {
	return l.message
}

func (l *line) Time() time.Time {
	return l.ts
}

func (l *line) Entry() *cynosure.LogEntry {
	return &cynosure.LogEntry{
		Pos:     l.line,
		Time:    l.ts.UTC().String(),
		Source:  l.class,
		Message: l.message,
		Fields:  l.fields.JSON(),
	}
}
