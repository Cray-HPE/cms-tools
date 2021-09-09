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
	"fmt"
	"github.com/go-resty/resty"
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

// List of services which are just aliases for other
// services, as far as cmsdev testing is concerned.
// Note that if service a, b, and c are identical
// to cmsdev, all but one of them should be listed here.
// That is, do not list all three of them here. It doesn't
// matter which one you omit.
var CMSServicesDuplicates = []string{
	"gitea",
	"ipxe",
}

var runTags []string
var runStartTimes []time.Time
var artifactDirectory, artifactFilePrefix, runTag, testService string

// Set and unset the run sub-tag
func SetRunSubTag(tag string) {
	if len(runTags) < 1 {
		return
	}
	runStartTimes = append(runStartTimes, time.Now())
	runTags = append(runTags, tag)
	runTag = strings.Join(runTags, "-")
	Printf("Starting sub-run, tag: %s\n", runTag)
}

func UnsetRunSubTag() {
	if len(runTags) <= 1 {
		return
	}
	Printf("Ended sub-run, tag: %s (duration: %v)\n", runTag, time.Since(runStartTimes[len(runStartTimes)-1]))
	runStartTimes = runStartTimes[:len(runStartTimes)-1]
	runTags = runTags[:len(runTags)-1]
	runTag = strings.Join(runTags, "-")
}

func ChangeRunSubTag(tag string) {
	UnsetRunSubTag()
	SetRunSubTag(tag)
}

func ArtifactCommand(label, cmdName string, cmdArgs ...string) {
	if len(artifactDirectory) == 0 {
		return
	}
	t := time.Now()
	outfilename := artifactDirectory + "/" + artifactFilePrefix + label + "-" + t.Format(time.RFC3339Nano) + ".txt"
	cmdStr := cmdName + strings.Join(cmdArgs, " ")
	Debugf("Running command: %s", cmdStr)
	cmd := exec.Command(cmdName, cmdArgs...)
	Debugf("Storing output in %s", outfilename)
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
		Debugf("Command completed without error")
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

func SetTestService(service string) {
	if testService == "" {
		SetRunSubTag(service)
	} else {
		ChangeRunSubTag(service)
	}
	if len(artifactDirectory) > 0 {
		artifactFilePrefix = service + "-"
	}
	testService = service
}

func UnsetTestService() {
	if testService == "" {
		return
	}
	UnsetRunSubTag()
	testService = ""
	if len(artifactDirectory) > 0 {
		artifactFilePrefix = ""
	}
}

func InitArtifacts() {
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
	Infof("artifactDirectory=" + artifactDirectory)
}

func init() {
	// Set default values
	runStartTimes = append(runStartTimes, time.Now())
	runTags = append(runTags, AlnumString(5))
	logFile, testLog = nil, nil
	printInfo, printWarn, printError, printResults = true, true, true, true
	printVerbose = false
	runTag, artifactDirectory, artifactFilePrefix, testService = "", "", "", ""
	// Call the init function for the printlog source file
	printlogInit()
}
