package local

const DefaultPort = 8001

type ListenerOptions struct {
	Port int
}

func (o *ListenerOptions) setDefault() {
	if o.Port == 0 {
		o.Port = DefaultPort
	}
}

type Listener struct {
	address string
}

func NewListener(options ListenerOptions) {
	panic("not implemented")
	//options.setDefault()

	//return Listener{}
}

// todo rewrite, weird interface Start/Stop
func (l Listener) Start() <-chan Advertisment {
	panic("not implemented")
}

type Advertisment struct {
}
