/*
 * version.go
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// This constant is updated by a git pre-commit hook with the first
// line of the .rpm_version_cray-cmstools-crayctldeploy file
const cmsdevVersion = "NONE"

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
