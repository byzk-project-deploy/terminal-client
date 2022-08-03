package cmd

import (
	"fmt"
	"github.com/byzk-project-deploy/grumble"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/byzk-project-deploy/terminal-client/loading"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/byzk-project-deploy/terminal-client/utils"
	transport_stream "github.com/go-base-lib/transport-stream"
	"golang.org/x/exp/slices"
	"strings"
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

			havePrefix := prefix != ""

			result := make([]string, 0, 18)
			if len(args) > 0 {
				return result
			}

			currentModelInfo := cmdmodel.CurrentModelInfo()

			_ = currentModelInfo.RangeWithContextWrapper(nil, func(name string, stream *transport_stream.Stream, info *cmdmodel.ServerConn) error {
				r, err := serverclientcommon.CmdSystemShellList.Exchange(stream)
				if err != nil {
					return nil
				}

				var res []string
				if err = r.UnmarshalJson(&res); err != nil {
					return nil
				}

				for i := range res {
					s := res[i]
					if havePrefix {
						if !strings.HasPrefix(s, prefix) {
							continue
						}
					}

					if slices.Contains(result, s) {
						continue
					}

					result = append(result, s)
				}
				return nil
			})

			if !havePrefix {
				result = append(result, "-h", "--help")
			}

			return result
		},
		Run: cmdErrRunWrapper(func(c *ContextWrapper) error {
			currentModel := cmdmodel.CurrentModelInfo()
			l := currentModel.Len()
			return currentModel.RangeWithContextWrapper(c, func(name string, stream *transport_stream.Stream, info *cmdmodel.ServerConn) error {
				settingShellPath := c.Args.String("shell")
				if settingShellPath != "" {
					if _, err := serverclientcommon.CmdSystemShellCurrentSetting.ExchangeWithData(&serverclientcommon.ShellSettingOption{
						Name: settingShellPath,
						Args: c.Args.StringList("shellArgs"),
					}, stream); err != nil {
						c.PrintError(fmt.Errorf("服务器[%s]发生错误: %s", name, err.Error()))
					}
					return nil
				}

				r, err := serverclientcommon.CmdSystemShellCurrent.Exchange(stream)
				if err != nil {
					c.PrintError(err)
					return nil
				}

				res := strings.TrimSpace(string(r))

				if l > 1 {
					fmt.Printf("服务[%s]响应:\n%s\n\n", name, res)
				} else {
					fmt.Println(res)
				}

				return nil
			})
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
			p := strings.TrimSpace(c.Args.String("path"))
			if p == "" {
				p = user.HomeDir()
			}
			p = server.JoinPath(p)
			if p == server.CurrentPath() {
				return nil
			}
			c.SuccessCallback(func() error {
				server.CurrentPathChange(p)
				return nil
			})
			s := loading.Loading("正在加载服务列表")
			defer s.Stop()

			currentModel := cmdmodel.CurrentModelInfo()

			return currentModel.RangeWithContextWrapper(c, func(name string, stream *transport_stream.Stream, info *cmdmodel.ServerConn) error {
				s.UpdateSuffix(fmt.Sprintf("正在[%s]上执行", name))
				_, err := serverclientcommon.CmdSystemDirPath.ExchangeWithData(p, stream)
				if err != nil {
					c.PrintError(fmt.Errorf("获取服务器[%s]上的地址[%s]失败: %s", name, p, err.Error()))
				}
				return nil
			})
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
			currentModel := cmdmodel.CurrentModelInfo()
			return currentModel.RangeWithContextWrapper(c, func(name string, stream *transport_stream.Stream, info *cmdmodel.ServerConn) error {
				if err := utils.ExecSystemCall(stream, info.Info, c.Args.StringList("args")); err != nil {
					c.PrintError(fmt.Errorf("在服务器[%s]上执行命令失败, 错误原因: %s", name, err.Error()))
				}
				return nil
			})
		}),
	}
)

func noFindCommandHandle(app *grumble.App, args []string) error {
	args = append([]string{"c"}, args...)
	return app.RunCommand(args)
}
