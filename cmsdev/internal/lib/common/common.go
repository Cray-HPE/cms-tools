//
//  MIT License
//
//  (C) Copyright 2019-2025 Hewlett Packard Enterprise Development LP
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
 * common.go
 *
 * Library of general utility functions
 *
 */

package common

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	resty "gopkg.in/resty.v1"
)

const BASEHOST = "api-gw-service-nmn.local"
const BASEURL = "https://" + BASEHOST
const LOCALHOST = "http://localhost:5000"
const NAMESPACE string = "services"

// List of RPMs version captured at the start of the test. In case of failure, rpm -qa will be captured.
var RPMLIST = []string{
	"craycli",
	"docs-csm",
	"csm-testing",
	"goss-servers",
}

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

// List of Kubernetes things to collect for debug in case of failure
var kubernetesThingsToCollect = []string{
	"nodes",
	"namespaces",
	"pods",
	"pv",
	"pvc",
	"services",
	"daemonsets",
	"statefulsets",
	"deployments",
	"etcd",
	"configmaps",
	"secrets",
	"endpoints",
	"postgresqls",
	"cronjobs",
	"jobs",
	"sealedsecrets",
	"etcdbackups",
}

var TmpDir string

var runTags []string
var runTag, testService string
var runStartTimes []time.Time

var artifactDirectory, artifactFilePrefix string
var artifactDirectoryCreated, artifactsLogged bool

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

// Capture csmdev version used for testing

func GetPackageVersion(packageName string) string {
	cmd := exec.Command("rpm", "-q", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		Warnf("Package %s not installed", packageName)
		return ""
	}
	installedPkg := strings.TrimSpace(string(output))
	return installedPkg
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
	artifactsLogged = true
	err = cmd.Wait()
	if err != nil {
		Warnf("Command failed; %s", err.Error())
	} else {
		Debugf("Command completed without error")
	}
}

func ArtifactGetAdditionalInfo() {
	ArtifactCommand("rpm-qa", "rpm", "-qa")
}

func ArtifactGetAllThings(thing string) {
	ArtifactCommand("k8s-get-"+thing, "kubectl", "get", thing, "-A", "-o", "wide", "--show-labels=true")
}

func ArtifactDescribeNodes() {
	ArtifactCommand("k8s-describe-nodes", "kubectl", "describe", "nodes")
}

func ArtifactDescribeNamespacePods(namespace string, podNames []string) {
	Infof("Collecting information about current Kubernetes state of following '%s' namespace pods: %s",
		namespace, strings.Join(podNames, " "))
	for _, podName := range podNames {
		describeLabel := "k8s-describe-pod-" + namespace + "-" + podName
		ArtifactCommand(describeLabel, "kubectl", "describe", "pod", "-n", namespace, podName, "--show-events=true")
		logsLabel := "k8s-logs-" + namespace + "-" + podName
		ArtifactCommand(logsLabel, "kubectl", "logs", "-n", namespace, podName, "--all-containers=true", "--timestamps=true", "--prefix=true")
	}
}

func ArtifactDescribeNamespacePvcs(namespace string, pvcNames []string) {
	Infof("Collecting information about current Kubernetes state of following '%s' namespace PVCs: %s",
		namespace, strings.Join(pvcNames, " "))
	for _, pvcName := range pvcNames {
		describeLabel := "k8s-describe-pvc-" + namespace + "-" + pvcName
		ArtifactCommand(describeLabel, "kubectl", "describe", "pvc", "-n", namespace, pvcName, "--show-events=true")
	}
}

func ArtifactsKubernetes() {
	Infof("Collecting information about current Kubernetes state")
	for _, thing := range kubernetesThingsToCollect {
		ArtifactGetAllThings(thing)
	}
	ArtifactDescribeNodes()
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

// doRest() performs RESTful calls using the provided client and parameters.
func doRest(method, url string, params Params, client *resty.Client) (*resty.Response, error) {
	var err error
	var resp *resty.Response

	switch method {
	case "GET":
		resp, err = client.R().Get(url)
	case "POST":
		if len(params.JsonStr) != 0 {
			// payload passed as string
			resp, err = client.R().
				SetBody(params.JsonStr).
				Post(url)
		} else {
			// payload passed as byte array
			resp, err = client.R().
				SetBody(params.JsonStrArray).
				Post(url)
		}
	case "PATCH":
		resp, err = client.R().
			SetBody(params.JsonStrArray).
			Patch(url)
	case "DELETE":
		resp, err = client.R().Delete(url)
	}

	return resp, err
}

// Restful() performs CMS RESTful calls
func Restful(method, url string, params Params) (*resty.Response, error) {
	client := resty.New()
	client.SetHeaders(map[string]string{
		"Accept":       "application/json",
		"User-Agent":   "cmsdev",
		"Content-Type": "application/json",
	})
	client.SetAuthToken(params.Token)
	return doRest(method, url, params, client)
}

// Restful() performs CMS RESTful calls on behalf of the specified tenant
func RestfulTenant(method, url, tenant string, params Params) (*resty.Response, error) {
	client := resty.New()
	client.SetHeaders(map[string]string{
		"Accept":           "application/json",
		"User-Agent":       "cmsdev",
		"Content-Type":     "application/json",
		"Cray-Tenant-Name": tenant,
	})
	client.SetAuthToken(params.Token)
	return doRest(method, url, params, client)
}

func CreateDirectoryIfNeeded(path string) (error, bool) {
	// bool is True if we create the directory, false if it
	// already exists.

	// First see if the path already exists
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		// It does not exist, so let's create it
		return os.MkdirAll(path, 0755), true
	} else if err != nil {
		return err, false
	}

	// It exists, so make sure it is a directory
	if fileInfo.IsDir() {
		return nil, false
	}
	return fmt.Errorf("Path exists but is not a directory: %s", path), false
}

