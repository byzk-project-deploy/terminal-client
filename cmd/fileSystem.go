package cmd

import (
	"fmt"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/desertbit/grumble"
)

var (
	// shellCommand shell命令
	shellCommand = &grumble.Command{
		Name: "shell",
		Help: "查看与设置当前系统调用的shell环境",
		Args: func(a *grumble.Args) {
			a.String("shell", "要设置为当前shell的shell路径", grumble.Default(""))
			a.StringList("shellArgs", "shell的启动参数", grumble.Default([]string{}))
		},
		Completer: func(prefix string, args []string) []string {
			//	var (
			//		r *serverclientcommon.Result
			//	)
			//	result := make([]string, 0, 18)
			//	if len(args) > 0 {
			//		return result
			//	}
			//	s, err := server.OpenUnixServer()
			//	if err != nil {
			//		goto End
			//	}
			//	r, err = serverclientcommon.CmdSystemShellList.Exchange(s.ReadWriter)
			//	if err == nil {
			//		_ = r.Data.Unmarshal(&result)
			//	}
			//	if prefix != "" {
			//		for i := len(result) - 1; i >= 0; i-- {
			//			if !strings.HasPrefix(result[i], prefix) {
			//				result = append(result[:i], result[i+1:]...)
			//			}
			//		}
			//	}
			//End:
			//	if prefix == "" {
			//		result = append(result, "-h", "--help")
			//	}
			//	return result
			return nil
		},
		Run: cmdErrRunWrapper(func(c *ContextWrapper) error {
			info := server.NewUnixServerInfo()
			stream, err := info.ConnToStream()
			if err != nil {
				panic(err)
			}
			data, err := serverclientcommon.CmdHello.ExchangeWithData("Hello", stream)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data))
			// var res string
			// serverList := CurrentServerList()
			// l := serverList.Len()
			// return serverList.RangeWWithContextWrapper(c, func(name string, info *cmdmodel.ServerResult) error {

			// 	settingShellPath := c.Args.String("shell")
			// 	if settingShellPath != "" {
			// 		if _, err := serverclientcommon.CmdSystemShellCurrentSetting.ExchangeWithData(&serverclientcommon.ShellSettingOption{
			// 			Name: settingShellPath,
			// 			Args: c.Args.StringList("shellArgs"),
			// 		}, info.ReadWriter); err != nil {
			// 			c.PrintError(fmt.Errorf("服务器[%s]发生错误: %s", name, err.Error()))
			// 		}
			// 		return nil
			// 	}

			// 	r, err := serverclientcommon.CmdSystemShellCurrent.Exchange(info.ReadWriter)
			// 	if err != nil {
			// 		c.PrintError(err)
			// 		return nil
			// 	}

			// 	if err = r.Data.Unmarshal(&res); err != nil {
			// 		c.PrintError(err)
			// 		return nil
			// 	}

			// 	res = strings.TrimSpace(res)

			// 	if l > 0 {
			// 		fmt.Printf("服务[%s]响应:\n%s\n\n", name, res)
			// 	} else {
			// 		fmt.Println(res)
			// 	}

			// 	return nil

			// })
			return nil
			// func(c *grumble.Context) error {
			// 	serverList := CurrentServerList()
			// 	l := serverList.Len()
			// 	serverList.RangeWWithContextWrapper()
			// 	settingShellPath := c.Args.String("shell")
			// 	if settingShellPath != "" {
			// 		viper.Set("system.callShellPath", settingShellPath)
			// 		viper.Set("system.callShellArgs", c.Args.StringList("shellArgs"))
			// 		viper.WriteConfig()
			// 		return nil
			// 	}
			// 	serverclientcommon.CmdSystemShellCurrent.Exchange()
			// 	argsStr := strings.Join(config.Current().System.CallShellArgs, " ")
			// 	c.App.Printf("%s %s\n", config.Current().System.CallShellPath, argsStr)
			// 	return nil
			// }
		}),
	}

	// pwdCommand pwd命令
	pwdCommand = &grumble.Command{
		Name: "pwd",
		Help: "查看当前所在路径",
		Run: func(c *grumble.Context) error {
			_, _ = c.App.Println(server.CurrentPath())
			return nil
		},
	}

	// cdCommand cd命令
	cdCommand = &grumble.Command{
		Name: "cd",
		Args: func(a *grumble.Args) {
			a.String("path", "要切换的路径", grumble.Default(user.HomeDir()))
		},
		Help: "变更系统中当前的路径, 如果目录不存在则报错",
		Run: cmdErrRunWrapper(func(c *ContextWrapper) error {
			// var ok *bool
			// p := strings.TrimSpace(c.Args.String("path"))
			// if p == "" {
			// 	p = user.HomeDir()
			// }
			// p = server.JoinPath(p)
			// if p == server.CurrentPath() {
			// 	return nil
			// }
			// c.SuccessCallback(func() error {
			// 	server.CurrentPathChange(p)
			// 	return nil
			// })
			// s := loading.Loading("正在加载服务列表")
			// defer s.Stop()

			// serverList := CurrentServerList()

			// serverList.RangeWWithContextWrapper(c, func(name string, info *cmdmodel.ServerResult) error {
			// 	s.UpdateSuffix(fmt.Sprintf("正在[%s]上执行", name))
			// 	r, err := serverclientcommon.CmdSystemDirPath.ExchangeWithData(p, info.ReadWriter)
			// 	if err != nil {
			// 		c.PrintError(fmt.Errorf("获取服务器[%s]上的地址[%s]失败: %s", name, p, err.Error()))
			// 		return nil
			// 	}

			// 	if err = r.Data.Unmarshal(&ok); err != nil || !*ok {
			// 		c.PrintError(fmt.Errorf("服务器[%s]地址[%s]不存在", name, p))
			// 	}
			// 	return nil
			// })

			return nil
		}),
	}

	// 系统调用命令
	callCommand = &grumble.Command{
		Name:    "call",
		Aliases: []string{"c"},
		Usage: `call cmd [flags...] [args...]
  c cmd [flags...] [args...]`,
		Args: func(a *grumble.Args) {
			a.StringList("args", "系统命令与参数")
		},
		Help: "调用系统内部的命令",
		Run: cmdErrRunWrapper(func(c *ContextWrapper) error {
			return nil
			// serverList := CurrentServerList()
			// return serverList.RangeWWithContextWrapper(c, func(name string, info *cmdmodel.ServerResult) error {
			// 	utils.ExecSystemCall(c.App, info.ServerInfo, c.Args.StringList("args"))
			// 	return nil
			// })
		}),
	}
)

func noFindCommandHandle(app *grumble.App, args []string) error {
	args = append([]string{"c"}, args...)
	return app.RunCommand(args)
}
