package mock

type DummyLogger struct {
}

func (l *DummyLogger) Info(format string, v ...interface{}) {}

func (l *DummyLogger) Warn(format string, v ...interface{}) {}

func (l *DummyLogger) Fatal(format string, v ...interface{}) {}

func (l *DummyLogger) Error(format string, v ...interface{}) {}
