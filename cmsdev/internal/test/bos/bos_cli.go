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
package bos

/*
 * bos_cli.go
 *
 * bos CLI helpers
 *
 */

import (
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/common"
	"stash.us.cray.com/SCMS/cms-tools/cmsdev/internal/lib/test"
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
