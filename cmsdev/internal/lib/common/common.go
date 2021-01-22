/*
Copyright 2019, Cray Inc.  All Rights Reserved.
Author: Torrey Cuthbert <tcuthbert@cray.com>
*/
package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"math/rand"
	"strings"
	c "github.com/fatih/color"
	"github.com/go-resty/resty"
)

const BASEURL = "https://api-gw-service-nmn.local"
const LOCALHOST = "http://localhost:5000"
const NAMESPACE string = "services"
const DEFAULT_LOG_FILE_DIR string = "/opt/cray/tests"

// struct to hold endpoint METHOD operation details
type endpointMethod struct {
	parameters string    				// TODO: change to interface
	responses []int
	summary string
}
 
// data structure to endpoints, URL, descriptions
type Endpoint struct {
	Methods map[string]*endpointMethod  // endpoint METHOD data
	Url string							// url string appended to base to reach endpoint
	Version string 		`default:"v1"`	// endpoint version 
}

// Restful() parameters
type Params struct {
	Token string
	JsonStrArray []byte
	JsonStr string 
}

// pod name prefixes
var PodServiceNamePrefixes = map[string]string{
	"bos": "cray-bos",
	"bosPvc": "cray-bos-etcd",
	"cfs": "cray-cfs",
	"cfs-api": "cray-cfs-api",
	"cfs-operator": "cray-cfs-operator",
	"cfsServices": "^(cray-cfs-operator|cray-cfs-api)",
	"conman": "cray-conman",
	"conmanPvc": "cray-conman-data-claim",
	"ims": "cray-ims",
	"imsPvc": "cray-ims-data-claim",
	"ipxe": "cray-ipxe",
	"tftp": "cray-tftp",
	"tftpPvc": "cray-tftp-shared-pvc",
	"vcs": "gitea-vcs",
}

// list of CMS services
var CMSServices = []string {
	"bos", 
	"cfs", 
	"conman", 
	"gitea",
	"ims", 
	"ipxe", 
	"tftp", 
	"vcs",
}

// list of supported CMS services test types
var CMSServicesTestTypes = []string {
	"api",
	"smoke",
	"ct",
}

// log file handle
var Log *logrus.Logger
var TestLog *logrus.Entry
var PrintInfo, PrintWarn, PrintError, PrintResults, PrintVerbose bool
var RunTag, RunBaseTag string

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

// print and/or log messages to the appropriate level
func Infof(format string, a ...interface{}) {
	if PrintInfo { fmt.Printf(format+"\n", a...) }
	if TestLog != nil { TestLogInfof(format, a...) }
}

func Warnf(format string, a ...interface{}) {
	if PrintWarn { fmt.Printf("WARNING (run tag "+RunTag+"): "+format+"\n", a...) }
	if TestLog != nil { TestLogWarnf(format, a...) }
}

func Errorf(format string, a ...interface{}) {
	if PrintError { fmt.Printf("ERROR (run tag "+RunTag+"): "+format+"\n", a...) }
	if TestLog != nil { TestLogErrorf(format, a...) }
}

func Error(err error) {
	Errorf(err.Error())
}

func Resultsf(format string, a ...interface{}) {
	if PrintResults { fmt.Printf("(run tag "+RunTag+"): "+format+"\n", a...) }
	if TestLog != nil { TestLogInfof(format, a...) }
}

