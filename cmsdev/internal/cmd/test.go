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
 * test.go
 *
 * test commons file
 *
 */
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/bos"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/cfs"
	con "stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/conman"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/ims"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/ipxe_tftp"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/test/vcs"
	"strconv"
	"strings"
	"time"
)

// Test timeouts (in seconds)
// (test will not retry after this amount of time)
var TestTimeouts = map[string]int64{
	"bos":    300,
	"cfs":    300,
	"conman": 300,
	"ipxe":   300,
	"tftp":   300,
	"vcs":    300,
	"gitea":  300,
}

// Default test timeout (in seconds)
// This value is used if service is not listed in previous map
const DefaultTestTimeout int64 = 120

// Run the specified test
func RunTest(service string) bool {
	switch service {
	case "bos":
		return bos.IsBOSRunning()
	case "cfs":
		return cfs.IsCFSRunning()
	case "conman":
		return con.IsConmanRunning()
	case "ims":
		return ims.IsIMSRunning()
	case "ipxe", "tftp":
		return ipxe_tftp.AreTheyRunning()
	case "vcs", "gitea":
		return vcs.IsVCSRunning()
	}
	common.Usagef("Programming logic error: this line should never be reached. Invalid service (%s), but it should already have been validated!", service)
	return false
}

func GetTimeout(service string) (timeout int64) {
	timeout, ok := TestTimeouts[service]
	if !ok {
		timeout = DefaultTestTimeout
	}
	return
}

// The first sleep time is 5 seconds, then it is increased by 5 seconds for each
// attempt, maxing out at 1/6 of the maximum retry time. The sleep time is truncated
// as the maximum retry time limit is approached.
func GetSleepDuration(attempt int, stopTime, timeout int64) time.Duration {
	var sleepSeconds int64
	timeLeft := stopTime - time.Now().Unix()
	if timeLeft <= 1 {
		sleepSeconds = 1
	} else {
		maxSleep := timeout / 6
		sleepSeconds = int64(attempt) * 5
		if sleepSeconds > maxSleep {
			sleepSeconds = maxSleep
		}
		if sleepSeconds >= timeLeft {
			sleepSeconds = timeLeft - 1
		} else if (sleepSeconds * 2) > timeLeft {
			sleepSeconds = (timeLeft / 2) + 1
		}
	}
	return time.Duration(sleepSeconds) * time.Second
}

func DoTestWithRetry(service string) bool {
	testPassed, finalTry := false, false
	timeout := GetTimeout(service)
	common.Infof("Test retry timeout for this service test is %d seconds", timeout)
	stopTime := timeout + time.Now().Unix()
	for n := 1; ; n++ {
		if (time.Now().Unix() + 1) >= stopTime {
			common.Infof("Final attempt")
			finalTry = true
		} else {
			common.Infof("Attempt #%d", n)
		}
		common.SetRunSubTag(strconv.Itoa(n))
		testPassed = RunTest(service)
		common.UnsetRunSubTag()
		if testPassed {
			return true
		} else if finalTry {
			return false
		} else if time.Now().Unix() >= (stopTime + 30) {
			common.Infof("Not retrying because stop time has already been exceeded by at least 30 seconds")
			return false
		}
		sleepDuration := GetSleepDuration(n, stopTime, timeout)
		common.Infof("Attempt failed; waiting %v before retrying", sleepDuration)
		time.Sleep(sleepDuration)
	}
	common.Errorf("PROGRAMMING LOGIC ERROR: DoTestWithRetry: This line should never be reached")
	return false
}

func DoTest(service string, retry bool) bool {
	if retry {
		return DoTestWithRetry(service)
	} else {
		return RunTest(service)
	}
}

func GetTestNamesList(excludeAliases bool) []string {
	var s string
	services := make([]string, 0)

	if !excludeAliases {
		// Start with the "all" alias
		services = append(services, "all")
	}
	// Append the remaining services (omitting aliases if specified)
	for _, s = range common.CMSServices {
		if excludeAliases && common.StringInArray(s, common.CMSServicesDuplicates) {
			// Skip this one, as it is an alias of another one
			continue
		}
		services = append(services, s)
	}
	return services
}

func GetTestNamesString(excludeAliases bool) string {
	return strings.Join(GetTestNamesList(excludeAliases), " ")
}

