//go:build unix

package main

import (
	"os"
	"syscall"
)

func getFileGID(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Gid)
	}
	return 0
}
