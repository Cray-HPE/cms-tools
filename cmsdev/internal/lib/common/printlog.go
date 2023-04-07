//
//  MIT License
//
//  (C) Copyright 2019-2023 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//
/*
 * printlog.go
 *
 * Functions relating to output and logging
 *
 */

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	c "github.com/fatih/color"
	"github.com/sirupsen/logrus"
	resty "gopkg.in/resty.v1"
	"os"
	"runtime"
	"strings"
	"time"
)

const DEFAULT_LOG_FILE_DIR string = "/opt/cray/tests/install/logs/cmsdev"

// Relative path to this source file within its repo
const RELATIVE_PATH_TO_THIS_FILE = "cmsdev/internal/lib/common/printlog.go"

var srcPrefixSubstring string
var logFileDir string

// log file handle
var logFile *logrus.Logger
var testLog *logrus.Entry
var printInfo, printWarn, printError, printResults, printVerbose bool

func printlogInit() {
	_, fn, _, _ := runtime.Caller(0) // Find our own filename
	// We want to specify the relative paths to our source files when logging
	// So given that we know the relative path to ourselves (specified by
	// the RELATIVE_PATH_TO_THIS_FILE constant), we will determine the path
	// prefix in our build environment. We can then strip this prefix away
	// from all source file names when logging
	substringIndex := strings.Index(fn, RELATIVE_PATH_TO_THIS_FILE)
	if substringIndex == -1 {
		// This is unexpected, but not much we can do about it. In this
		// case, we will set the prefix to the empty string, which will
		// result in all source files using full paths when logging
		srcPrefixSubstring = ""
		return
	}
	srcPrefixSubstring = fn[:substringIndex]
}

// Get the relative path of the source file within the repo
func srcString(callerFileName string, callerLineNum int) string {
	fname := callerFileName
	if len(srcPrefixSubstring) > 0 {
		substringIndex := strings.Index(callerFileName, srcPrefixSubstring)
		if substringIndex == 0 {
			fname = callerFileName[len(srcPrefixSubstring):]
		}
	}
	return fmt.Sprintf("%s:%d", fname, callerLineNum)
}

func logFields(callerFileName string, callerLineNum int) logrus.Fields {
	return logrus.Fields{"src": srcString(callerFileName, callerLineNum), "run": runTag, "service": testService}
}

// Wrappers to Debugf, Infof,  Warnf, and Errorf test log functions
func TestLogDebugf(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	testLog.WithFields(logFields(callerFileName, callerLineNum)).Debugf(format, a...)
}

func TestLogInfof(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	testLog.WithFields(logFields(callerFileName, callerLineNum)).Infof(format, a...)
}

func TestLogWarnf(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	testLog.WithFields(logFields(callerFileName, callerLineNum)).Warnf(format, a...)
}

func TestLogErrorf(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	testLog.WithFields(logFields(callerFileName, callerLineNum)).Errorf(format, a...)
}

// Wrapper for default print function, in case we want to do anything in the future to
// control whether or where things are printed
func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// print and/or log messages to the appropriate level
func InfofWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if printInfo {
		fmt.Printf(format+"\n", a...)
	}
	if testLog != nil {
		TestLogInfof(callerFileName, callerLineNum, format, a...)
	}
}

func Infof(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	InfofWithCallerInfo(fn, line, format, a...)
}

func DebugfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if printVerbose {
		fmt.Printf(format+"\n", a...)
	}
	if testLog != nil {
		TestLogDebugf(callerFileName, callerLineNum, format, a...)
	}
}

func Debugf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	DebugfWithCallerInfo(fn, line, format, a...)
}

func InfoOverridefWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	if testLog != nil {
		TestLogInfof(callerFileName, callerLineNum, format, a...)
	}
}

func InfoOverridef(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	InfoOverridefWithCallerInfo(fn, line, format, a...)
}

func WarnfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if printWarn {
		if runTag != "" {
			fmt.Printf("WARNING (run tag "+runTag+"): "+format+"\n", a...)
		} else {
			fmt.Printf("WARNING: "+format+"\n", a...)
		}
	}
	if testLog != nil {
		TestLogWarnf(callerFileName, callerLineNum, format, a...)
	}
}

func Warnf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	WarnfWithCallerInfo(fn, line, format, a...)
}

func ErrorfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if printError {
		if runTag != "" {
			fmt.Printf("ERROR (run tag "+runTag+"): "+format+"\n", a...)
		} else {
			fmt.Printf("ERROR: "+format+"\n", a...)
		}
	}
	if testLog != nil {
		TestLogErrorf(callerFileName, callerLineNum, format, a...)
	}
}

func Errorf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ErrorfWithCallerInfo(fn, line, format, a...)
}

func Error(err error) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ErrorfWithCallerInfo(fn, line, err.Error())
}

// If format is not blank, call Infof with format + a
// In verbose mode, also print green OK
func VerboseOkayfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if len(format) > 0 {
		InfofWithCallerInfo(callerFileName, callerLineNum, format, a...)
	}
	if printVerbose {
		c.HiGreen("OK")
	}
}

