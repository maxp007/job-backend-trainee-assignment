package logger

import (
	"fmt"
	"io"
	"log"
	"sync"
)

type ILogger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type FileLogger struct {
	logLevel int64
	mu       sync.Mutex
	log      *log.Logger
}

const (
	L_INFO  = 2
	L_ERROR = 4
)

func NewLogger(out io.Writer, prefix string, logLevel int64) *FileLogger {
	stdLogger := log.New(out, prefix, log.Lshortfile|log.LstdFlags|log.Lmicroseconds)

	return &FileLogger{
		logLevel: logLevel,
		mu:       sync.Mutex{},
		log:      stdLogger,
	}
}

func (l *FileLogger) Info(format string, v ...interface{}) {
	var lvl int64
	var logger *log.Logger

	l.mu.Lock()
	lvl = l.logLevel
	logger = l.log
	l.mu.Unlock()

	if lvl <= L_INFO {
		err := logger.Output(2, " INFO \t"+fmt.Sprintf(format, v...)+"\n")
		if err != nil {
			log.Printf("Failed to write logs with Output method, err %v", err)
		}
	}
}

func (l *FileLogger) Error(format string, v ...interface{}) {
	var lvl int64
	var logger *log.Logger

	l.mu.Lock()
	lvl = l.logLevel
	logger = l.log
	l.mu.Unlock()

	if lvl <= L_ERROR {
		err := logger.Output(2, " ERROR \t"+fmt.Sprintf(format, v...)+"\n")
		if err != nil {
			log.Printf("Failed to write logs with Output method, err %v", err)
		}
	}
}
