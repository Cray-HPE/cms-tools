//
//  MIT License
//
//  (C) Copyright 2020-2025 Hewlett Packard Enterprise Development LP
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

package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	resty "gopkg.in/resty.v1"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
)

// struct to hold current pod status
type PodStats struct {
	Phase                       string
	ContainerStateWaitingReason map[string]string
}

// struct to hold container environment variables
type ContainerEnvVar struct {
	Name, Value string
}

var KubectlPath, VcsUser, VcsPass string
var kubeConfig *rest.Config

// Retry constants for k8s
const MaxRetries = 45
const RetryIntervalSeconds = 2

func getKubeconfigEnvVar() (kubeconfigEnvVar string) {
	kubeconfigEnvVar = os.Getenv("KUBECONFIG")
	if len(kubeconfigEnvVar) == 0 {
		kubeconfigEnvVar = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		common.Debugf("KUBECONFIG env var not set. Using default = %s", kubeconfigEnvVar)
		return
	}
	common.Debugf("Found env var KUBECONFIG = %s", kubeconfigEnvVar)
	return
}

func getKubeConfig() (*rest.Config, error) {
	var err error
	if kubeConfig == nil {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", getKubeconfigEnvVar())
	}
	return kubeConfig, err
}

func GetClientset() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, err
}

func GetKubectlPath() (string, error) {
	if len(KubectlPath) == 0 {
		common.Debugf("Trying to look up path of kubectl")
		path, err := exec.LookPath("kubectl")
		if err != nil {
			return "", err
		} else if len(path) == 0 {
			return "", fmt.Errorf("Empty path returned for kubectl path")
		}
		KubectlPath = path
	}
	return KubectlPath, nil
}

func GetTenants() (tenantList []string, err error) {
	var path string
	var cmdList = []string{"get", "tenants", "-n", "tenants", "-o", "custom-columns=:.metadata.name ", "--no-headers"}
	path, err = GetKubectlPath()
	if err != nil {
		return
	}
	cmd := exec.Command(path, cmdList...)
	common.Debugf("Running command: %s", cmd)
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	// Trim whitespace
	tenantListString := strings.TrimSpace(string(cmdOut))
	if len(tenantListString) == 0 {
		common.Infof("No tenants defined on the system")
		return
	}
	// Split by newline
	tenantList = strings.Split(tenantListString, "\n")
	common.Infof("The following %d tenants are defined on the system: %s", len(tenantList), strings.Join(tenantList, ", "))
	return
}

func RunCommandInContainer(podName, namespace, containerName string, cmdStrings ...string) (string, error) {
	k8sCmdList := [...]string{"exec", "-q", podName, "-n", namespace, "-c", containerName, "--stdin=false", "--"}
	cmdList := append(k8sCmdList[:], cmdStrings...)
	path, err := GetKubectlPath()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path, cmdList...)
	common.Debugf("Running command: %s", cmd)
	cmdOut, err := cmd.CombinedOutput()
	if len(cmdOut) > 0 {
		common.Debugf("Command output: %s", cmdOut)
	} else {
		common.Debugf("No command output")
	}
	return string(cmdOut), err
}

func GetVcsUsernamePassword() (vcsUsername, vcsPassword string, err error) {
	if len(VcsUser) != 0 && len(VcsPass) != 0 {
		common.Debugf("Using cached values of vcs user and password")
		vcsUsername = VcsUser
		vcsPassword = VcsPass
		return
	}

	secret, err := GetSecret("services", "vcs-user-credentials")
	if err != nil {
		return
	}

	vcsUsername, err = GetDataFieldFromSecret(secret, "vcs_username")
	if err != nil {
		return
	}

	vcsPassword, err = GetDataFieldFromSecret(secret, "vcs_password")
	if err != nil {
		return
	}
	VcsUser = vcsUsername
	VcsPass = vcsPassword
	return
}

func GetOauthClientSecret() (string, error) {
	return GetSecretDataField("default", "admin-client-auth", "client-secret")
}

