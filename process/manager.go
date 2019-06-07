package process

var envList = map[string][]string{}

/*

type ProcessManager interface {
	Create(c Commander) Process
	Get(id string) Process
	List() (list []Process)
	Quit()
	Run()
}

type processManager struct {
	sync.RWMutex

	ch          chan bool
	processList map[string]Process
}

var _ ProcessManager = (*processManager)(nil)

func NewProcessManager() ProcessManager {
	return &processManager{
		ch:          make(chan bool),
		processList: map[string]Process{},
	}
}

func (p *processManager) Create(c Commander) Process {
	process := NewProcess(c)
	p.set(process.ID(), process)
	go process.Loop()

	for i := 0; i < 10 && process.PID() == -1; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	return process
}

func (p *processManager) Get(id string) Process {
	p.RLock()
	defer p.RUnlock()

	if id == "" && len(p.processList) == 1 {
		for id = range p.processList {
			break
		}
	}

	return p.processList[id]
}

func (p *processManager) List() (list []Process) {
	p.RLock()
	defer p.RUnlock()

	for _, process := range p.processList {
		list = append(list, process)
	}

	return list
}

func (p *processManager) Quit() {
	p.Lock()
	defer p.Unlock()

	for _, process := range p.processList {
		process.Close()
	}
	close(p.ch)
}

func (p *processManager) Run() {
	for {
		select {
		case <-p.ch:
			return
		}
	}
}

func (p *processManager) set(id string, process Process) {
	p.Lock()
	defer p.Unlock()

	p.processList[id] = process
}
*/
