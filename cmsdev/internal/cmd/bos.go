/*
 * bos.go
 *
 * bos command file
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
