//
//  MIT License
//
//  (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
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
/*
 * version.go
 */
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// The placeholder version string will be replaced at repo build time by
// the runBuildPrep.sh script
const cmsdevVersion = "@VERSION@"

// versionCmd command functions
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long: `version displays version of cmsdev tool
Example Commands:

cmsdev version
  # displays version of cmsdev command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", cmsdevVersion)
		return
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