func GetAccessJSON(params ...string) ([]byte, error) {
	var err error
	var clientSecret string = ""

	if len(params) == 1 {
		// check if param is valid otherwise discard
		// add a regex here to check if string is valid
		clientSecret = params[0]
		re, _ := regexp.MatchString("^([0-9a-f]*-){4}[0-9a-f]*$", clientSecret)
		if re != true {
			return nil, errors.New("received invalid client-secret")
		}
	} else {
		clientSecretByte, err := GetOauthClientSecret()
		if err != nil {
			return nil, err
		}
		clientSecret = string(clientSecretByte)
	}
	client := resty.New()
	client.SetTimeout(common.API_TIMEOUT_SECONDS)
	client.SetHeader("Content-Type", "application/json")
	client.SetFormData(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "admin-client",
		"client_secret": clientSecret,
	})
	resp, err := client.R().Post("https://api-gw-service-nmn.local/keycloak/realms/shasta/protocol/openid-connect/token")
	if err != nil {
		return nil, err
	} else if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("POST to Keycloak: expected status code %d, got %d", http.StatusOK, resp.StatusCode())
	}
	return resp.Body(), nil
}

func GetAccessToken(params ...string) (string, error) {
	// TODO: clean this up. remove curl and replace with clientgo or a rest call
	var cmdOut []byte
	var err error
	var jsonData map[string]interface{}

	cmdOut, err = GetAccessJSON(params...)
	if err != nil {
		return "", err
	}
	common.Debugf("Parsing JSON object containing access token")
	if err = json.Unmarshal(cmdOut, &jsonData); err != nil {
		return "", err
	}
	return jsonData["access_token"].(string), nil
}

// Given a container, return a map of its environment variables
func GetEnvVars(container coreV1.Container) (envVars []ContainerEnvVar) {
	envVars = make([]ContainerEnvVar, 0, len(container.Env))
	for _, evar := range container.Env {
		envVars = append(envVars, ContainerEnvVar{Name: evar.Name, Value: evar.Value})
	}
	return
}

// Given a pod, return a list of its containers
func GetContainers(pod coreV1.Pod) (containers []coreV1.Container) {
	containers = make([]coreV1.Container, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c)
	}
	return
}

// Given a namespace and a cronjob name, verify that it exists
func VerifyCronJobExists(namespace, name string) error {
	clientset, err := GetClientset()
	if err != nil {
		return err
	}
	allCronJobs, err := clientset.BatchV1().CronJobs(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cj := range allCronJobs.Items {
		if cj.ObjectMeta.Name == name {
			return nil
		}
	}
	return fmt.Errorf("No kubernetes CronJob found in namespace %s with name %s", namespace, name)
}

// Given an optional regex, return an array of Nodes (whose name match the regex, if specified)
func GetNodes(params ...string) ([]coreV1.Node, error) {
	var nodes []coreV1.Node

	clientset, err := GetClientset()
	if err != nil {
		return nodes, err
	}
	allNodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nodes, err
	}
	for _, node := range allNodes.Items {
		if len(params) > 0 {
			match, _ := regexp.MatchString(params[0], node.ObjectMeta.Name)
			if match {
				nodes = append(nodes, node)
			}
			continue
		} else {
			nodes = append(nodes, node)
		}
	}
	return nodes, err
}

// Given an optional regex, return an array of Node names
func GetNodeNames(params ...string) ([]string, error) {
	var names []string

	nodes, err := GetNodes(params...)
	if err != nil {
		return names, err
	}
	for _, node := range nodes {
		names = append(names, node.ObjectMeta.Name)
	}
	return names, err
}

