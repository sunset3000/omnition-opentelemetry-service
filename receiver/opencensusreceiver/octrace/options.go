package octrace

// Option interface defines for configuration settings to be applied to receivers.
//
// WithReceiver applies the configuration to the given receiver.
type Option func(*Receiver)

// WithBackPressure is used to enable the server to return backpressure to
// its callers.
func WithBackPressure() Option {
	return func(r *Receiver) {
		r.backPressureOn = true
	}
}

// WithMaxServerStream allows one to specify the options for starting a gRPC server.
func WithMaxServerStream(maxServerStreams int64) Option {
	return func(r *Receiver) {
		r.maxServerStreams = maxServerStreams
	}
}
