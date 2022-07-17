package utils

import (
	"errors"
	"fmt"
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

func WriteRuneToCurrentDial(r rune) bool {
	lock.Lock()
	defer lock.Unlock()
	if currentDial == nil {
		return false
	}
	serverclientcommon.SuccessResult(r).WriteTo(currentDial)
	return true
}

func ExecSystemCall(app *grumble.App, addressName string, cmdAndArgs []string) error {

	// if addressName == "unix" {
	// 	return ExecSystemCmdWithCurrentShell(app, cmdAndArgs)
	// }

	serve, err := server.OpenServer(addressName)
	if err != nil {
		return fmt.Errorf("与远程服务器建立连接失败: %s", err.Error())
	}

	clientRand := passwd.Generator()
	systemCallData := &serverclientcommon.SystemCallOption{
		Rand: clientRand,
	}
	r, err := serverclientcommon.CmdSystemCall.ExchangeWithData(systemCallData, serve)
	if err != nil {
		return err
	}
	if r.Error {
		return fmt.Errorf("%s: %s", r.Code, r.Msg)
	}
	if err = r.Data.Unmarshal(&systemCallData); err != nil {
		return fmt.Errorf("数据包解析失败: %s", err.Error())
	}

	i := strings.IndexByte(systemCallData.Rand, '_')
	userName := systemCallData.Rand[:i]
	serverKey := systemCallData.Rand + clientRand

	t, err := remote.New(userName, serverKey, "127.0.0.1"+systemCallData.Addr)
	if err != nil {
		return fmt.Errorf("命令执行失败: %s", err.Error())
	}
	defer t.Close()

	// t.Run("/usr/bin/zsh -i -c \"" + strings.Join(cmdAndArgs, " ") + "\"")
	t.Run("sh  -i -c \"" + strings.Join(cmdAndArgs, " ") + "\"")

	// // 创建 ssh 配置
	// sshConfig := &ssh.ClientConfig{
	// 	User: userName,
	// 	Auth: []ssh.AuthMethod{
	// 		ssh.Password(serverKey),
	// 	},
	// 	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	// 	Timeout:         5 * time.Second,
	// }

	// if systemCallData.Addr[0] != ':' {
	// 	systemCallData.Addr = ":" + systemCallData.Addr
	// }
	// // 创建 client
	// // client, err := ssh.Dial("tcp", addressName+systemCallData.Addr, sshConfig)
	// client, err := ssh.Dial("tcp", "127.0.0.1"+systemCallData.Addr, sshConfig)
	// if err != nil {
	// 	return fmt.Errorf("创建命令转发通道失败: %s", err.Error())
	// }
	// defer client.Close()

	// session, err := client.NewSession()
	// if err != nil {
	// 	return fmt.Errorf("创建会话失败: %s", err.Error())
	// }
	// defer session.Close()

	// fd := int(os.Stdin.Fd())
	// // make raw
	// state, err := terminal.MakeRaw(fd)
	// if err != nil {
	// 	return fmt.Errorf("转换原始终端失败: %s", err.Error())
	// }
	// defer terminal.Restore(fd, state)

	return nil
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
