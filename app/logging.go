package app

// logger provides access to Host Callbacks that interact with the logging system within Tarmac. The callbacks
// within logger provided all of the logic and error handlings of accessing and interacting with a logger.
type logger struct{}

// Info will take the incoming byte slice data and call the internal Tarmac logger converting the data to a string.
func (l *logger) Info(b []byte) ([]byte, error) {
	log.Infof("%s", b)
	return []byte(""), nil
}

// Error will take the incoming byte slice data and call the internal Tarmac logger converting the data to a string.
func (l *logger) Error(b []byte) ([]byte, error) {
	log.Errorf("%s", b)
	return []byte(""), nil
}

// Debug will take the incoming byte slice data and call the internal Tarmac logger converting the data to a string.
func (l *logger) Debug(b []byte) ([]byte, error) {
	log.Debugf("%s", b)
	return []byte(""), nil
}

// Trace will take the incoming byte slice data and call the internal Tarmac logger converting the data to a string.
func (l *logger) Trace(b []byte) ([]byte, error) {
	log.Tracef("%s", b)
	return []byte(""), nil
}

// Warn will take the incoming byte slice data and call the internal Tarmac logger converting the data to a string.
func (l *logger) Warn(b []byte) ([]byte, error) {
	log.Warnf("%s", b)
	return []byte(""), nil
}
