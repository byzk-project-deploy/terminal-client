package cmd

import (
	"github.com/desertbit/grumble"
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

var (
	historyFile = filepath.Join(os.TempDir(), "bypt.hist")

	pwdPath, _ = os.Getwd()

	current *grumble.App
)

func initCmd() {
	initHistoryCmd()
	initFileSystemCmd()
}

func Run() {

	current = grumble.New(&grumble.Config{
		Name:                  "bypt",
		Prompt:                "bypt » ",
		HistoryFile:           historyFile,
		PromptColor:           color.New(color.FgGreen, color.Bold),
		HelpHeadlineColor:     color.New(color.FgGreen),
		HelpHeadlineUnderline: true,
		HelpSubCommands:       true,
	})

	initCmd()

	current.SetInterruptHandler(func(a *grumble.App, count int) {
	})
	current.OnClose(func() error {
		return nil
	})
	current.OnClosing(func() error {
		return nil
	})
	current.SetPrintASCIILogo(func(a *grumble.App) {
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

	if err := current.Run(); err != nil {
		return
	}

	os.Exit(1)
}
