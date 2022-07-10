package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/byzk-project-deploy/terminal-client/config"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/desertbit/grumble"
	"github.com/spf13/viper"
)

func initSystemCmd() {
	current.AddCommand(&grumble.Command{
		Name: "shell",
		Help: "查看与设置当前系统调用的shell环境",
		Args: func(a *grumble.Args) {
			a.String("shell", "要设置为当前shell的shell路径", grumble.Default(""))
			a.StringList("shellArgs", "shell的启动参数", grumble.Default([]string{}))
		},
		Run: func(c *grumble.Context) error {
			settingShellPath := c.Args.String("shell")
			if settingShellPath != "" {
				viper.Set("system.callShellPath", settingShellPath)
				viper.Set("system.callShellArgs", c.Args.StringList("shellArgs"))
				viper.WriteConfig()
				return nil
			}
			argsStr := strings.Join(config.Current().System.CallShellArgs, " ")
			c.App.Printf("%s %s\n", config.Current().System.CallShellPath, argsStr)
			return nil
		},
	})

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
		Name:    "call",
		Aliases: []string{"c"},
		Usage: `call cmd [flags...] [args...]
  c cmd [flags...] [args...]`,
		Args: func(a *grumble.Args) {
			a.StringList("args", "系统命令与参数")
		},
		Help: "调用系统内部的命令",
		Run: func(c *grumble.Context) error {
			return utils.ExecSystemCmdWithCurrentShell(c.App, c.Args.StringList("args"))
		},
	})

}
