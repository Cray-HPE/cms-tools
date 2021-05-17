/*
 * common.go
 *
 * Library of general utility functions
 *
 * Copyright 2019-2021 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * (MIT License)
 */

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	c "github.com/fatih/color"
	"github.com/go-resty/resty"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const BASEHOST = "api-gw-service-nmn.local"
const BASEURL = "https://" + BASEHOST
const LOCALHOST = "http://localhost:5000"
const NAMESPACE string = "services"
const DEFAULT_LOG_FILE_DIR string = "/opt/cray/tests"

// struct to hold endpoint METHOD operation details
type endpointMethod struct {
	parameters string // TODO: change to interface
	responses  []int
	summary    string
}

// data structure to endpoints, URL, descriptions
type Endpoint struct {
	Methods map[string]*endpointMethod // endpoint METHOD data
	Url     string                     // url string appended to base to reach endpoint
	Version string                     `default:"v1"` // endpoint version
}

// Restful() parameters
type Params struct {
	Token        string
	JsonStrArray []byte
	JsonStr      string
}

// pod name prefixes
var PodServiceNamePrefixes = map[string]string{
	"bos":              "cray-bos",
	"bosPvc":           "cray-bos-etcd",
	"cfs":              "cray-cfs",
	"cfs-api":          "cray-cfs-api",
	"cfs-operator":     "cray-cfs-operator",
	"cfsServices":      "^(cray-cfs-operator|cray-cfs-api)",
	"console":          "cray-console",
	"console-data":     "cray-console-data",
	"console-node":     "cray-console-node",
	"console-operator": "cray-console-operator",
	"crus":             "cray-crus",
	"ims":              "cray-ims",
	"ipxe":             "cray-ipxe",
	"tftp":             "cray-tftp",
	"vcs":              "gitea-vcs",
}

// list of CMS services
var CMSServices = []string{
	"bos",
	"cfs",
	"conman",
	"crus",
	"gitea",
	"ims",
	"ipxe",
	"tftp",
	"vcs",
}

// log file handle
var Log *logrus.Logger
var TestLog *logrus.Entry
var PrintInfo, PrintWarn, PrintError, PrintResults, PrintVerbose bool
var RunTag, RunBaseTag string
var artifactDirectory, artifactFilePrefix string

// Set and unset the run sub-tag
func SetRunSubTag(tag string) {
	RunTag = fmt.Sprintf("%s-%s", RunBaseTag, tag)
	fmt.Printf("Starting sub-run, tag: %s\n", RunTag)
}

func UnsetRunSubTag() {
	fmt.Printf("Ended sub-run, tag: %s\n", RunTag)
	RunTag = RunBaseTag
}

// Wrappers to Infof,  Warnf, and Errorf test log functions
func TestLogInfof(format string, a ...interface{}) {
	if RunTag == "" {
		TestLog.Infof(format, a...)
	} else {
		TestLog.WithFields(logrus.Fields{"run": RunTag}).Infof(format, a...)
	}
}

func TestLogWarnf(format string, a ...interface{}) {
	if RunTag == "" {
		TestLog.Warnf(format, a...)
	} else {
		TestLog.WithFields(logrus.Fields{"run": RunTag}).Warnf(format, a...)
	}
}

func TestLogErrorf(format string, a ...interface{}) {
	if RunTag == "" {
		TestLog.Errorf(format, a...)
	} else {
		TestLog.WithFields(logrus.Fields{"run": RunTag}).Errorf(format, a...)
	}
}

// Wrapper for default print function, in case we want to do anything in the future to
// control whether or where things are printed
func Printf(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// print and/or log messages to the appropriate level
func Infof(format string, a ...interface{}) {
	if PrintInfo {
		fmt.Printf(format+"\n", a...)
	}
	if TestLog != nil {
		TestLogInfof(format, a...)
	}
}

func InfoOverridef(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	if TestLog != nil {
		TestLogInfof(format, a...)
	}
}

func Warnf(format string, a ...interface{}) {
	if PrintWarn {
		fmt.Printf("WARNING (run tag "+RunTag+"): "+format+"\n", a...)
	}
	if TestLog != nil {
		TestLogWarnf(format, a...)
	}
}

func Errorf(format string, a ...interface{}) {
	if PrintError {
		fmt.Printf("ERROR (run tag "+RunTag+"): "+format+"\n", a...)
	}
	if TestLog != nil {
		TestLogErrorf(format, a...)
	}
}

func Error(err error) {
	Errorf(err.Error())
}

func Resultsf(format string, a ...interface{}) {
	if PrintResults {
		fmt.Printf("(run tag "+RunTag+"): "+format+"\n", a...)
	}
	if TestLog != nil {
		TestLogInfof(format, a...)
	}
}

// print/log result and exit with specified code
func Exitf(rc int, format string, a ...interface{}) {
	var res string
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
		Resultsf(res+": "+format, a...)
	} else {
		Resultsf(res)
	}
	if Log != nil {
		Log.Exit(rc)
	} else {
		os.Exit(rc)
	}
}

