// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	TRACE LogLevel = 5
	DEBUG LogLevel = 10
	INFO  LogLevel = 20
	WARN  LogLevel = 30
	ERROR LogLevel = 40
)

var (
	trace    *log.Logger
	dbg      *log.Logger
	info     *log.Logger
	warn     *log.Logger
	err      *log.Logger
	minLevel LogLevel = TRACE
)

func Init(file *os.File, level string) error {
	var e error
	minLevel, e = parseLevel(level)
	if e != nil {
		return e
	}
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	if file != nil {
		trace = log.New(file, "TRACE ", flags)
		dbg = log.New(file, "DEBUG ", flags)
		info = log.New(file, "INFO  ", flags)
		warn = log.New(file, "WARN  ", flags)
		err = log.New(file, "ERROR ", flags)
	}
	return nil
}

func parseLevel(value string) (LogLevel, error) {
	switch strings.ToLower(value) {
	case "trace":
		return TRACE, nil
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn", "warning":
		return WARN, nil
	case "err", "error":
		return ERROR, nil
	}
	return 0, fmt.Errorf("%s: invalid log level", value)
}

func Tracef(message string, args ...interface{}) {
	if trace == nil || minLevel > TRACE {
		return
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	_ = trace.Output(2, message)
}

func Debugf(message string, args ...interface{}) {
	if dbg == nil || minLevel > DEBUG {
		return
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	_ = dbg.Output(2, message)
}

func Infof(message string, args ...interface{}) {
	if info == nil || minLevel > INFO {
		return
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	_ = info.Output(2, message)
}

func Warnf(message string, args ...interface{}) {
	if warn == nil || minLevel > WARN {
		return
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	_ = warn.Output(2, message)
}

func Errorf(message string, args ...interface{}) {
	if err == nil || minLevel > ERROR {
		return
	}
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	_ = err.Output(2, message)
}
