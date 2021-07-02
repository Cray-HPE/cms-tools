/*
 * get.go
 *
 * get objects command file
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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
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