func GetDataFieldFromSecret(secret *coreV1.Secret, field_name string) (string, error) {
	common.Debugf("Retrieving '%s' data field from %s in namespace %s",
		field_name, secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
	fieldBytes, keyFound := secret.Data[field_name]
	if !keyFound {
		return "", fmt.Errorf("No '%s' data field found in Kubernetes secret %s in %s namespace",
			field_name, secret.ObjectMeta.Name, secret.ObjectMeta.Namespace)
	} else if len(fieldBytes) == 0 {
		return "", fmt.Errorf("Value for %s is an empty string", field_name)
	}
	return strings.TrimSpace(string(fieldBytes)), nil
}

// Given a namespace and a name, returns the matching secret
func GetSecret(namespace, name string) (*coreV1.Secret, error) {
	common.Debugf("Retrieving kubernetes secret %s in %s namespace", name, namespace)
	clientset, err := GetClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Secrets(namespace).Get(
		context.TODO(),
		name,
		v1.GetOptions{},
	)
}

func GetSecretDataField(namespace, name, field_name string) (string, error) {
	k8sSecret, err := GetSecret(namespace, name)
	if err != nil {
		return "", err
	}
	return GetDataFieldFromSecret(k8sSecret, field_name)
}

// Given a namespace and name, returns the matching configmap
func GetConfigMap(namespace, name string) (cm coreV1.ConfigMap, err error) {
	common.Debugf("Retrieving Kubernetes ConfigMap in namespace %s with name %s", namespace, name)

	clientset, err := GetClientset()
	if err != nil {
		return
	}

	cmlist, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return
	}

	for _, cm = range cmlist.Items {
		if cm.ObjectMeta.Name == name {
			common.Debugf("Found Kubernetes ConfigMap in namespace %s with name %s", namespace, name)
			return
		}
	}
	err = fmt.Errorf("No Kubernetes ConfigMap found in namespace %s with name %s", namespace, name)
	return
}

// Given a namespace, configmap name, and data field name, return the specified data field as a byte slice.
func GetConfigMapDataField(namespace, cm_name, field_name string) (field_bytes []byte, err error) {
	var cm coreV1.ConfigMap

	cm, err = GetConfigMap(namespace, cm_name)
	if err != nil {
		return
	}

	common.Debugf("Retrieve Data field '%s' from ConfigMap '%s' in namespace '%s'", field_name, cm_name, namespace)
	dataField, keyFound := cm.Data[field_name]
	if !keyFound {
		err = fmt.Errorf("No field named '%s' found in Kubernetes ConfigMap %s in namespace %s", field_name, cm_name, namespace)
		return
	}

	// Make sure we can convert the field to a byte slice
	common.Debugf("Convert %s field to byte slice", field_name)
	field_bytes = []byte(dataField)
	return
}

// Given a namespace and name, returns the matching service
func GetService(namespace, name string) (service coreV1.Service, err error) {
	clientset, err := GetClientset()
	if err != nil {
		return
	}
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return
	}
	for _, srv := range services.Items {
		if name == srv.GetName() {
			service = srv
			return
		}
	}
	err = fmt.Errorf("No kubernetes service found in namespace %s with name %s", namespace, name)
	return
}

// Given a namespace, and an optional regex, return an array of Pods (whose name match the regex, if specified)
func GetPods(namespace string, params ...string) ([]coreV1.Pod, error) {
	var pods []coreV1.Pod

	clientset, err := GetClientset()
	if err != nil {
		return pods, err
	}
	allPods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return pods, err
	}
	for _, pod := range allPods.Items {
		if len(params) > 0 {
			match, _ := regexp.MatchString(params[0], pod.GetName())
			if match {
				pods = append(pods, pod)
			}
			continue
		} else {
			pods = append(pods, pod)
		}
	}
	return pods, err
}

// Given a namespace and the name of a pod, return its start time
func GetPodStartTime(namespace, podName string) (nodeStart v1.Time, err error) {
	pods, err := GetPods(namespace, podName)
	if err != nil {
		return
	}
	for _, pod := range pods {
		if pod.GetName() == podName {
			nodeStart = pod.ObjectMeta.CreationTimestamp
			return
		}
	}
	err = fmt.Errorf("No pod in namespace %s found with name %s", namespace, podName)
	return
}

