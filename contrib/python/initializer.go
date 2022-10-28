package python

import "github.com/hazelcast/hazelcast-commandline-client/internal/plug"

type Initializer struct{}

func (in Initializer) Init(cc plug.InitContext) error {
	cc.AddCommandGroup("python", "Python")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("50-python", &Initializer{})
}
