package logger

type DummyLogger struct {
}

func (l *DummyLogger) Info(format string, v ...interface{}) {}

func (l *DummyLogger) Error(format string, v ...interface{}) {}