// print/log result and exit with specified code
func Exitf(rc int, format string, a ...interface{}) {
	var res string
	switch rc {
		case 0:     res = "SUCCESS"
		case 1:     res = "FAILURE"
		case 2:     res = "USAGE ERROR"
		default:    res = "UNKNOWN ERROR"
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
	Exitf(2,format,a...)
}

func Successf(format string, a ...interface{}) {
	Exitf(0,format,a...)
}

func Success() {
	Exitf(0,"")
}

func Failuref(format string, a ...interface{}) {
	Exitf(1,format,a...)
}

func Failure() {
	Exitf(1,"")
}

// determines if a string exists in a slice/array
func StringInArray(str string, services []string) bool {
	for _, v := range services {
		if v == str {
			return true
		}
	}
	return false
}

// return a random string
func GetRandomString(len int) []byte {
	randomNum := func(min, max int) int {
		return min + rand.Intn(max-min)
	}
	bytes := make([]byte, len)
	for i:= 0; i < len; i++ {
		bytes[i] = byte(randomNum(97, 122))
	}
	return bytes
}

// methodEndpoint constructor
func newMethodEndpoint(parameters, summary string, responses []int) *endpointMethod {
	return &endpointMethod {
		parameters: parameters,
		summary: summary,
		responses: responses,
	}
}

// GetEndpoints() loads all CMS services endpoints into a data structure
// TODO: load from a config file or API spec
func GetEndpoints() map[string]map[string]*Endpoint {

	endpoints := make(map[string]map[string]*Endpoint)

	// BOS service endpoints
	endpoints["bos"] = make(map[string]*Endpoint)
	endpoints["bos"]["/"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Get API versions", []int{200}), 
			},
		Url: "/",
		Version: "v1",
	}
	endpoints["bos"]["v1"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Get API versions", []int{200, 500}), 
			},
		Url: "/apis/bos/v1",
		Version: "v1",
	}
	endpoints["bos"]["sessiontemplate"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "List session template", []int{200}), 
			"POST": newMethodEndpoint("", "Create session template", []int{200, 400}), 
			},
		Url: "/apis/bos/v1/sessiontemplate",
		Version: "v1",
	}
	endpoints["bos"]["sessiontemplate/{session_template_id}"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("session_template_id", "Get session template by id", []int{200, 404}), 
			"DELETE": newMethodEndpoint("session_template_id", "Delete a session template", []int{204, 404}), 
			},
		Url: "/apis/bos/v1/sessiontemplate/{session_template_id}",
		Version: "v1",
	}
	endpoints["bos"]["session"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "List sessions", []int{200}), 
			"POST": newMethodEndpoint("", "Create a session", []int{200, 400}), 
			},
		Url: "/apis/bos/v1/session",
		Version: "v1",
	}
	endpoints["bos"]["session/{session_id}"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("session_id", "Get sessions details by Id", []int{200, 404}), 
			"DELETE": newMethodEndpoint("session_id", "Delete a session", []int{204, 404}), 
			},
		Url: "/apis/bos/v1/session/{session_id}",
		Version: "v1",
	}
	endpoints["bos"]["version"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Get API version", []int{200, 500}), 
			},
		Url: "/apis/bos/v1/version",
		Version: "v1",
	}

	// IMS service endpoints
	endpoints["ims"] = make(map[string]*Endpoint)
	endpoints["ims"]["images"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"DELETE": newMethodEndpoint("", "Delete all image records", []int{204, 500}), 
			"GET": newMethodEndpoint("", "Retrieve all image records", []int{200, 500}), 
			"POST": newMethodEndpoint("", "Create an image record", []int{201, 400, 422, 500}), 
			},
		Url: "/apis/ims/images",
		Version: "v2",
	}
	endpoints["ims"]["jobs"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"DELETE": newMethodEndpoint("", "Delete all job records", []int{204, 500}), 
			"GET": newMethodEndpoint("", "Retrieve all job records", []int{200, 500}), 
			"POST": newMethodEndpoint("", "Create a job record", []int{201, 400, 422, 500}), 
			},
		Url: "/apis/ims/jobs",
		Version: "v2",
	}
	endpoints["ims"]["public_keys"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"DELETE": newMethodEndpoint("", "Delete all public key records", []int{204, 500}), 
			"GET": newMethodEndpoint("", "Retrieve all public key records", []int{200, 500}), 
			"POST": newMethodEndpoint("", "Create a public key record", []int{201, 400, 422, 500}), 
			},
		Url: "/apis/ims/public-keys",
		Version: "v2",
	}
	// The status codes for these recipes IMS endpoints don't match the IMS openapi spec currently,
	// because of bug CASMCMS-5225. The status codes here reflect the reality of how IMS works.
	endpoints["ims"]["recipes"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"DELETE": newMethodEndpoint("", "Delete all recipe records", []int{204, 500}), 
			"GET": newMethodEndpoint("", "Retrieve all recipe records", []int{200, 500}), 
			"POST": newMethodEndpoint("", "Create a recipe record", []int{201, 400, 422, 500}), 
			},
		Url: "/apis/ims/recipes",
		Version: "v2",
	}
	endpoints["ims"]["version"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Get API version", []int{200, 500}), 
			},
		Url: "/apis/ims/version",
		Version: "v2",
	}
	endpoints["ims"]["live"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Retrieve IMS Liveness Probe", []int{200, 500}), 
			},
		Url: "/apis/ims/healthz/live",
		Version: "v2",
	}
	endpoints["ims"]["ready"] = &Endpoint {
		Methods: map[string]*endpointMethod {
			"GET": newMethodEndpoint("", "Retrieve IMS Readiness Probe", []int{200, 500}), 
			},
		Url: "/apis/ims/healthz/ready",
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
	for k, _ := range m {
		for i, _ := range m[k] {
			if search == true {
				if i != key {
					continue
				}
			}
			if i != "/" {
				fmt.Printf("/%s:\n" ,i)
			} else {
				fmt.Printf("%s:\n", i)
			}

			// print methods
			for j, _ := range m[k][i].Methods {
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
			"Accept": "application/json",
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
			"Accept": "application/json",
			"User-Agent": "cmsdev",
		})
		resp, err = client.R().
	  		SetAuthToken(params.Token).
			SetBody(params.JsonStrArray).
	  		Patch(url)
	case "DELETE":
		client.SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"User-Agent": "cmsdev",
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
	if !PrintVerbose { return }
	var prettyJSON bytes.Buffer

	err := json.Indent(&prettyJSON, resp.Body(), "", "   ")
	if err != nil {
		fmt.Printf("%v\n", resp)
	} else {
		fmt.Printf("%s\n", strings.TrimSpace(string(prettyJSON.Bytes())))
	}
}

func AddString(slice []string, element string) []string {
	n := len(slice)
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

// create log file and directory provided by path if one does not exist
// if no path is provided, use DEFAULT_LOG_FILE_DIR
func CreateLogFile(path, stage, service string, api, ct, local, logs, smoke, verbose bool) {
	var err error
	if verbose { 
		PrintVerbose = true
	} else if ct { 
		// These are disabled for CT tests unless verbose flag is present
		PrintInfo, PrintWarn = false, false
	}
	if !logs { 
		return
	} else if len(path) == 0 {
		path = DEFAULT_LOG_FILE_DIR
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}
	logfile := path + "/cmsdev.log"
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	Log = logrus.New()
	if err != nil {
		panic(err)
	}
	Log.SetOutput(f)
	args := make([]string, 0, 5)
	if api      { args = AddString(args, "api") }
	if ct       { args = AddString(args, "ct") }
	if local    { args = AddString(args, "local") }
	if smoke    { args = AddString(args, "smoke") }
	if verbose  { args = AddString(args, "verbose") }
	RunBaseTag = AlnumString(5)
	RunTag = RunBaseTag
	TestLog = Log.WithFields(logrus.Fields{"service": service, "stage": stage, "args": strings.Join(args, ",")})
	Infof("cmsdev starting")
	fmt.Printf("Starting main run, tag: %s\n", RunTag)
}

func init() {
	// Set default values
	Log, TestLog = nil, nil
	PrintInfo, PrintWarn, PrintError, PrintResults = true, true, true, true
	PrintVerbose = false
	RunTag, RunBaseTag = "", ""
}