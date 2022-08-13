package cmd

import (
	"errors"
	"fmt"
	"github.com/byzk-project-deploy/grumble"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/fatih/color"
	transportstream "github.com/go-base-lib/transport-stream"
	"github.com/gosuri/uitable"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
)

var (
	// remoteServerConvertCmd 插件模式转换命令
	remoteServerConvertCmd = &grumble.Command{
		Name: "server",
		Help: "切换为远程服务集群模式",
		Run: func(c *grumble.Context) error {
			return cmdmodel.ModelServer.Convert()
		},
	}

	// remoteServerJoinCmd 远程服务器添加
	remoteServerJoinCmd = &grumble.Command{
		Name: "join",
		Help: "添加服务器",
		LongHelp: `添加服务器到当前服务器的控制节点中
如果对端服务器不存在bypt程序将会提示自动安装, 需要给予root权限, 否则将会安装失败
`,
		Args: func(a *grumble.Args) {
			a.StringList("serverIp", "要添加的服务器ip", grumble.Min(1))
		},
		Flags: func(f *grumble.Flags) {
			f.String("u", "username", "", "服务器用户名")
			f.Uint("p", "sshPort", 22, "服务器ssh端口")
		},
		Run: func(c *grumble.Context) error {
			serverIpList := c.Args.StringList("serverIp")
			if len(serverIpList) == 0 {
				return fmt.Errorf("缺失要添加的服务器IP列表")
			}

			sshPort := c.Flags.Uint("sshPort")
			if sshPort == 0 {
				sshPort = 22
			}

			sshPortBytes, err := transportstream.IntToBytes[uint16](uint16(sshPort))
			if err != nil {
				return fmt.Errorf("转换ssh端口失败: %s", err.Error())
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			username := c.Flags.String("username")
			_, err = serverclientcommon.CmdRemoteServerAdd.ExchangeWithOption(stream, &serverclientcommon.ExchangeOption{
				StreamHandle: func(exchangeData serverclientcommon.ExchangeData, stream *transportstream.Stream) (res serverclientcommon.ExchangeData, err error) {
					defer func() {
						if err != nil && err != transportstream.StreamIsEnd {
							err = nil
						}
					}()
					ipStr := string(exchangeData)
					color.Yellow("正在添加服务器[%s]", ipStr)
					serverUsername := username
					if serverUsername == "" {
						if serverUsername, err = c.ShellTools.Prompt(utils.PromptNotEmptyVerify(fmt.Sprintf("%s用户名", ipStr))); err != nil {
							c.App.PrintError(errors.New("本次操作已被中断"))
							return nil, transportstream.StreamIsEnd
						}
					}

					var serverPassword string
					if serverPassword, err = c.ShellTools.Prompt(utils.PromptPassword("用户密码")); err != nil {
						c.App.PrintError(errors.New("本次操作已被中断"))
						return nil, transportstream.StreamIsEnd
					}

					if err = stream.WriteMsg(sshPortBytes, transportstream.MsgFlagSuccess); err != nil {
						c.App.PrintError(fmt.Errorf("添加服务器[%s]失败: %s", ipStr, err.Error()))
						return
					}

					if err = stream.WriteMsg([]byte(serverUsername), transportstream.MsgFlagSuccess); err != nil {
						c.App.PrintError(fmt.Errorf("添加服务器[%s]失败: %s", ipStr, err.Error()))
						return
					}

					if err = stream.WriteMsg([]byte(serverPassword), transportstream.MsgFlagSuccess); err != nil {
						c.App.PrintError(fmt.Errorf("添加服务器[%s]失败: %s", ipStr, err.Error()))
						return
					}

					if _, err = stream.ReceiveMsg(); err != nil {
						c.App.PrintError(fmt.Errorf("添加服务器[%s]失败: %s", ipStr, err.Error()))
						return
					}

					_, _ = c.App.Println()
					return nil, nil
				},
				Data: serverIpList,
			})
			return err
		},
	}

	// remoteServerListCmd 远程服务器列表命令
	remoteServerListCmd = &grumble.Command{
		Name:      "ls",
		Help:      "查询当前服务器列表信息",
		Completer: remoteServerNameListByPrefixAndExclude,
		Args: func(a *grumble.Args) {
			a.StringList("searchKeywords", "关键字列表, 查询带关键字的服务器IP以及别名", grumble.Min(0))
		},
		Run: func(c *grumble.Context) error {
			searchServerList := c.Args.StringList("searchKeywords")

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			serverListBytes, err := serverclientcommon.CmdRemoteServerList.ExchangeWithData(searchServerList, stream)
			if err != nil {
				return err
			}

			var serverList []*serverclientcommon.ServerInfo
			if err = serverListBytes.UnmarshalJson(&serverList); err != nil {
				return fmt.Errorf("反序列化服务器数据失败: %s", err.Error())
			}

			table := uitable.New()
			table.MaxColWidth = 30
			table.AddRow("IP", "端口", "别名", "ssh用户", "加入时间", "状态", "消息")

			for i := range serverList {
				serverInfo := serverList[i]

				port := ""
				if serverInfo.Port != 0 {
					port = strconv.Itoa(serverInfo.Port)
				}

				statusText := ""
				switch serverInfo.Status {
				case serverclientcommon.ServerStatusNoCheck:
					statusText = color.YellowString("未检测")
					if serverInfo.EndMsg == "" {
						serverInfo.EndMsg = "服务器未检测是否可用"
					}
				case serverclientcommon.ServerStatusCheckErr:
					statusText = color.RedString("可用性检测失败")
				case serverclientcommon.ServerStatusUserErr:
					statusText = color.RedString("SSH认证失败")
				case serverclientcommon.ServerStatusNeedInstall:
					statusText = color.RedString("未安装主程序")
				case serverclientcommon.ServerStatusNetworkErr:
					statusText = color.RedString("网络错误")
				case serverclientcommon.ServerStatusNoRun:
					statusText = color.YellowString("未启动")
				case serverclientcommon.ServerRunning:
					statusText = color.GreenString("已启动")
				}

				table.AddRow(serverInfo.Id, port, strings.Join(serverInfo.Alias, ","), serverInfo.SSHUser, utils.TimeFormat(serverInfo.JoinTime), statusText, serverInfo.EndMsg)
			}

			_, _ = c.App.Println(table)
			_, _ = c.App.Println()

			return nil
		},
	}

	// remoteServerInfoCmd 查看服务器详细信息
	remoteServerInfoCmd = &grumble.Command{
		Name: "info",
		Help: "查看服务器详细内容",
		Completer: func(prefix string, args []string) []string {
			if len(args) > 0 {
				return nil
			}
			return remoteServerNameListByPrefixAndExclude(prefix, nil)
		},
		Args: func(a *grumble.Args) {
			a.String("server", "服务器IP或别名")
		},
		Run: func(c *grumble.Context) error {
			serverFlag := strings.TrimSpace(c.Args.String("server"))
			if serverFlag == "" {
				return fmt.Errorf("服务器IP或别名不能为空")
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			serverInfoBytes, err := serverclientcommon.CmdRemoteServerInfo.ExchangeWithData(serverFlag, stream)
			if err != nil {
				return err
			}

			var serverInfo *serverclientcommon.ServerInfo
			if err = serverInfoBytes.UnmarshalJson(&serverInfo); err != nil {
				return fmt.Errorf("反学列话服务器数据失败: %s", err.Error())
			}

			statusText := ""
			switch serverInfo.Status {
			case serverclientcommon.ServerStatusNoCheck:
				statusText = color.YellowString("未检测")
				if serverInfo.EndMsg == "" {
					serverInfo.EndMsg = "服务器未检测是否可用"
				}
			case serverclientcommon.ServerStatusCheckErr:
				statusText = color.RedString("可用性检测失败")
			case serverclientcommon.ServerStatusUserErr:
				statusText = color.RedString("SSH认证失败")
			case serverclientcommon.ServerStatusNeedInstall:
				statusText = color.RedString("未安装主程序")
			case serverclientcommon.ServerStatusNetworkErr:
				statusText = color.RedString("网络错误")
			case serverclientcommon.ServerStatusNoRun:
				statusText = color.YellowString("未启动")
			case serverclientcommon.ServerRunning:
				statusText = color.GreenString("已启动")
			}

			portStr := ""
			if serverInfo.Port != 0 {
				portStr = strconv.Itoa(serverInfo.Port)
			}

			table := uitable.New()
			table.AddRow("IP地址", serverInfo.IP)
			table.AddRow("远程端口", portStr)
			table.AddRow("服务别名", strings.Join(serverInfo.Alias, ","))
			table.AddRow("SSH用户", serverInfo.SSHUser)
			table.AddRow("SSH密码", serverInfo.SSHPassword)
			table.AddRow("加入时间", utils.TimeFormat(serverInfo.JoinTime))
			table.AddRow("当前状态", statusText)
			table.AddRow("最后消息", color.RedString(serverInfo.EndMsg))

			_, _ = c.App.Println(table)
			_, _ = c.App.Println()

			return nil
		},
	}

	// remoteServerPasswordUpdateCmd 远程服务密码更改命令
	remoteServerPasswordUpdateCmd = &grumble.Command{
		Name:      "password",
		Help:      "单个或批量修改远程服务的SSH用户密码",
		Completer: remoteServerNameListByPrefixAndExclude,
		Args: func(a *grumble.Args) {
			a.StringList("serverNameList", "要修改密码的服务器IP或别名列表", grumble.Min(1))
		},
		Run: func(c *grumble.Context) error {
			serverNameList := c.Args.StringList("serverNameList")

			password, err := c.ShellTools.Prompt(utils.PromptPassword("要修改的服务器密码"))
			if err != nil {
				c.App.PrintError(errors.New("操作已取消"))
				return nil
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			var serverInfo *serverclientcommon.ServerInfo
			for i := range serverNameList {
				serverName := serverNameList[i]

				serverInfoBytes, err := serverclientcommon.CmdRemoteServerInfo.ExchangeWithData(serverName, stream)
				if err != nil {
					c.App.PrintError(fmt.Errorf("获取[%s]的信息失败: %s", serverName, err.Error()))
					continue
				}

				if err = serverInfoBytes.UnmarshalJson(&serverInfo); err != nil {
					c.App.PrintError(fmt.Errorf("反序列化服务器[%s]的信息失败: %s", serverName, err.Error()))
					continue
				}

				serverInfo.SSHPassword = password
				if _, err = serverclientcommon.CmdRemoteServerUpdate.ExchangeWithData(&serverInfo, stream); err != nil {
					c.App.PrintError(fmt.Errorf("修改[%s]的密码失败: %s", serverName, err.Error()))
				}
			}

			_, _ = c.App.Println()

			return nil
		},
	}

	// remoteServerAliasUpdateCmd 远程服务器别名
	remoteServerAliasUpdateCmd = &grumble.Command{
		Name: "alias",
		Help: "设置服务器别名, 方便记忆或书写",
		Completer: func(prefix string, args []string) []string {
			if len(args) != 0 {
				return nil
			}
			return remoteServerIpListByPrefixAndExclude(prefix, nil)
		},
		Args: func(a *grumble.Args) {
			a.String("ip", "服务器IP地址")
			a.StringList("alias", "要设置的别名, 可以有多个, 留空则清除", grumble.Min(0))
		},
		Run: func(c *grumble.Context) error {

			ip := c.Args.String("ip")
			aliasList := c.Args.StringList("alias")

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			if err = serverclientcommon.CmdRemoteServerUpdateAlias.SendCommand(stream); err != nil {
				return err
			}

			if err = stream.WriteMsg([]byte(ip), transportstream.MsgFlagSuccess); err != nil {
				return err
			}

			if err = stream.WriteJsonMsg(aliasList); err != nil {
				return err
			}

			if _, err = stream.ReceiveMsg(); err != nil && err != transportstream.StreamIsEnd {
				_ = stream.WriteEndMsg()
			}

			return nil
		},
	}

	// remoteServerRemoveCmd 删除服务器命令
	remoteServerRemoveCmd = &grumble.Command{
		Name:      "rm",
		Help:      "根据服务器IP或别名删除服务器信息",
		Completer: remoteServerNameListByPrefixAndExclude,
		Flags: func(f *grumble.Flags) {
			f.Bool("y", "yes", false, "跳过删除确认")
		},
		Args: func(a *grumble.Args) {
			a.StringList("serverNameList", "要删除服务器名称列表", grumble.Min(1))
		},
		Run: func(c *grumble.Context) error {
			confirm := c.Flags.Bool("yes")
			serverNameList := c.Args.StringList("serverNameList")
			if len(serverNameList) == 0 {
				return nil
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			for i := range serverNameList {
				sn := serverNameList[i]
				if !confirm {
					if _, err = c.ShellTools.Prompt(utils.PromptConfirm("确认删除 " + sn)); err != nil {
						if err.Error() == "" {
							continue
						}
						c.App.PrintError(fmt.Errorf("操作被取消"))
						return nil
					}
				}

				if _, err = serverclientcommon.CmdRemoteServerDel.ExchangeWithData(sn, stream); err != nil {
					c.App.PrintError(fmt.Errorf("删除服务器[%s]失败: %s", sn, err.Error()))
				}
			}
			return nil
		},
	}
)

func remoteServerNameListByPrefixAndExclude(prefix string, excludeList []string) []string {
	return remoteServerAllNameList(true, prefix, excludeList)
}

func remoteServerIpListByPrefixAndExclude(prefix string, excludeList []string) []string {
	return remoteServerAllNameList(false, prefix, excludeList)
}

func remoteServerAllNameList(includeAlias bool, prefix string, excludeList []string) (res []string) {
	unixServerInfo := server.NewUnixServerInfo()
	defer unixServerInfo.Close()

	stream, err := unixServerInfo.ConnToStream()
	if err != nil {
		return
	}

	serverInfoListBytes, err := serverclientcommon.CmdRemoteServerList.Exchange(stream)
	if err != nil {
		return
	}

	var serverInfoList []*serverclientcommon.ServerInfo
	if err = serverInfoListBytes.UnmarshalJson(&serverInfoList); err != nil {
		return
	}

	res = make([]string, 0, len(serverInfoList)*2)
	for i := range serverInfoList {
		serverInfo := serverInfoList[i]
		if slices.Contains(excludeList, serverInfo.Id) {
			continue
		}

		if strings.HasPrefix(serverInfo.Id, prefix) {
			res = append(res, serverInfo.Id)
		}

		if includeAlias {
			for j := range serverInfo.Alias {
				alias := serverInfo.Alias[j]
				if !strings.HasPrefix(alias, prefix) || slices.Contains(excludeList, alias) {
					continue
				}

				res = append(res, alias)
			}
		}
	}
	return
}