func RemoveEmptyDirectory(path string) error {
	// First see if the path already exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		// It does not exist, so nothing to do
		return fmt.Errorf("Path does not exist: %s", path)
	}
	// It exists, so make sure it is a directory
	if !dirInfo.IsDir() {
		return fmt.Errorf("Path exists but is not a directory: %s", path)
	}
	// Remove it
	return os.Remove(path)
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
	var err error

	artifactDirectory = os.Getenv("ARTIFACTS")
	if len(artifactDirectory) == 0 {
		if len(logFileDir) == 0 {
			Warnf("ARTIFACTS environment variable not set and test logging disabled; no artifacts will be saved")
			return
		}
		// Default to log_directory/timestamp
		artifactDirectory = logFileDir + "/" + "artifacts-" + time.Now().Format(time.RFC3339Nano)
		Debugf("ARTIFACTS environment variable not set. Defaulting to '%s'", artifactDirectory)
	} else {
		Debugf("ARTIFACTS environment variable set to '%s'", artifactDirectory)
	}
	err, artifactDirectoryCreated = CreateDirectoryIfNeeded(artifactDirectory)
	if err != nil {
		Warnf(err.Error())
		Warnf("Error with artifact directory \"%s\"; no artifacts will be saved", artifactDirectory)
		artifactDirectory = ""
		return
	}
	Infof("artifactDirectory=%s", artifactDirectory)
}

func CompressArtifacts() {
	if len(artifactDirectory) == 0 {
		// No artifact directory set, so nothing to do.
		return
	}
	if !artifactsLogged {
		// No artifacts logged, but there is an artifact directory.
		// If we created it, then we will delete it.
		if !artifactDirectoryCreated {
			Debugf("No artifacts saved. We did not create the artifact directory, so we will not remove it.")
			return
		}
		Infof("No artifacts saved. Removing empty artifact directory: '%s'", artifactDirectory)
		err := RemoveEmptyDirectory(artifactDirectory)
		artifactDirectory = ""
		if err != nil {
			Warnf(err.Error())
		}
		return
	}
	// Artifacts were logged, so compress them and delete the uncompressed artifacts
	compressedArtifactsFile := artifactDirectory + ".tgz"
	Infof("Compressing saved test artifacts to '%s'", compressedArtifactsFile)
	artifactParentDirectory := filepath.Dir(artifactDirectory)
	artifactDirectoryBasename := filepath.Base(artifactDirectory)
	Debugf("artifactParentDirectory=%s, artifactDirectoryBasename=%s", artifactParentDirectory,
		artifactDirectoryBasename)
	cmdResult, err := RunName("tar", "-C", artifactParentDirectory, "--remove-files",
		"-czvf", compressedArtifactsFile, artifactDirectoryBasename)
	artifactDirectory = ""
	if err != nil {
		Warnf(err.Error())
		return
	}
	if cmdResult.Rc == 0 {
		// Command passed
		return
	}
	Warnf("Error compressing artifacts (tar return code = %d)", cmdResult.Rc)
}

// The caller of this function is responsible for removing the directory
func CreateTmpDir() (err error) {
	// Pass empty string for directory name to use the default tmp directory
	Debugf("Creating temporary directory")
	TmpDir, err = ioutil.TempDir("", "cmsdev-tmpdir")
	return
}

func DeleteTmpDir() {
	if len(TmpDir) == 0 {
		return
	}
	Debugf("Removing temporary directory: '%s'", TmpDir)
	if err := os.RemoveAll(TmpDir); err != nil {
		Warnf("Error removing temporary directory '%s': %v", TmpDir, err)
		return
	}
	Debugf("Successfully removed temporary directory: '%s'", TmpDir)
	TmpDir = ""
	return
}

func init() {
	// Set default values
	runStartTimes = append(runStartTimes, time.Now())
	runTags = append(runTags, AlnumString(5))
	logFile, testLog = nil, nil
	printInfo, printWarn, printError, printResults = true, true, true, true
	artifactDirectoryCreated, artifactsLogged, printVerbose = false, false, false
	runTag, artifactDirectory, artifactFilePrefix, testService, logFileDir, TmpDir = "", "", "", "", "", ""
	// Call the init function for the printlog source file
	printlogInit()
}
