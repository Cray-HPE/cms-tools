/*
 * get.go
 *
 * get objects command file
 *
 * Copyright 2019-2020, Cray Inc.
 */
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"strings"
)

// valid command arguments
var getValidCmds = []string{
	"bos",
	"k",
}

// getCmd command function
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "gets an object",
	Long: `performs a get on an object given a noun. For now the get cmd only operates 
on kubernetes objects. 
Example commands:

cmsdev get k token --print 
  # returns k8s access token 
cmsdev get k client-secret --print
  # returns k8s client secret`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a get flag argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		getCmdArgStr := strings.TrimSpace(args[0])
		validCmd := common.StringInArray(getCmdArgStr, getValidCmds)
		if validCmd == false {
			fmt.Printf("valid get commands are: %s\n", strings.Join(getValidCmds, " "))
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(kCmd)
	getCmd.AddCommand(bosCmd)
}
