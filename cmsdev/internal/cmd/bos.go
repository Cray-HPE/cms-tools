/*
 * bos.go
 * 
 * bos command file  
 *
 * Copyright 2019, Cray Inc.  All Rights Reserved.
 * Author: Torrey Cuthbert <tcuthbert@cray.com>
 */
package cmd

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/cms"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
)

// valid bos command arguments
var bosValidCmds = []string{
	"/",
	"endpoints",
	"logs",
	"session",
	"sessiontemplate",
	"v1",
	"version",
}

// variable used for parsing command line 
var argStr string

// bosCmd functions
var bosCmd = &cobra.Command{
	Use:   "bos",
	Short: "bos",
	Long: `bos command 
Example commands:

cmsdev get bos sessiontemplate --endpoint
  # describes bos's sessiontemplate endpoint
cmsdev get bos endpoints
  # returns all bos endpoints descriptions
cmsdev get bos logs 
  # returns bos container logs`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("args required, %s\n", strings.Join(bosValidCmds, " "))
			return		
		}
		argStr = strings.TrimSpace(args[0])
		ep, _ := cmd.Flags().GetBool("endpoint")
		if argStr == "endpoints" {
			common.PrintEndpoints("bos")
		} else if ep == true {
			common.PrintEndpoints("bos", argStr)
		} else if argStr == "logs" {
			if err := cms.PrintServiceLogs("bos"); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(bosCmd)
	bosCmd.Flags().Bool("endpoint", false, "display single endpoint")
}
