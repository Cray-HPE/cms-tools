package k8s

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	coreV1 "k8s.io/api/core/v1"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
)

// struct to hold current pod status
type PodStats struct {
	Phase string
	ContainerStateWaitingReason map[string]string
}

// struct to hold container environment variables
type ContainerEnvVar struct {
	Name, Value string
}

func getKubeConfig() (*rest.Config, error) {
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	//TODO: check if KUBECONFIG is set
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
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

func GetOauthClientSecret() (string, error) {
	// TODO: clean this up. remove exec.Command and replace with clientgo
    var cmdOut []byte
    var err error
    var cmd *exec.Cmd

	common.Infof("Trying to look up path of kubectl")
    path, err := exec.LookPath("kubectl")
    if err != nil {
		return "", err
    }
	cmd = exec.Command(path, "get", "secrets", "admin-client-auth", "-ojsonpath='{.data.client-secret}'")
	common.Infof("Running command: %s", cmd)
    cmdOut, err = cmd.CombinedOutput()
    if err != nil {
		if len(cmdOut) > 0 {
			common.Infof("Command output: %s", cmdOut)
		}
        return "", err
    }
	cmdStr := fmt.Sprintf("echo %s | base64 -d", strings.Trim(string(cmdOut), "'"))
	common.Infof("Running command: %s", cmdStr)
    cmdOut, err = exec.Command("bash", "-c", cmdStr).Output()
    if err != nil {
		if len(cmdOut) > 0 {
			common.Infof("Command output: %s", cmdOut)
		}
        return "", err
    }
    return string(cmdOut), nil
}

func GetAccessJSON(params ...string) ([]byte, error) {
	// TODO: clean this up. remove curl and replace with clientgo or a rest call
    var cmdOut []byte
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

	cmdStr := fmt.Sprintf("curl -k -s -d grant_type=client_credentials -d client_id=admin-client -d client_secret=%s " +
				 "https://api-gw-service-nmn.local/keycloak/realms/shasta/protocol/openid-connect/token", clientSecret)
	common.Infof("Running command: bash -c \"%s\"", cmdStr)
	cmdOut, err = exec.Command("bash", "-c", cmdStr).Output()
    if err != nil {
		if len(cmdOut) > 0 {
			common.Infof("Command output: %s", cmdOut)
		}
        return nil, err
    } else {
		return cmdOut, nil
	}
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
	common.Infof("Parsing JSON object containing access token")
	if err = json.Unmarshal(cmdOut, &jsonData); err != nil {
		return "", err
	}
    return jsonData["access_token"].(string), nil
}

// Given a container, return a map of its environment variables
func GetEnvVars(container coreV1.Container) (envVars []ContainerEnvVar) {
	envVars = make([]ContainerEnvVar, 0, len(container.Env))
	for _, evar := range container.Env {
		envVars = append(envVars, ContainerEnvVar{ Name: evar.Name, Value: evar.Value })
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

// Given a namespace, and an optional regex, return an array of Pods (whose name match the regex, if specified)
func GetPods(namespace string, params ...string) ([]coreV1.Pod, error) {
	var pods []coreV1.Pod

	clientset, err := GetClientset()
	if err != nil {
		return pods, err
	}
	allPods, err := clientset.CoreV1().Pods(namespace).List(v1.ListOptions{})
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
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(v1.ListOptions{})
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
	pod, err := clientset.CoreV1().Pods(namespace).Get(
		podName, 
		v1.GetOptions{},
	)
	if err != nil {
		return status, err
	}
	return string(pod.Status.Phase), err
}

// returns pod stats
func GetPodStats(namespace, podName string) (stats *PodStats, err error) {
	stats = new(PodStats)
	clientset, err := GetClientset()
	if err != nil {
		return
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(
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
	podLogs, err := req.Stream()
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
