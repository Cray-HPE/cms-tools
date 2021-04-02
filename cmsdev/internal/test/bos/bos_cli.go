package bos

/*
 * bos_cli.go
 *
 * bos CLI helpers
 *
 * Copyright 2021 Hewlett Packard Enterprise Development LP
 */

import (
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/cms-tools/cmsdev/internal/lib/test"
	"strconv"
)

func runCLICommand(vnum int, cmdArgs ...string) []byte {
	if vnum == 0 {
		return test.RunCLICommandJSON("bos", cmdArgs...)
	} else if vnum > 0 {
		cliCmdArgs := append([]string{"v" + strconv.Itoa(vnum)}, cmdArgs...)
		return test.RunCLICommandJSON("bos", cliCmdArgs...)
	} else {
		common.Errorf("PROGRAMMING LOGIC ERROR: sessionTestCLI: Negative vnum value (%d)", vnum)
		return nil
	}
}