func Usagef(format string, a ...interface{}) {
	Exitf(2, format, a...)
}

func Successf(format string, a ...interface{}) {
	Exitf(0, format, a...)
}

func Success() {
	Exitf(0, "")
}

func Failuref(format string, a ...interface{}) {
	Exitf(1, format, a...)
}

func Failure() {
	Exitf(1, "")
}

func ArtifactCommand(label, cmdName string, cmdArgs ...string) {
	if len(artifactDirectory) == 0 {
		return
	}
	t := time.Now()
	outfilename := artifactDirectory + "/" + artifactFilePrefix + label + "-" + t.Format(time.RFC3339Nano) + ".txt"
	cmdStr := cmdName + strings.Join(cmdArgs, " ")
	Infof("Running command: %s", cmdStr)
	cmd := exec.Command(cmdName, cmdArgs...)
	Infof("Storing output in %s", outfilename)
	outfile, err := os.Create(outfilename)
	if err != nil {
		Warnf("Error creating output file; %s", err.Error())
		return
	}
	defer outfile.Close()
	cmd.Stdout, cmd.Stderr = outfile, outfile
	err = cmd.Start()
	if err != nil {
		Warnf("Error starting command; %s", err.Error())
		return
	}
	err = cmd.Wait()
	if err != nil {
		Warnf("Command failed; %s", err.Error())
	} else {
		Infof("Command completed without error")
	}
}

func ArtifactDescribeNodes() {
	ArtifactCommand("k8s-describe-nodes", "kubectl", "describe", "nodes")
}

func ArtifactGetAllPods() {
	ArtifactCommand("k8s-get-pods", "kubectl", "get", "pods", "-A", "-o", "wide", "--show-labels=true")
}

func ArtifactGetAllPvcs() {
	ArtifactCommand("k8s-get-pvc", "kubectl", "get", "pvc", "-A", "-o", "wide", "--show-kind=true", "--show-labels=true")
}

func ArtifactDescribeNamespacePods(namespace string, podNames ...string) {
	for _, podName := range podNames {
		describeLabel := "k8s-describe-pod-" + namespace + "-" + podName
		ArtifactCommand(describeLabel, "kubectl", "describe", "pod", "-n", namespace, podName, "--show-events=true")
		logsLabel := "k8s-logs-" + namespace + "-" + podName
		ArtifactCommand(logsLabel, "kubectl", "logs", "-n", namespace, podName, "--all-containers=true", "--timestamps=true")
	}
}

func ArtifactDescribePods(podNames ...string) {
	ArtifactDescribeNamespacePods(NAMESPACE, podNames...)
}

func ArtifactDescribeNamespacePvcs(namespace string, pvcNames ...string) {
	for _, pvcName := range pvcNames {
		describeLabel := "k8s-describe-pvc-" + namespace + "-" + pvcName
		ArtifactCommand(describeLabel, "kubectl", "describe", "pvc", "-n", namespace, pvcName, "--show-events=true")
	}
}

func ArtifactDescribePvcs(pvcNames ...string) {
	ArtifactDescribeNamespacePvcs(NAMESPACE, pvcNames...)
}

func ArtifactsPodsPvcsNamespace(namespace string, podNames, pvcNames []string) {
	Infof("Collecting information about current kubernetes state")
	if len(podNames) > 0 {
		ArtifactDescribeNamespacePods(namespace, podNames...)
	}
	if len(pvcNames) > 0 {
		ArtifactDescribeNamespacePvcs(namespace, pvcNames...)
	}
	ArtifactGetAllPods()
	ArtifactGetAllPvcs()
	ArtifactDescribeNodes()
}

func ArtifactsPodsPvcs(podNames, pvcNames []string) {
	ArtifactsPodsPvcsNamespace(NAMESPACE, podNames, pvcNames)
}

