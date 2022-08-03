package cmdmodel

import (
	"fmt"
	"github.com/byzk-project-deploy/terminal-client/server"
	"os"

	"github.com/byzk-project-deploy/grumble"
	"github.com/byzk-project-deploy/terminal-client/stdio"
	"github.com/fatih/color"
)

var (
	currentApp       *grumble.App
	currentModelInfo *ModelInfo

	modelMap = make(map[ModelName]*ModelInfo)
)

// CurrentModelInfo 获取当前模块信息
func CurrentModelInfo() *ModelInfo {
	return currentModelInfo
}

// Registry 注册模式
func Registry(name ModelName, info *ModelInfo) {
	info.Name = name
	if info.ServerList == nil {
		info.ServerList = newServerList()
	}
	info.ServerList.Add("unix", &ServerConn{
		Info: server.NewUnixServerInfo(),
	})
	modelMap[name] = info
}

func InitApp(currentModelName ModelName, historyFile string) {
	if info, ok := modelMap[currentModelName]; !ok {
		fmt.Println("缺失的终端模式信息...")
		os.Exit(1)
	} else {
		currentModelInfo = info
	}

	currentApp = grumble.New(&grumble.Config{
		Name:                  "bypt",
		HistoryFile:           historyFile,
		PromptColor:           color.New(color.FgGreen, color.Bold),
		HelpHeadlineColor:     color.New(color.FgGreen),
		HelpHeadlineUnderline: true,
		HelpSubCommands:       true,
		Stdin:                 stdio.Stdin,
	})

	currentApp.SetInterruptHandler(func(a *grumble.App, count int) {
	})
	currentApp.OnClose(func() error {
		return nil
	})
	currentApp.OnClosing(func() error {
		return nil
	})
	currentApp.SetPrintASCIILogo(func(a *grumble.App) {
		_, _ = a.Println(` _`)
		_, _ = a.Println(`| |                 _`)
		_, _ = a.Println(`| |__  _   _ ____ _| |_`)
		_, _ = a.Println(`|  _ \| | | |  _ (_   _)`)
		_, _ = a.Println(`| |_) ) |_| | |_| || |_`)
		_, _ = a.Println(`|____/ \__  |  __/  \__)`)
		_, _ = a.Println(`      (____/|_|`)
		_, _ = a.Println()
		_, _ = a.Println("             版本: 3.0.0")
		_, _ = a.Println("             作者: 无&痕")
		_, _ = a.Println()
		_, _ = a.Println("应用部署管理平台终端客户端一切只为便捷、高效与可靠的部署和管理应用^-^")
		_, _ = a.Println()
		_, _ = a.Println()
	})

	currentModelInfo.Convert()

	if err := currentApp.Run(); err != nil {
		return
	}

	os.Exit(0)
}
