//go:build !windows

package main

import "os/exec"

func hideHelperWindow(cmd *exec.Cmd){}
