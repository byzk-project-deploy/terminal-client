package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/desertbit/grumble"
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
			a.String("path", "要切换的路径", grumble.Default(user.HomeDir()))
		},
		Help: "变更当前系统路径",
		Run: func(c *grumble.Context) error {
			if err := os.Chdir(c.Args.String("path")); err != nil {
				return fmt.Errorf("变更当前目录失败: " + err.Error())
			}
			return nil
		},
	})

	current.AddCommand(&grumble.Command{
		Name: "call",
		Args: func(a *grumble.Args) {
			a.StringList("args", "系统命令与参数")
		},
		Help: "调用系统内部的命令",
		Run: func(c *grumble.Context) error {
			return utils.ExecSystemCmdWithBash(c.App, c.Args.StringList("args"))
		},
	})

}