func VerboseOkayf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	VerboseOkayfWithCallerInfo(fn, line, format, a...)
}

func VerboseOkay() {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	VerboseOkayfWithCallerInfo(fn, line, "")
}

// If format is not blank, call Errorf with format + a
// In verbose mode, also print red Failed
func VerboseFailedfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if len(format) > 0 {
		ErrorfWithCallerInfo(callerFileName, callerLineNum, format, a...)
	}
	if printVerbose {
		c.Red("Failed")
	}
}

func VerboseFailedf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	VerboseFailedfWithCallerInfo(fn, line, format, a...)
}

func VerboseFailed() {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	VerboseFailedfWithCallerInfo(fn, line, "")
}

func Verbosef(format string, a ...interface{}) {
	if printVerbose {
		fmt.Printf(format+"\n", a...)
	}
}

// Print a dividing line to stdout if in verbose mode
func VerbosePrintDivider() {
	Verbosef("---\n")
}

// pretty print resty json responses
func PrettyPrintJSON(resp *resty.Response) {
	if !printVerbose {
		return
	}
	var prettyJSON bytes.Buffer

	err := json.Indent(&prettyJSON, resp.Body(), "", "   ")
	if err != nil {
		fmt.Printf("%v\n", resp)
	} else {
		fmt.Printf("%s\n", strings.TrimSpace(string(prettyJSON.Bytes())))
	}
}

func ResultsfWithCallerInfo(callerFileName string, callerLineNum int, format string, a ...interface{}) {
	if printResults {
		if runTag != "" {
			fmt.Printf("(run tag "+runTag+"): "+format+"\n", a...)
		} else {
			fmt.Printf(format+"\n", a...)
		}
	}
	if testLog != nil {
		TestLogInfof(callerFileName, callerLineNum, format, a...)
	}
}

func Resultsf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ResultsfWithCallerInfo(fn, line, format, a...)
}

// print/log result and exit with specified code
func ExitfWithCallerInfo(callerFileName string, callerLineNum, rc int, format string, a ...interface{}) {
	var res string
	for len(runTags) > 1 {
		UnsetRunSubTag()
	}
	if len(runTags) == 1 {
		fmt.Printf("Ended run, tag: %s (duration: %v)\n", runTag, time.Since(runStartTimes[len(runStartTimes)-1]))
		runStartTimes = runStartTimes[:len(runStartTimes)-1]
		runTags = runTags[:len(runTags)-1]
		runTag = ""
	}

	switch rc {
	case 0:
		res = "SUCCESS"
	case 1:
		res = "FAILURE"
	case 2:
		res = "USAGE ERROR"
	default:
		res = "UNKNOWN ERROR"
	}
	if len(format) > 0 {
		ResultsfWithCallerInfo(callerFileName, callerLineNum, res+": "+format, a...)
	} else {
		ResultsfWithCallerInfo(callerFileName, callerLineNum, res)
	}
	if logFile != nil {
		logFile.Exit(rc)
	} else {
		os.Exit(rc)
	}
}

func Exitf(rc int, format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, rc, format, a...)
}

func Usagef(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, 2, format, a...)
}

func Successf(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, 0, format, a...)
}

func Success() {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, 0, "")
}

func Failuref(format string, a ...interface{}) {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, 1, format, a...)
}

func Failure() {
	_, fn, line, _ := runtime.Caller(1) // the file and line number of the caller
	ExitfWithCallerInfo(fn, line, 1, "")
}

// create log file and directory provided by path if one does not exist
// if no path is provided, use DEFAULT_LOG_FILE_DIR
func CreateLogFile(path, version string, logs, retry, quiet, verbose bool) {
	var err error

	if verbose {
		printVerbose = true
	} else if quiet {
		printInfo, printWarn = false, false
	}
	if !logs {
		return
	} else if len(path) == 0 {
		path = DEFAULT_LOG_FILE_DIR
	}
	err, _ = CreateDirectoryIfNeeded(path)
	if err != nil {
		fmt.Printf("Error with log directory: %s\n", path)
		panic(err)
	}
	logfile := path + "/cmsdev.log"
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logFile = logrus.New()
	if err != nil {
		panic(err)
	}
	// Log everything debug and above
	logFile.SetLevel(logrus.DebugLevel)

	// We want nanosecond precision in log file entries
	logFile.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	logFile.SetOutput(f)
	args := make([]string, 0, 5)
	if retry {
		args = append(args, "retry")
	}
	if quiet {
		args = append(args, "quiet")
	}
	if verbose {
		args = append(args, "verbose")
	}
	runTag = strings.Join(runTags, "-")
	testLog = logFile.WithFields(logrus.Fields{"version": version, "args": strings.Join(args, ",")})
	logFileDir = path
	Infof("cmsdev starting")
	fmt.Printf("Starting main run, version: %s, tag: %s\n", version, runTag)
}