// Given a namespace and the name of a pod, return the name of its node
func GetPodNodeName(namespace, podName string) (nodeName string, err error) {
	pods, err := GetPods(namespace, podName)
	if err != nil {
		return
	}
	for _, pod := range pods {
		if pod.GetName() == podName {
			nodeName = pod.Spec.NodeName
			return
		}
	}
	err = fmt.Errorf("No pod in namespace %s found with name %s", namespace, podName)
	return
}

// Given a namespace, and an optional regex, return an array of Pod names
func GetPodNames(namespace string, params ...string) ([]string, error) {
	var names []string

	pods, err := GetPods(namespace, params...)
	if err != nil {
		return names, err
	}
	for _, pod := range pods {
		names = append(names, pod.GetName())
	}
	return names, err
}

// Given a namespace, and an optional regex, return an array of PVC Pod names
func GetPVCNames(namespace string, params ...string) ([]string, error) {
	var names []string

	clientset, err := GetClientset()
	if err != nil {
		return names, err
	}
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return names, err
	}
	for _, pvc := range pvcs.Items {
		if len(params) > 0 {
			match, _ := regexp.MatchString(params[0], pvc.GetName())
			if match {
				names = append(names, pvc.GetName())
			}
			continue
		} else {
			names = append(names, pvc.GetName())
		}
	}
	return names, err
}

// returns phase for pods
func GetPodStatus(namespace, podName string) (string, error) {
	var status string
	clientset, err := GetClientset()
	if err != nil {
		return status, err
	}

	for retries := 0; retries < MaxRetries; retries++ {
		pod, err := clientset.CoreV1().Pods(namespace).Get(
			context.TODO(),
			podName,
			v1.GetOptions{},
		)
		if err != nil {
			return status, err
		}

		status = string(pod.Status.Phase)
		if status != "Pending" {
			return status, nil
		}

		// Log and wait before retrying
		common.Infof("Pod %s in namespace %s is in 'Pending' state. Retrying in %d seconds...", podName, namespace, RetryIntervalSeconds)
		time.Sleep(time.Duration(RetryIntervalSeconds) * time.Second)
	}

	return status, fmt.Errorf("pod %s in namespace %s is still in 'Pending' state after %d retries", podName, namespace, MaxRetries)
}

// returns pod stats
func GetPodStats(namespace, podName string) (stats *PodStats, err error) {
	stats = new(PodStats)
	clientset, err := GetClientset()
	if err != nil {
		return
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(
		context.TODO(),
		podName,
		v1.GetOptions{},
	)
	if err != nil {
		return
	}
	stats.Phase = string(pod.Status.Phase)
	stats.ContainerStateWaitingReason = make(map[string]string)
	for _, cst := range pod.Status.ContainerStatuses {
		if cst.State.Waiting != nil {
			stats.ContainerStateWaitingReason[cst.Name] = cst.State.Waiting.Reason
		}
	}
	return
}

// given a pvc name, function returns its phase
func GetPVCStatus(namespace, pvcName string) (status string, err error) {
	clientset, err := GetClientset()
	if err != nil {
		return
	}
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(
		context.TODO(),
		pvcName,
		v1.GetOptions{},
	)
	if err != nil {
		return
	}
	status = string(pvc.Status.Phase)
	return
}

// given a pod's name and container, returns service logs
func GetPodLogs(namespace, podName string, containerName ...string) (string, error) {
	clientset, err := GetClientset()
	if err != nil {
		return "", err
	}
	var req *rest.Request
	if len(containerName) == 0 {
		req = clientset.CoreV1().Pods(namespace).GetLogs(
			podName,
			&coreV1.PodLogOptions{},
		)
	} else {
		req = clientset.CoreV1().Pods(namespace).GetLogs(
			podName,
			&coreV1.PodLogOptions{Container: containerName[0]},
		)
	}
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}
