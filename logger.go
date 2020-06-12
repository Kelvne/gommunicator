package gommunicator

import (
	"fmt"
	"log"
	"os"
	"time"
)

const reset = "\033[0m"
const red = "\033[31m"
const green = "\033[32m"
const yellow = "\033[33m"
const blue = "\033[34m"
const purple = "\033[35m"
const whiteRed = "\033[97;31m"

func formatLog(info bool, values ...interface{}) []interface{} {
	now := time.Now()
	nowDate := now.Format("06-01-02")
	nowTime := now.Format("15:04:05")
	header := fmt.Sprintf("%s[   INFO   ]%s", green, reset)

	if !info {
		header = fmt.Sprintf("%s[ %s! %sERROR %s! %s]%s", red, yellow, red, yellow, red, reset)
	}

	ts := fmt.Sprintf("[%s%s%s %s%s%s] %s", purple, nowDate, reset, yellow, nowTime, reset, header)

	messages := []interface{}{}

	qt := len(values) - 1
	for i, value := range values {
		extra := "\r\n"
		if i == qt {
			extra = ""
		}
		messages = append(messages, fmt.Sprintf("%s %v %s", ts, value, extra))
	}

	return messages
}

// Logger simple logger
type Logger struct {
	*log.Logger
}

func (logger *Logger) log(info bool, values ...interface{}) {
	logger.Println(formatLog(info, values...)...)
}

var logger *Logger

func getLogger() *Logger {
	if logger == nil {
		newLogger := log.New(os.Stdout, "\r\n", 0)

		logger = &Logger{newLogger}
	}

	return logger
}

func logInfo(values ...interface{}) {
	getLogger().log(true, values...)
}

func logErr(values ...interface{}) {
	getLogger().log(false, values...)
}

func formatWithID(id, key, message string) string {
	return fmt.Sprintf("%s [%s%s%s: %s%s%s]", message, yellow, key, reset, whiteRed, id, reset)
}
