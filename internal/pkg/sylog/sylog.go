// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

// +build sylog

package sylog

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	apexlog "github.com/apex/log"
)

type messageLevel int

const (
	fatal    messageLevel = iota - 4 // fatal    : -4
	error                            // error    : -3
	warn                             // warn     : -2
	log                              // log      : -1
	_                                // SKIP     : 0
	info                             // info     : 1
	verbose                          // verbose  : 2
	verbose2                         // verbose2 : 3
	verbose3                         // verbose3 : 4
	debug                            // debug    : 5
)

func (l messageLevel) String() string {
	str, ok := messageLabels[l]

	if !ok {
		str = "????"
	}

	return str
}

var messageLabels = map[messageLevel]string{
	fatal:    "FATAL",
	error:    "ERROR",
	warn:     "WARNING",
	log:      "LOG",
	info:     "INFO",
	verbose:  "VERBOSE",
	verbose2: "VERBOSE",
	verbose3: "VERBOSE",
	debug:    "DEBUG",
}

var messageColors = map[messageLevel]string{
	fatal: "\x1b[31m",
	error: "\x1b[31m",
	warn:  "\x1b[33m",
	info:  "\x1b[34m",
}

var colorReset = "\x1b[0m"

var loggerLevel messageLevel

func init() {
	_levelint := int(messageLevel(info))
	_levelstr, ok := os.LookupEnv("SINGULARITY_MESSAGELEVEL")
	if ok {
		_leveli, err := strconv.Atoi(_levelstr)
		if err == nil {
			_levelint = _leveli
		}
	}
	SetLevel(_levelint)
}

func prefix(level messageLevel) string {
	messageColor, ok := messageColors[level]
	if !ok {
		messageColor = "\x1b[0m"
	}

	// This section builds and returns the prefix for levels < debug
	if loggerLevel < debug {
		return fmt.Sprintf("%s%-8s%s ", messageColor, level.String()+":", colorReset)
	}

	pc, _, _, ok := runtime.Caller(3)
	details := runtime.FuncForPC(pc)

	var funcName string
	if ok && details == nil {
		fmt.Printf("Unable to get details of calling function\n")
		funcName = "UNKNOWN CALLING FUNC"
	} else {
		funcNameSplit := strings.Split(details.Name(), ".")
		funcName = funcNameSplit[len(funcNameSplit)-1] + "()"
	}

	uid := os.Geteuid()
	pid := os.Getpid()
	uidStr := fmt.Sprintf("[U=%d,P=%d]", uid, pid)

	return fmt.Sprintf("%s%-8s%s%-19s%-30s", messageColor, level, colorReset, uidStr, funcName)
}

func writef(w io.Writer, level messageLevel, format string, a ...interface{}) {
	if loggerLevel < level {
		return
	}

	message := fmt.Sprintf(format, a...)
	message = strings.TrimSuffix(message, "\n")

	fmt.Fprintf(w, "%s%s\n", prefix(level), message)
}

// Fatalf is equivalent to a call to Errorf followed by os.Exit(255). Code that
// may be imported by other projects should NOT use Fatalf.
func Fatalf(format string, a ...interface{}) {
	writef(os.Stderr, fatal, format, a...)
	os.Exit(255)
}

// Errorf writes an ERROR level message to the log but does not exit. This
// should be called when an error is being returned to the calling thread
func Errorf(format string, a ...interface{}) {
	writef(os.Stderr, error, format, a...)
}

// Warningf writes a WARNING level message to the log.
func Warningf(format string, a ...interface{}) {
	writef(os.Stderr, warn, format, a...)
}

// Infof writes an INFO level message to the log. By default, INFO level messages
// will always be output (unless running in silent)
func Infof(format string, a ...interface{}) {
	writef(os.Stderr, info, format, a...)
}

// Verbosef writes a VERBOSE level message to the log. This should probably be
// deprecated since the granularity is often too fine to be useful.
func Verbosef(format string, a ...interface{}) {
	writef(os.Stderr, verbose, format, a...)
}

// Debugf writes a DEBUG level message to the log.
func Debugf(format string, a ...interface{}) {
	writef(os.Stderr, debug, format, a...)
}

// SetLevel explicitly sets the loggerLevel
func SetLevel(l int) {
	loggerLevel = messageLevel(l)
	// set the apex log level, for umoci
	if loggerLevel <= log {
		// quiet or silent options
		apexlog.SetLevel(apexlog.ErrorLevel)
	} else if loggerLevel <= info {
		// verbose option or default
		apexlog.SetLevel(apexlog.InfoLevel)
	} else {
		// debug option
		apexlog.SetLevel(apexlog.DebugLevel)
	}
}

// DisableColor for the logger
func DisableColor() {
	messageColors = map[messageLevel]string{
		fatal: "",
		error: "",
		warn:  "",
		info:  "",
	}
	colorReset = ""
}

// GetLevel returns the current log level as integer
func GetLevel() int {
	return int(loggerLevel)
}

// GetEnvVar returns a formatted environment variable string which
// can later be interpreted by init() in a child proc
func GetEnvVar() string {
	return fmt.Sprintf("SINGULARITY_MESSAGELEVEL=%d", loggerLevel)
}

// Writer returns an io.Writer to pass to an external packages logging utility.
// i.e when --quiet option is set, this function returns ioutil.Discard writer to ignore output
func Writer() io.Writer {
	if loggerLevel <= -1 {
		return ioutil.Discard
	}

	return os.Stderr
}
