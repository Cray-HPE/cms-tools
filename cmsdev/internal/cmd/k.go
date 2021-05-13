/*
  Copyright 2019-2021 Hewlett Packard Enterprise Development LP

  Permission is hereby granted, free of charge, to any person obtaining a
  copy of this software and associated documentation files (the "Software"),
  to deal in the Software without restriction, including without limitation
  the rights to use, copy, modify, merge, publish, distribute, sublicense,
  and/or sell copies of the Software, and to permit persons to whom the
  Software is furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included
  in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
  OTHER DEALINGS IN THE SOFTWARE.

  (MIT License)
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/k8s"
	"strings"
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
		if !common.StringInArray(kObject, validKObjects) {
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
