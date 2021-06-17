package app

type logger struct{}

func (l *logger) Info(b []byte) ([]byte, error) {
	log.Infof("%s", b)
	return []byte(""), nil
}

func (l *logger) Error(b []byte) ([]byte, error) {
	log.Errorf("%s", b)
	return []byte(""), nil
}

func (l *logger) Debug(b []byte) ([]byte, error) {
	log.Debugf("%s", b)
	return []byte(""), nil
}

func (l *logger) Trace(b []byte) ([]byte, error) {
	log.Tracef("%s", b)
	return []byte(""), nil
}

func (l *logger) Warn(b []byte) ([]byte, error) {
	log.Warnf("%s", b)
	return []byte(""), nil
}
