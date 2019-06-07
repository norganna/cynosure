package deps

// Broker can get a Dep from a wait string.
type Broker interface {
	Dep(wait string) (Depender, error)
}
