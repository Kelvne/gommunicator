package gommunicator

import (
	"fmt"
	"log"
	"os"
)

const reset = "\033[0m"
const red = "\033[31m"
const green = "\033[32m"
const yellow = "\033[33m"
const blue = "\033[34m"
const purple = "\033[35m"
const whiteRed = "\033[97;31m"

func formatLog(values ...interface{}, info bool) []interface{} {
	now := time.Now()
	nowDate := now.Format("06-01-02")
	nowTime := now.Format("15:04:05")
	titleColor := blue
	header := fmt.Sprintf("%s[   INFO   ]%s", green, reset)

	if !info {
		header = fmt.Sprintf("%s[ %s! %sERROR %s! %s]%s", red, yellow, red, yellow, red, reset)
	}
	
	if !info {
		titleColor = red
	}

	ts := fmt.Sprintf("[%s%s%s %s%s%s] %s", purple, nowDate, reset, yellow, nowTime, reset, header)

	messages := []interface{}{ts}

	qt := len(values) - 1
	for i, value := values {
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

func (logger *Logger) log(values ...interface{}, info bool) {
	logger.Println(formatLog(values..., info)...)
}

var logger *Logger

func getLogger() *Logger {
	if logger == nil {
		newLogger, err := log.New(os.Stdout, "\r\n", 0)
		defer newLogger.Sync()
		if err != nil {
			return nil
		}

		logger = &Logger{newLogger}
	}

	return logger
}

func logInfo(values ...interface{}) {
	getLogger().log(values, true)
}

func logErr(values ...interface{}) {
	getLogger().log(values, false)
}
