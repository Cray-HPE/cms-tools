/*
Copyright 2019, Cray Inc.  All Rights Reserved.
Author: Torrey Cuthbert <tcuthbert@cray.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
   "fmt"
   "strings"
   "github.com/spf13/cobra"
   "os"
   "os/exec"
   "stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
   "stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
)

var debug bool
var validKObjects = []string{
	"client-secret",
	"token",
}

// kCmd represents the k command
var kCmd = &cobra.Command{
   Use:   "k",
   Short: "Kubernetes objects",
   Long: `References anything that are Kubernetes objects.   
k can be thought of as kubectl. 
Example commands:

cmsdev get k client-secret --print
  # prints the kubernetes admin-client secret
cmsdev get k token --print
  # prints the kubernetes access token`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("argument required, valid args: %v\n", validKObjects)
			return
		}
		kObject := strings.TrimSpace(args[0])
		display, _ := cmd.Flags().GetBool("print")
		set, _ := cmd.Flags().GetBool("set")
		if (!common.StringInArray(kObject, validKObjects)) {
			return
		}
		switch kObject {
		case "client-secret":
			clientSecret, err := k8s.GetOauthClientSecret()
	  		if err != nil {
				fmt.Printf("my err: %s\n", err.Error())
				return
	  		}
			if display == true {
	  			fmt.Println(clientSecret)
			}
		case "token":
			tokenStr, _ := k8s.GetAccessToken()
				// TODO: handle failure
			if display == true {
				fmt.Println(tokenStr)
			} else if set == true {
				// TODO: only print this in --debug or --verbose
				os.Setenv("TOKEN", tokenStr)
				cmdStr := fmt.Sprintf("TOKEN=%s", tokenStr)	
				_, err := exec.Command("bash", "-c", cmdStr).Output()
				if err != nil {
					fmt.Printf("failed to set token, %s\n", err.Error())
				}
			}
		}
   },
}

func init() {
   rootCmd.AddCommand(kCmd)
   kCmd.Flags().Bool("print", false, "print output")
   kCmd.Flags().Bool("set", false, "assign token to environment variable")
}
