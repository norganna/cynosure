// Package kube is a provider that checks running instances in a kubernetes cluster.
//
// Accepts config:
//    type: "in-cluster", "json" or "config"
//    config: PATH
//    file: PATH
//
// If type is "in-cluster", will use the in-cluster credentials to connect to the cluster.
//
// If type is "config", will use the kube config file at the specified PATH.
package kube

import (
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "kube"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(config common.StringMap) (deps.Broker, error) {
	b := &broker{}

	return b, nil
}

type broker struct {
}

func (b *broker) Dep(wait string) (deps.Depender, error) {
	return &dep{
		b:    b,
		wait: wait,
	}, nil
}

type dep struct {
	b    *broker
	wait string
}

func (d *dep) Check() (msg string, ok bool) {
	panic("not implemented")
}

/*

type kubeManager struct {
	sync.RWMutex

	clients map[string]*kubeClient
}

func (m *kubeManager) GetClient(config string) (client *kubeClient, err error) {
	func () {
		m.RLock()
		defer m.RUnlock()

		if c, ok := m.clients[config]; ok {
			client = c
		}
	}()

	if client != nil {
		return client, nil
	}

	m.Lock()
	defer m.Unlock()

	var c *rest.Config
	var data []byte

	if config == "in-cluster" {
		c, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		if strings.HasPrefix(config, "file=") {
			kubeConfig := config[5:]
			data, err = ioutil.ReadFile(kubeConfig)
			if err != nil {
				return nil, common.Error(err, "failed to read kube config file")
			}
		} else if strings.HasPrefix(config, "json=") {
			data = []byte(config[5:])
		} else {
			return nil, deps.ErrUnknownConfig
		}

		c, err = clientcmd.RESTConfigFromKubeConfig(data)
		if err != nil {
			return nil, common.Error(err, "failed to load kube config")
		}
	}

	cs, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, common.Error(err, "failed to create kube client set from config")
	}

	client = &kubeClient{
		cs:   cs,
		deps: map[string]*kubeDeployments{},
	}

	m.clients[config] = client
	return client, nil
}

type kubeClient struct {
	sync.RWMutex

	cs   *kubernetes.Clientset
	deps map[string]*kubeDeployments
}

func (c *kubeClient) GetDeployments(ns string) (deps *kubeDeployments, err error) {
	now := time.Now()

	func () {
		c.RLock()
		defer c.RUnlock()

		if d, ok := c.deps[ns]; ok {
			if now.Sub(d.ts) <= 10*time.Second {
				deps = d
			}
		}
	}()

	if deps != nil {
		return deps, nil
	}

	c.Lock()
	defer c.Unlock()

	deps = &kubeDeployments{
		ts: now,
	}

	deps.list, err = c.cs.AppsV1().Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, common.Error(err, "failed to get kube deployment list")
	}

	c.deps[ns] = deps
	return deps, nil
}

type kubeDeployments struct {
	ts time.Time
	list *appsv1.DeploymentList
}

var kubeMan = &kubeManager{
	clients: map[string]*kubeClient{},
}

type kube struct {
	client   *kubeClient
	deps     []*ReqNsItem
	messages []string
}

var _ Depender = (*kube)(nil)

func NewKube(config string) (Depender, error) {
	client, err := kubeMan.GetClient(config)
	if err != nil {
		return nil, err
	}

	k := &kube{
		client: client,
	}
	return k, nil
}

func (d *kube) Check() bool {
	if len(d.deps) == 0 {
		return true
	}



	pods, err := d.cs.AppsV1().Deployments(d.ns).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	waiting := map[string]string{}

depLoop:
	for _, dep := range d.deps {
		for _, pod := range pods.Items {
			if pod.Name == dep {
				if pod.Status.ReadyReplicas > 0 {
					continue depLoop
				}
				waiting[dep] = "STARTING"
				continue depLoop
			}
		}
		waiting[dep] = "PENDING"
	}

	var messages []string
	for _, dep := range d.deps {
		if state, exists := waiting[dep]; exists {
			messages = append(messages, fmt.Sprintf("%s [%s]", dep, state))
			continue
		}
		messages = append(messages, fmt.Sprintf("\n  %s [OK]", dep))
	}

	d.messages = messages

	if len(waiting) > 0 {
		return false
	}
	return true
}

func (d *kube) Messages() []string {
	return d.messages
}

func (d *kube) Add(items ...interface{}) error {
	for _, item := range items {
		dep, ok := item.(string)
		if !ok {
			return ErrIncorrectType
		}
		d.deps = append(d.deps, dep)
	}
	return nil
}

*/
