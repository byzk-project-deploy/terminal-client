package cmd

import (
	"errors"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/desertbit/grumble"
	"os"
)

func initFileSystemCmd() {
	current.AddCommand(&grumble.Command{
		Name: "pwd",
		Help: "查看当前所在路径",
		Run: func(c *grumble.Context) error {
			wdPath, err := os.Getwd()
			if err != nil {
				return errors.New("获取当前路径失败: " + err.Error())
			}
			_, _ = c.App.Println(wdPath)
			return nil
		},
	})

	current.AddCommand(&grumble.Command{
		Name: "cd",
		Args: func(a *grumble.Args) {
			a.String("path", "要切换的路径")
		},
		Help: "调用系统内置的cd命令, 变更当前系统路径",
		Run: func(c *grumble.Context) error {
			return utils.ExecSystemCmdWithApp(c.App, "cd.exe", c.Args.String("path"))
		},
	})

}
