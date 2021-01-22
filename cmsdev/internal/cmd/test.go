/*
 * test.go
 *
 * test commons file
 *
 * Copyright 2019-2020 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"github.com/spf13/cobra"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/bos"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/cfs"
	con "stash.us.cray.com/cms-tools/cmsdev/internal/test/conman"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/crus"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/ims"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/ipxe_tftp"
	"stash.us.cray.com/cms-tools/cmsdev/internal/test/vcs"
	"strconv"
	"strings"
	"time"
)

// stage4 CT test timeouts (in seconds)
// (test will not retry after this amount of time)
//
// Currently the following tests only consist of
// checks of k8s pods and their statuses:
// bos, cfs, conman, ipxe, tftp, vcs
// Because these checks are the ones most often
// to succeed after retry (as the services are
// starting up during stage4), we give those
// components longer retry times.
//
// Note that once CASMCMS-5657 is implemented,
// this will no longer be necessary.
var Stage4CTTestTimeouts = map[string]int64{
	"bos":    300,
	"cfs":    300,
	"conman": 300,
	"crus":   300,
	"ipxe":   300,
	"tftp":   300,
	"vcs":    300,
	"gitea":  300,
}

// Default stage4 CT test timeout (in seconds)
// This value is used if service is not listed in previous map
const DefaultStage4CTTestTimeout int64 = 60

// Run the specified CT test
func RunCTTest(service, stage string, local, smoke, ct bool) bool {
	switch service {
	case "bos":
		return bos.IsBOSRunning(local, smoke, ct, stage)
	case "cfs":
		return cfs.IsCFSRunning(local, smoke, ct, stage)
	case "conman":
		return con.IsConmanRunning(local, smoke, ct, stage)
	case "crus":
		return crus.IsCRUSRunning(local, smoke, ct, stage)
	case "ims":
		return ims.IsIMSRunning(local, smoke, ct, stage)
	case "ipxe", "tftp":
		return ipxe_tftp.AreTheyRunning(local, smoke, ct, stage)
	case "vcs", "gitea":
		return vcs.IsVCSRunning(local, smoke, ct, stage)
	}
	common.Usagef("Programming logic error: this line should never be reached. Invalid service (%s), but it should already have been validated!", service)
	return false
}

func GetTimeout(service string) (timeout int64) {
	timeout, ok := Stage4CTTestTimeouts[service]
	if !ok {
		timeout = DefaultStage4CTTestTimeout
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

// testCmd command functions
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "cms test services command",
	Long: `test command runs cms ct, smoke, and api services tests.
Example Commands:

cmsdev test conman --ct
  # runs conman ct tests at default level 4
cmsdev test tftp --ct --logs
  # runs tftp ct tests with logging enabled
cmsdev test ims --ct --crayctlStage=2
  # runs ims ct tests at crayctl stage level 2
cmsdev test bos --api
  # runs entire bos api test suite
cmsdev test cfs --smoke --verbose
  # runs cfs smoke tests with verbosity`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: add custom flag validation
		if len(args) < 1 {
			common.Usagef("argument required, provide cms service name: %s\n", strings.Join(common.CMSServices, " "))
		}
		service := strings.TrimSpace(args[0])

		// TODO: pass these flags as args in a more eloquent way
		api, _ := cmd.Flags().GetBool("api")
		ct, _ := cmd.Flags().GetBool("ct")
		smoke, _ := cmd.Flags().GetBool("smoke")
		stage, _ := cmd.Flags().GetString("crayctl-stage")
		local, _ := cmd.Flags().GetBool("local")
		logs, _ := cmd.Flags().GetBool("logs")
		logsDir, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// create log file if logs, ignore logsDir if !logs
		common.CreateLogFile(logsDir, stage, service, api, ct, local, logs, smoke, verbose)

		// do some command line args checking
		if validService := common.StringInArray(service, common.CMSServices); !validService {
			common.Usagef("supported cms services are %s", strings.Join(common.CMSServices, " "))
		} else if !(api || ct || smoke) {
			common.Usagef("argument required, available cms tests types are: %s",
				strings.Join(common.CMSServicesTestTypes, " "))
		} else if ct {
			i, _ := strconv.Atoi(stage)
			if i < 0 || i > 5 {
				common.Usagef("valid stages are crayctl-stage=1..5")
			}
		}

		if ct {
			// Initialize variables related to saving CT test artifacts
			common.InitArtifacts(stage, service)

			if stage != "4" {
				if RunCTTest(service, stage, local, smoke, ct) {
					common.Success()
				} else {
					common.Failure()
				}
			} else {
				// CASMCMS-4429, CASMCMS-4574: Add automatic retries for CT tests in stage 4
				// We will sleep between retries until the test passes or we exceed the maximum
				// retry time for that test.
				testPassed, finalTry := false, false
				timeout := GetTimeout(service)
				common.Infof("Test retry timeout is %d seconds", timeout)
				stopTime := timeout + time.Now().Unix()
				for n := 1; ; n++ {
					if (time.Now().Unix() + 1) >= stopTime {
						common.Infof("Final attempt")
						finalTry = true
					} else {
						common.Infof("Attempt #%d", n)
					}
					common.SetRunSubTag(strconv.Itoa(n))
					testPassed = RunCTTest(service, stage, local, smoke, ct)
					common.UnsetRunSubTag()
					if testPassed {
						common.Success()
					} else if finalTry {
						common.Failure()
					} else if time.Now().Unix() >= (stopTime + 30) {
						common.Failuref("Not retrying because stop time has already been exceeded by at least 30 seconds")
					}
					sleepDuration := GetSleepDuration(n, stopTime, timeout)
					common.Infof("Attempt failed; waiting %v before retrying", sleepDuration)
					time.Sleep(sleepDuration)
				}
			}
		}

		switch service {
		case "bos":
			if ct || smoke {
				_ = bos.IsBOSRunning(local, smoke, ct, stage)
			} else { // api
				if len(args) == 2 {
					if !common.StringInArray(args[1], bos.APITestTypes) {
						common.Usagef("argument required, current supported api tests are: %s\n%s\n",
							strings.Join(bos.APITestTypes, " "),
							"remove arguments to run entire tests suite")
					}
					bos.RunAPITests(local, args[1])
				} else {
					bos.RunAPITests(local)
				}
			}
		case "cfs":
			if ct || smoke {
				_ = cfs.IsCFSRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		case "conman":
			if ct || smoke {
				_ = con.IsConmanRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		case "crus":
			if ct || smoke {
				_ = crus.IsCRUSRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		case "ims":
			if ct || smoke {
				_ = ims.IsIMSRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		case "ipxe", "tftp":
			if ct || smoke {
				_ = ipxe_tftp.AreTheyRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		case "vcs", "gitea":
			if ct || smoke {
				_ = vcs.IsVCSRunning(local, smoke, ct, stage)
			} else { // api
				common.Usagef("no api tests currently found for %s\n", service)
			}
		}
		common.Infof("cmsdev completed")
		return
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().Bool("api", false, "api tests")
	testCmd.Flags().Bool("ct", false, "ct tests")
	testCmd.Flags().Bool("smoke", false, "run smoke tests")
	// TODO: ensure that these flags are persistent
	testCmd.Flags().BoolP("local", "l", false, "run tests locally, not implemented")
	testCmd.Flags().BoolP("logs", "", false, "enable test script logging")
	testCmd.Flags().StringP("output", "o", "", "output logging to a file, requires --logs")
	testCmd.Flags().BoolP("verbose", "v", false, "verbose mode")
	testCmd.Flags().String("crayctl-stage", "4", "ct test stage")
}