var longHelpText = fmt.Sprintf(`test command runs service tests.

Valid service tests: %s

Example Commands:

cmsdev test -l
  # list all valid services to test
cmsdev test -l --exclude-aliases
  # list all valid services to test, excluding aliases
cmsdev test conman
  # runs conman tests
cmsdev test tftp --no-log -q
  # runs tftp tests in quiet mode with logging disabled
cmsdev test cfs -r --verbose
  # runs cfs tests with verbosity and retry on failure`, GetTestNamesString(false))

// testCmd command functions
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "cms test services command",
	Long:  longHelpText,
	Run: func(cmd *cobra.Command, args []string) {
		noLogs, _ := cmd.Flags().GetBool("no-log")
		logsDir, _ := cmd.Flags().GetString("log-dir")
		retry, _ := cmd.Flags().GetBool("retry")
		quiet, _ := cmd.Flags().GetBool("quiet")
		verbose, _ := cmd.Flags().GetBool("verbose")
		listTests, _ := cmd.Flags().GetBool("list")
		excludeAliases, _ := cmd.Flags().GetBool("exclude-aliases")

		if quiet && verbose {
			common.Usagef("--quiet and --verbose are mutually exclusive")
		}

		if listTests {
			// --list was passed
			if noLogs || logsDir != "" || retry || quiet || verbose {
				common.Usagef("--no-log, --log-dir, --retry, --quiet, and --verbose are not valid with --list")
			} else if len(args) > 0 {
				common.Usagef("Invalid arguments specified with --list: %s", strings.Join(args, " "))
			}
			common.Infof(GetTestNamesString(excludeAliases))
			return
		}

		// --list was not passed
		if excludeAliases {
			common.Usagef("--exclude-aliases is only valid with --list")
		} else if len(args) < 1 {
			common.Usagef("Argument required, provide one or more of the following: %s\n", GetTestNamesString(false))
		}

		var s string
		services := make([]string, 0)
		allServices := false
		for _, a := range args {
			s = strings.TrimSpace(a)
			// do some command line args checking
			if s == "all" {
				allServices = true
			} else if !common.StringInArray(s, common.CMSServices) {
				common.Usagef("Invalid test: '%s'. Supported tests are: %s", s, GetTestNamesString(false))
			} else if !allServices {
				services = append(services, s)
			}
		}
		if allServices {
			services = GetTestNamesList(true)
		} else if len(services) == 0 {
			common.Usagef("Argument required, provide one or more of the following: %s\n", GetTestNamesString(false))
		}

		logs := !noLogs

		// create log file if logs, ignore logsDir if !logs
		// cmsdevVersion is found in version.go
		common.CreateLogFile(logsDir, cmsdevVersion, logs, retry, quiet, verbose)

		// Initialize variables related to saving CT test artifacts
		common.InitArtifacts()

		if len(services) == 1 {
			common.SetTestService(services[0])
			ok := DoTest(services[0], retry)
			common.UnsetTestService()
			// Compress test artifacts, if any
			common.CompressArtifacts()
			if ok {
				common.Success()
			} else {
				common.Failure()
			}
		}

		// Testing multiple services
		var passed, failed []string
		for _, s = range services {
			common.SetTestService(s)
			if DoTest(s, retry) {
				passed = append(passed, s)
			} else {
				failed = append(failed, s)
			}
			common.UnsetTestService()
		}
		// Compress test artifacts, if any
		common.CompressArtifacts()
		if len(failed) > 0 {
			common.Failuref("%d service tests FAILED (%s), %d passed (%s)", len(failed), strings.Join(failed, ", "),
				len(passed), strings.Join(passed, ", "))
		} else {
			common.Successf("All %d service tests passed: %s", len(passed), strings.Join(passed, ", "))
		}

		common.Errorf("PROGRAMMING LOGIC ERROR: test.go: This line should never be reached")
		return
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringP("log-dir", "", "", "specify log directory")
	testCmd.Flags().BoolP("no-log", "", false, "do not log to a file")
	testCmd.Flags().BoolP("retry", "r", false, "retry on failure")
	testCmd.Flags().BoolP("quiet", "q", false, "quiet mode")
	testCmd.Flags().BoolP("verbose", "v", false, "verbose mode")
	testCmd.Flags().BoolP("list", "l", false, "list valid service tests")
	testCmd.Flags().BoolP("exclude-aliases", "", false, "exclude aliases from list of valid service tests")
}
