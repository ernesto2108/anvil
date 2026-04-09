package instrumentation

// NoopEmitter is a no-op implementation of Emitter for tests and no-dashboard mode.
type NoopEmitter struct{}

func (n *NoopEmitter) Emit(_ Event) {}
func (n *NoopEmitter) Start()       {}
func (n *NoopEmitter) Stop()        {}