func ArtifactsPodsNamespace(namespace string, podNames []string) {
	var noPvcs []string
	ArtifactsPodsPvcsNamespace(namespace, podNames, noPvcs)
}

func ArtifactsPods(podNames []string) {
	ArtifactsPodsNamespace(NAMESPACE, podNames)
}

// Determines index of string in slice, otherwise returns -1
func GetStringIndexInSlice(slice []string, element string) int {
	for i, str := range slice {
		if str == element {
			return i
		}
	}
	return -1
}

// determines if a string exists in a slice/array
func StringInArray(str string, services []string) bool {
	return (GetStringIndexInSlice(services, str) != -1)
}

// return a random string
func GetRandomString(len int) []byte {
	randomNum := func(min, max int) int {
		return min + rand.Intn(max-min)
	}
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomNum(97, 122))
	}
	return bytes
}

// methodEndpoint constructor
func newMethodEndpoint(parameters, summary string, responses []int) *endpointMethod {
	return &endpointMethod{
		parameters: parameters,
		summary:    summary,
		responses:  responses,
	}
}

// GetEndpoints() loads all CMS services endpoints into a data structure
// TODO: load from a config file or API spec
func GetEndpoints() map[string]map[string]*Endpoint {

	endpoints := make(map[string]map[string]*Endpoint)

	// BOS service endpoints
	endpoints["bos"] = make(map[string]*Endpoint)
	endpoints["bos"]["/"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Get API versions", []int{200}),
		},
		Url:     "/",
		Version: "v1",
	}
	endpoints["bos"]["v1"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Get API versions", []int{200, 500}),
		},
		Url:     "/apis/bos/v1",
		Version: "v1",
	}
	endpoints["bos"]["sessiontemplate"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET":  newMethodEndpoint("", "List session template", []int{200}),
			"POST": newMethodEndpoint("", "Create session template", []int{200, 400}),
		},
		Url:     "/apis/bos/v1/sessiontemplate",
		Version: "v1",
	}
	endpoints["bos"]["sessiontemplate/{session_template_id}"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET":    newMethodEndpoint("session_template_id", "Get session template by id", []int{200, 404}),
			"DELETE": newMethodEndpoint("session_template_id", "Delete a session template", []int{204, 404}),
		},
		Url:     "/apis/bos/v1/sessiontemplate/{session_template_id}",
		Version: "v1",
	}
	endpoints["bos"]["sessiontemplatetemplate"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Get example session template", []int{200}),
		},
		Url:     "/apis/bos/v1/sessiontemplatetemplate",
		Version: "v1",
	}
	endpoints["bos"]["session"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET":  newMethodEndpoint("", "List sessions", []int{200}),
			"POST": newMethodEndpoint("", "Create a session", []int{200, 400}),
		},
		Url:     "/apis/bos/v1/session",
		Version: "v1",
	}
	endpoints["bos"]["session/{session_id}"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET":    newMethodEndpoint("session_id", "Get sessions details by Id", []int{200, 404}),
			"DELETE": newMethodEndpoint("session_id", "Delete a session", []int{204, 404}),
		},
		Url:     "/apis/bos/v1/session/{session_id}",
		Version: "v1",
	}
	endpoints["bos"]["version"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Get API version", []int{200, 500}),
		},
		Url:     "/apis/bos/v1/version",
		Version: "v1",
	}

	// CFS service endpoints (only the ones we are testing with cmsdsev at the moment)
	endpoints["cfs"] = make(map[string]*Endpoint)
	endpoints["cfs"]["components"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve state of CFS components", []int{200, 400}),
		},
		Url:     "/apis/cfs/v2/components",
		Version: "v2",
	}
	endpoints["cfs"]["configurations"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve CFS configurations", []int{200, 400}),
		},
		Url:     "/apis/cfs/v2/configurations",
		Version: "v2",
	}
	endpoints["cfs"]["options"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve CFS options", []int{200}),
		},
		Url:     "/apis/cfs/v2/options",
		Version: "v2",
	}
	endpoints["cfs"]["sessions"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve all CFS sessions", []int{200}),
		},
		Url:     "/apis/cfs/v2/sessions",
		Version: "v2",
	}
	endpoints["cfs"]["healthz"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve CFS service health", []int{200, 500}),
		},
		Url:     "/apis/cfs/healthz",
		Version: "v2",
	}

	// CRUS service endpoints (only the ones we are testing with cmsdsev at the moment)
	endpoints["crus"] = make(map[string]*Endpoint)
	endpoints["crus"]["session"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve CRUS sessions", []int{200}),
		},
		Url:     "/apis/crus/session",
		Version: "v1",
	}

	// IMS service endpoints
	endpoints["ims"] = make(map[string]*Endpoint)
	endpoints["ims"]["images"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"DELETE": newMethodEndpoint("", "Delete all image records", []int{204, 500}),
			"GET":    newMethodEndpoint("", "Retrieve all image records", []int{200, 500}),
			"POST":   newMethodEndpoint("", "Create an image record", []int{201, 400, 422, 500}),
		},
		Url:     "/apis/ims/images",
		Version: "v2",
	}
	endpoints["ims"]["jobs"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"DELETE": newMethodEndpoint("", "Delete all job records", []int{204, 500}),
			"GET":    newMethodEndpoint("", "Retrieve all job records", []int{200, 500}),
			"POST":   newMethodEndpoint("", "Create a job record", []int{201, 400, 422, 500}),
		},
		Url:     "/apis/ims/jobs",
		Version: "v2",
	}
	endpoints["ims"]["public_keys"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"DELETE": newMethodEndpoint("", "Delete all public key records", []int{204, 500}),
			"GET":    newMethodEndpoint("", "Retrieve all public key records", []int{200, 500}),
			"POST":   newMethodEndpoint("", "Create a public key record", []int{201, 400, 422, 500}),
		},
		Url:     "/apis/ims/public-keys",
		Version: "v2",
	}
	// The status codes for these recipes IMS endpoints don't match the IMS openapi spec currently,
	// because of bug CASMCMS-5225. The status codes here reflect the reality of how IMS works.
	endpoints["ims"]["recipes"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"DELETE": newMethodEndpoint("", "Delete all recipe records", []int{204, 500}),
			"GET":    newMethodEndpoint("", "Retrieve all recipe records", []int{200, 500}),
			"POST":   newMethodEndpoint("", "Create a recipe record", []int{201, 400, 422, 500}),
		},
		Url:     "/apis/ims/recipes",
		Version: "v2",
	}
	endpoints["ims"]["version"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Get API version", []int{200, 500}),
		},
		Url:     "/apis/ims/version",
		Version: "v2",
	}
	endpoints["ims"]["live"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve IMS Liveness Probe", []int{200, 500}),
		},
		Url:     "/apis/ims/healthz/live",
		Version: "v2",
	}
	endpoints["ims"]["ready"] = &Endpoint{
		Methods: map[string]*endpointMethod{
			"GET": newMethodEndpoint("", "Retrieve IMS Readiness Probe", []int{200, 500}),
		},
		Url:     "/apis/ims/healthz/ready",
		Version: "v2",
	}

	return endpoints
}

