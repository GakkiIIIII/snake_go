package utils

import (
	"os"
	"os/exec"
	"runtime"
)

var clears map[string]func()

func Init() {
	clears = make(map[string]func())
	lm := func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}

	clears["linux"] = lm  // Linux
	clears["darwin"] = lm // Mac(arm64)

	clears["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") // Windows
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

func ClearTerminal() {
	clear, ok := clears[runtime.GOOS]
	if !ok {
		panic("ClearTerminal is not support your platform, bye...")
	}

	clear()
}
