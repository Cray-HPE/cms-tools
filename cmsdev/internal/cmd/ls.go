/*
 * ls.go
 *
 * ls command file
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
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/cms"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"strings"
)

// ls supported objects
var lsObjects = []string{
	"services",
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list",
	Long: `performs a list of a given object
Example commands:

cmsdev ls services
  # returns a list of currently installed cms services 
cmsdev ls bos --name 
  # returns the service pod name of bos
cmsdev ls services --status
  # returns a list of cms services pods with current phase
cmsdev ls services --count
  # returns the number of currently installed cms services.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("argument required, current supported objects: %v\n", lsObjects)
			return
		}
		obj := strings.TrimSpace(args[0])
		serviceNames, numServices := cms.GetCMSServiceNames()
		validObjects := append(lsObjects, serviceNames...)
		validObj := common.StringInArray(obj, validObjects)
		if validObj == false {
			fmt.Printf("valid ls objects are %s\n", strings.Join(validObjects, " "))
			return
		}
		status, _ := cmd.Flags().GetBool("status")
		name, _ := cmd.Flags().GetBool("name")
		count, _ := cmd.Flags().GetBool("count")
		switch obj {
		case "services":
			if count {
				_, numServices := cms.GetCMSServiceNames()
				fmt.Println(numServices)
				return
			}
			if status {
				// TODO: returns only a pod's phase
				cms.ListServices()
				return
			}
			if numServices != 0 {
				fmt.Printf("%s\n", strings.Join(serviceNames, " "))
			}
		default:
			if name {
				fmt.Println(cms.GetCMSServiceName(obj))
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().Bool("status", false, "list service status, default")
	lsCmd.Flags().Bool("count", false, "list cms service count")
	lsCmd.Flags().Bool("name", false, "service name")
}