// PrintEndpoints() prints services endpoints
// Optionally provided an argument will print only given service endpoint
func PrintEndpoints(service string, params ...string) {
	var search bool = false
	var key string

	// get key for searching through endpoints
	if len(params) > 0 {
		key = params[0]
		search = true
	}

	m := GetEndpoints()
	for k := range m {
		for i := range m[k] {
			if search == true {
				if i != key {
					continue
				}
			}
			if i != "/" {
				fmt.Printf("/%s:\n", i)
			} else {
				fmt.Printf("%s:\n", i)
			}

			// print methods
			for j := range m[k][i].Methods {
				fmt.Printf("     %s:\n", j)
				fmt.Printf("          summary: %s\n", m[k][i].Methods[j].summary)
				fmt.Printf("          responses: %v\n", m[k][i].Methods[j].responses)
				if len(m[k][i].Methods[j].parameters) > 0 {
					fmt.Printf("          parameters: %s\n", m[k][i].Methods[j].parameters)
				}
			}
			fmt.Printf("  url: %s\n", m[k][i].Url)
			fmt.Printf("  version: %s\n", m[k][i].Version)
		}
	}
}

// Restful() performs CMS RESTful calls
// returns true on success, otherwise false
// returns populated message string if success == false
func Restful(method, url string, params Params) (*resty.Response, error) {
	var err error
	var resp *resty.Response

	client := resty.New()

	switch method {
	case "GET":
		client.SetHeaders(map[string]string{
			"Accept":     "application/json",
			"User-Agent": "cmsdev",
		})
		resp, err = client.R().
			SetAuthToken(params.Token).
			Get(url)
	case "POST":
		if len(params.JsonStr) != 0 {
			// payload passed as string
			resp, err = client.R().
				SetAuthToken(params.Token).
				SetHeader("Content-Type", "application/json").
				SetBody(params.JsonStr).
				Post(url)
		} else {
			// payload passed as byte array
			resp, err = client.R().
				SetAuthToken(params.Token).
				SetHeader("Content-Type", "application/json").
				SetBody(params.JsonStrArray).
				Post(url)
		}
	case "PATCH":
		client.SetHeaders(map[string]string{
			"Accept":     "application/json",
			"User-Agent": "cmsdev",
		})
		resp, err = client.R().
			SetAuthToken(params.Token).
			SetBody(params.JsonStrArray).
			Patch(url)
	case "DELETE":
		client.SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"User-Agent":   "cmsdev",
		})
		resp, err = client.R().
			SetAuthToken(params.Token).
			Delete(url)
	}

	return resp, err

}

