package utils

import (
	"errors"
	"fmt"
	"github.com/desertbit/grumble"
	"os"
	"os/exec"
)

func ExecSystemCmdWithApp(app *grumble.App, cmd string, args ...string) error {
	wdPath, err := os.Getwd()
	if err != nil {
		return errors.New("获取当前系统路径失败: " + err.Error())
	}
	return ExecSystemCmdWithAppAndWorkDir(app, wdPath, cmd, args...)
}

func ExecSystemCmdWithAppAndWorkDir(app *grumble.App, workDir, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Env = os.Environ()
	command.Stdout = app.Stdout()
	command.Stderr = app.Stderr()
	command.Stdin = os.Stdin
	command.Dir = workDir
	if err := command.Run(); err != nil {
		return errors.New(fmt.Sprintf("执行系统命令[%s]失败: %s", cmd, err.Error()))
	}
	return nil
}
