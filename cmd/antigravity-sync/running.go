package main

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
)

var errAntigravityRunning = errors.New("Antigravity is running: close it first so its conversation databases aren't copied or replaced mid-write")

// antigravityRunning reports whether the Antigravity app is open.
// shortcut: macOS/Linux process names are unverified guesses; confirm them on those platforms.
func antigravityRunning() bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq Antigravity.exe", "/NH").Output()
		return err == nil && bytes.Contains(bytes.ToLower(out), []byte("antigravity.exe"))
	}
	return exec.Command("pgrep", "-x", "-i", "antigravity").Run() == nil
}