// If format is not blank, call Infof with format + a
// In verbose mode, also print green OK
func VerboseOkayf(format string, a ...interface{}) {
	if len(format) > 0 {
		Infof(format, a...)
	}
	if PrintVerbose {
		c.HiGreen("OK")
	}
}

func VerboseOkay() {
	VerboseOkayf("")
}

// If format is not blank, call Errorf with format + a
// In verbose mode, also print red Failed
func VerboseFailedf(format string, a ...interface{}) {
	if len(format) > 0 {
		Errorf(format, a...)
	}
	if PrintVerbose {
		c.Red("Failed")
	}
}

func VerboseFailed() {
	VerboseFailedf("")
}

func Verbosef(format string, a ...interface{}) {
	if PrintVerbose {
		fmt.Printf(format+"\n", a...)
	}
}

// Print a dividing line to stdout if in verbose mode
func VerbosePrintDivider() {
	Verbosef("---\n")
}

// pretty print resty json responses
func PrettyPrintJSON(resp *resty.Response) {
	if !PrintVerbose {
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

func CreateDirectoryIfNeeded(path string) error {
	// First see if the path already exists
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		// It does not exist, so let's create it
		return os.MkdirAll(path, 0755)
	} else if err != nil {
		return err
	}

	// It exists, so make sure it is a directory
	if fileInfo.IsDir() {
		return nil
	}
	return fmt.Errorf("Path exists but is not a directory: %s", path)
}

// create log file and directory provided by path if one does not exist
// if no path is provided, use DEFAULT_LOG_FILE_DIR
func CreateLogFile(path, service, version string, logs, retry, quiet, verbose bool) {
	var err error
	if verbose {
		PrintVerbose = true
	} else if quiet {
		PrintInfo, PrintWarn = false, false
	}
	if !logs {
		return
	} else if len(path) == 0 {
		path = DEFAULT_LOG_FILE_DIR
	}
	err = CreateDirectoryIfNeeded(path)
	if err != nil {
		fmt.Printf("Error with log directory: %s\n", path)
		panic(err)
	}
	logfile := path + "/cmsdev.log"
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	Log = logrus.New()
	if err != nil {
		panic(err)
	}
	// We want nanosecond precision in log file entries
	Log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	Log.SetOutput(f)
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
	RunBaseTag = AlnumString(5)
	RunTag = RunBaseTag
	TestLog = Log.WithFields(logrus.Fields{"version": version, "service": service, "args": strings.Join(args, ",")})
	Infof("cmsdev starting")
	fmt.Printf("Starting main run, version: %s, tag: %s\n", version, RunTag)
}

func InitArtifacts(service string) {
	artifactDirectory = os.Getenv("ARTIFACTS")
	if len(artifactDirectory) == 0 {
		Warnf("ARTIFACTS environment variable not set; no artifacts will be saved")
		return
	}
	err := CreateDirectoryIfNeeded(artifactDirectory)
	if err != nil {
		Warnf(err.Error())
		Warnf("Error with artifact directory \"%s\"; no artifacts will be saved", artifactDirectory)
		artifactDirectory = ""
		return
	}
	artifactFilePrefix = service + "-"
	Infof("artifactDirectory=%s, artifactFilePrefix=%s", artifactDirectory, artifactFilePrefix)
}

func init() {
	// Set default values
	Log, TestLog = nil, nil
	PrintInfo, PrintWarn, PrintError, PrintResults = true, true, true, true
	PrintVerbose = false
	RunTag, RunBaseTag, artifactDirectory, artifactFilePrefix = "", "", "", ""
}
