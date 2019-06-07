package process

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/NorgannasAddOns/go-uuid"
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
	"github.com/norganna/cynosure/pipes"
	"github.com/norganna/cynosure/proto/cynosure"
	psNet "github.com/shirou/gopsutil/net"
	psProcess "github.com/shirou/gopsutil/process"
)

// Processor allows lifecycle management of a process instance.
type Processor interface {
	Close()
	ID() string
	Loop()

	Process() *cynosure.Process
	Log() pipes.Logger
	PID() int
}

type proc struct {
	identity     string
	namespace    string
	environments []string

	ch chan bool
	c  *cynosure.Command

	cmd   *exec.Cmd
	deps  deps.DepList
	pipes pipes.Piper

	prevMsg string

	inc        time.Duration
	delay      time.Duration
	minDelay   time.Duration
	maxDelay   time.Duration
	resetAfter time.Duration

	started int64
}

var _ Processor = (*proc)(nil)

// NewProcess creates a new Processor.
func NewProcess(ns string, envs []string, c *cynosure.Command) Processor {
	return &proc{
		identity:     c.GetName() + "-" + uuid.New("p"),
		namespace:    ns,
		environments: envs,

		ch: make(chan bool),
		c:  c,

		inc:        500 * time.Millisecond,
		minDelay:   1 * time.Second,
		maxDelay:   30 * time.Second,
		resetAfter: 60 * time.Second,
	}
}

func (p *proc) Close() {
	close(p.ch)

	if pid := p.PID(); pid > 0 {
		// Send it a soft kill notification.
		err := syscall.Kill(pid, syscall.SIGINT)
		if err != nil {
			return
		}

		// Wait for up to 10 seconds for it to stop, then force kill it.
		go func() {
			for i := 0; i < 20; i++ {
				time.Sleep(500 * time.Millisecond)
				if p.PID() < 1 {
					return
				}
			}
			err = syscall.Kill(pid, syscall.SIGKILL)
		}()
	}
}

func (p *proc) Cmd() *exec.Cmd {
	var piper pipes.Piper = p.Log()
	if piper == nil {
		piper = pipes.Default
	}

	c := p.c

	cmd := exec.Command(c.GetEntry(), c.GetArgs()...)
	cmd.Args[0] = c.Name
	cmd.Stdout = piper.Out()
	cmd.Stderr = piper.Err()

	var envs [][]string
	for _, name := range p.environments {
		if e, ok := envList[name]; ok {
			envs = append(envs, e)
		}
	}
	envs = append(envs, c.Env)

	cmd.Env = buildEnv(envs)
	return cmd
}

func (p *proc) Deps() (deps.DepList, error) {
	if p.deps != nil {
		return p.deps, nil
	}

	depMap := deps.DepList{}
	for n, dd := range p.c.Requirements {
		for _, d := range dd.Deps {
			b := deps.Instance(d.Identity, p.namespace)
			dep, err := b.Dep(d.Wait)
			if err != nil {
				return nil, common.Error(err, "failed to build requirements %s/%s", n, d.Identity)
			}

			depMap[n] = append(depMap[n], dep)
		}
	}

	p.deps = depMap
	return depMap, nil
}

func (p *proc) ID() string {
	return p.identity
}

func (p *proc) Loop() {
	p.delay = p.minDelay

	for {
		select {
		case <-p.ch:
			return
		default:
			p.tryRun()
		}
	}
}

func (p *proc) Process() *cynosure.Process {
	started := p.Started()
	var lines int64
	if logging := p.Log(); logging != nil {
		lines = logging.Count()
	}
	now := time.Now().UnixNano() / int64(time.Millisecond)
	process := &cynosure.Process{
		Identifier: p.ID(),
		Namespace:  p.namespace,
		Pid:        int32(p.PID()),
		Started:    started,
		Running:    now - started,
		Ready:      false,
		Command: &cynosure.Command{
			Name:         p.c.GetName(),
			Image:        p.c.GetImage(),
			Entry:        p.cmd.Path,
			Args:         p.cmd.Args,
			Env:          p.cmd.Env,
			Requirements: p.c.GetRequirements(),
			Lines:        lines,
		},
		Ports:        p.Ports(),
		Observations: p.pipes.Observed(),
	}

	return process
}

func (p *proc) Log() pipes.Logger {
	if l, ok := p.pipes.(pipes.Logger); ok {
		return l
	}
	return nil
}

func (p *proc) PID() int {
	if cmd := p.cmd; cmd != nil {
		if proc := cmd.Process; proc != nil {
			if pid := proc.Pid; pid > 0 {
				if pp, err := os.FindProcess(pid); err == nil && pp != nil {
					return pid
				}
			} else {
				return pid
			}
		}
	}
	return -1
}

func (p *proc) Ports() []string {
	ps, err := psProcess.NewProcess(int32(p.PID()))
	if err != nil {
		return nil
	}

	conns, err := ps.Connections()
	if err != nil {
		return nil
	}

	list := make([]string, 0)
	for _, conn := range conns {
		if conn.Status == "LISTEN" {
			list = append(list, address(conn.Laddr))
		}
	}
	return list
}

func (p *proc) Started() int64 {
	return p.started
}

func (p *proc) tryRun() error {
	d, err := p.Deps()
	if err != nil {
		return common.Error(err, "failed checking deps")
	}

	mm, ok := d.Check()
	checkMsg := strings.Join(mm, "\n - ")
	if checkMsg != p.prevMsg {
		p.prevMsg = checkMsg
		_, _ = p.Log().Out().Write([]byte("Requirements:\n - " + checkMsg))
	}

	if ok {
		startTime := time.Now()

		fmt.Printf("Executing: %s\n", strings.Join(p.cmd.Args, " "))
		p.started = startTime.UnixNano() / int64(time.Millisecond)

		p.pipes.Clear()
		err := p.cmd.Run()

		p.started = 0
		if err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				if exit.Success() {
					p.delay = p.minDelay
				} else {
					fmt.Printf("Process exited with error\n")
				}

				if time.Now().Sub(startTime) > p.resetAfter {
					p.delay = p.minDelay
				}
			} else {
				fmt.Printf("Error running command: %#v\n", err.Error())
				os.Exit(1)
			}
		} else if elapsed := time.Now().Sub(startTime); elapsed > p.resetAfter {
			p.delay = p.minDelay
		}
	}

	time.Sleep(p.delay)
	p.delay += p.inc
	if p.delay > p.maxDelay {
		p.delay = p.maxDelay
	}

	return nil
}

func buildEnv(envs [][]string) []string {
	env := map[string]string{}

	var items []string
	items = append(items, os.Environ()...)
	for _, e := range envs {
		items = append(items, e...)
	}

	for _, e := range items {
		s := strings.SplitN(e, "=", 2)
		env[s[0]] = e
	}

	var keys []string
	for k := range env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	list := make([]string, len(keys))
	for i, k := range keys {
		list[i] = env[k]
	}

	return list
}

func address(a psNet.Addr) string {
	ip := a.IP
	if ip == "*" {
		ip = ""
	}
	return fmt.Sprintf("%s:%d", ip, a.Port)
}
