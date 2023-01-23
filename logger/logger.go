// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package logger

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	TRACE LogLevel = 5
	DEBUG LogLevel = 10
	INFO  LogLevel = 20
	WARN  LogLevel = 30
	ERROR LogLevel = 40
)

var nameToLevel = map[string]LogLevel{
	"trace": TRACE,
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"error": ERROR,
}

var (
	trace    *log.Logger
	dbg      *log.Logger
	info     *log.Logger
	warn     *log.Logger
	err      *log.Logger
	minLevel LogLevel = TRACE
)

func Init(file *os.File, level string) error {
	var found bool
	minLevel, found = nameToLevel[level]
	if !found {
		return fmt.Errorf("%s: invalid log level", level)
	}
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	if file != nil {
		trace = log.New(file, "trace ", flags)
		dbg = log.New(file, "debug ", flags)
		info = log.New(file, "info  ", flags)
		warn = log.New(file, "warn  ", flags)
		err = log.New(file, "error ", flags)
	}
	return nil
}

func LevelNames() []string {
	var names []string
	for n := range nameToLevel {
		names = append(names, n)
	}
	return names
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
