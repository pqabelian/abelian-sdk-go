package core

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

type Logger struct {
	*log.Logger
}

func NewLogger(name string) *Logger {
	return &Logger{
		Logger: log.New(os.Stderr, fmt.Sprintf("[%s] ", name), log.LstdFlags),
	}
}

func loggerEnabled() bool {
	enabled := strings.ToLower(os.Getenv("ABELSDK_DEBUG"))
	return enabled == "true" || enabled == "1" || enabled == "on" || enabled == "yes"
}

func (logger *Logger) debug(format string, v ...interface{}) {
	// Check if logger is enabled.
	if !loggerEnabled() {
		return
	}

	// Get caller file name and function name.
	pc, file, line, _ := runtime.Caller(1)
	fileName := file[strings.LastIndex(file, "/")+1:]
	funcName := runtime.FuncForPC(pc).Name()
	funcName = funcName[strings.LastIndex(funcName, ".")+1:]

	// Make caller info.
	callerInfo := fmt.Sprintf("%s:%d:%s ", fileName, line, funcName)

	// Print log.
	logger.Printf(callerInfo+format, v...)
}

var LOG = NewLogger("abelsdk")
