package pipes

import (
	"bytes"
	"io"
	"time"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/proto/cynosure"
)

// MaxLineLength is the line length at which we will split a log line.
const MaxLineLength = 1024 * 1024 * 2 // 2MiB

// Default outputs to StdOut/StdErr and is used if no specific piper is supplied.
var Default Piper = NewStandard()

// Observer provides log observation functionality to a piper.
type Observer interface {
	AddWatch(name string, watch *cynosure.Watch) bool
	Clear()
	Observation(watch *cynosure.Watch, name, line string)
	Observed() map[string]string
	Ready() bool
	Watches() map[string]*cynosure.Watch
}

// Piper can return an output and error stream.
type Piper interface {
	Observer

	Err() io.Writer
	Out() io.Writer
}

// Logger is a Piper that can process lines as log entries.
type Logger interface {
	Piper

	Count() int64
	Head(n int) (lines []Liner, count int64)
	Since(t time.Time) (lines []Liner, count int64)
	Tail(n int) (lines []Liner, count int64)
}

// Liner is a log line entry.
type Liner interface {
	Class() string
	Fields() common.ObjMap
	Message() string
	Time() time.Time

	Entry() *cynosure.LogEntry
}

func findListenData(watches map[string]*cynosure.Watch, line []byte) (watch *cynosure.Watch, name, found string) {
	for name, watch := range watches {
		if bytes.Contains(line, []byte(watch.Match)) {
			return watch, name, string(line)
		}
	}
	return nil, "", ""
}
