package utils

import (
	"errors"
	"fmt"
	transport_stream "github.com/go-base-lib/transport-stream"
	"os"
	"os/exec"
	"strings"
	"sync"

	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/config"
	"github.com/byzk-project-deploy/terminal-client/passwd"
	"github.com/byzk-project-deploy/terminal-client/remote"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/desertbit/grumble"
)

var (
	currentDial serverclientcommon.WriteStreamInterface

	lock = &sync.Mutex{}
)

func settingCurrentDial(w serverclientcommon.WriteStreamInterface) {
	lock.Lock()
	defer lock.Unlock()
	currentDial = w
}

func ClearCurrentDial() {
	lock.Lock()
	defer lock.Unlock()
	currentDial = nil
}

func ExecSystemCall(stream *transport_stream.Stream, serve *server.Info, cmdAndArgs []string) error {

	clientRand := passwd.Generator()
	systemCallData := &serverclientcommon.SystemCallOption{
		Name: serve.Name(),
		Rand: clientRand,
	}
	r, err := serverclientcommon.CmdSystemCall.ExchangeWithData(systemCallData, stream)
	if err != nil {
		return err
	}

	if err = r.UnmarshalJson(&systemCallData); err != nil {
		return fmt.Errorf("数据包解析失败: %s", err.Error())
	}

	i := strings.IndexByte(systemCallData.Rand, '_')
	userName := systemCallData.Rand[:i]
	serverKey := systemCallData.Rand + clientRand

	t, err := remote.New(userName, serverKey, systemCallData.Network, systemCallData.Addr)
	if err != nil {
		return fmt.Errorf("命令执行失败: %s", err.Error())
	}
	defer t.Close()

	return t.Run(strings.Join(cmdAndArgs, " "))
}

func ExecSystemCmdWithCurrentShell(app *grumble.App, cmdAndArgs []string) error {
	cmdAndArgs = append(config.Current().System.CallShellArgs, strings.Join(cmdAndArgs, " "))
	return ExecSystemCmdWithApp(app, config.Current().System.CallShellPath, cmdAndArgs...)
}

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
		return fmt.Errorf("执行系统命令[%s]失败: %s", cmd, err.Error())
	}
	return nil
}
