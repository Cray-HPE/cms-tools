/*
 * ls.go
 *
 * ls command file
 *
 * Copyright 2019-2020, Cray Inc.
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
