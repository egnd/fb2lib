package interfaces

// Logger is a logger interface.
type Logger interface {
	Errorf(error, string, ...interface{})
	Infof(string, ...interface{})
}
