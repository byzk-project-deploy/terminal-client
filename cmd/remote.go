package cmd

import (
	"errors"
	"fmt"
	"github.com/akrennmair/slice"
	"github.com/byzk-project-deploy/grumble"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/fatih/color"
	transportstream "github.com/go-base-lib/transport-stream"
	"github.com/gosuri/uitable"
	"github.com/pterm/pterm"
	"golang.org/x/exp/slices"
	"io"
	"math"
	"os"
	"path/filepath"
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
					statusText = color.RedString("连接失败")
				case serverclientcommon.ServerStatusNoRun:
					statusText = color.YellowString("未连接")
				case serverclientcommon.ServerRunning:
					statusText = color.GreenString("已连接")
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
				statusText = color.RedString("连接失败")
			case serverclientcommon.ServerStatusNoRun:
				statusText = color.YellowString("未连接")
			case serverclientcommon.ServerRunning:
				statusText = color.GreenString("已连接")
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
					if ok, err := c.ShellTools.Confirm("确认删除 " + sn); err != nil {
						c.App.PrintError(fmt.Errorf("操作被取消"))
					} else if !ok {
						continue
					}
				}

				if _, err = serverclientcommon.CmdRemoteServerDel.ExchangeWithData(sn, stream); err != nil {
					c.App.PrintError(fmt.Errorf("删除服务器[%s]失败: %s", sn, err.Error()))
				}
			}
			return nil
		},
	}

	// remoteServerRepairCmd 服务修复命令
	remoteServerRepairCmd = &grumble.Command{
		Name: "repair",
		Help: "尝试自动修复有问题的远程服务",
		Run: func(c *grumble.Context) error {
			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			resData, err := serverclientcommon.CmdRemoteServerRepair.Exchange(stream)
			if err != nil {
				return err
			}

			var repairResultList []*serverclientcommon.RemoteServerRepairResInfo
			if err = resData.UnmarshalJson(&repairResultList); err != nil {
				return fmt.Errorf("转换修复结果信息失败: %s", err.Error())
			}

			for i := range repairResultList {
				repairResult := repairResultList[i]
				if !repairResult.Success {
					c.App.PrintError(fmt.Errorf("服务器[%s]修复失败: %s", repairResult.Ip, repairResult.ErrMsg))
				}
			}

			return nil
		},
	}

	// remoteServerUploadCmd 远程服务文件上传命令
	remoteServerUploadCmd = &grumble.Command{
		Name: "upload",
		Help: "上传文件或目录到远程服务器中",
		Usage: `
[上传本地文件或目录到所有的服务器上]
upload /home/test/a.txt ~/a.txt

[上传本地文件到指定服务器]
upload -s 192.168.100.2,192.168.100.3 /home/test/a.txt ~/a.txt

[上传本地文件到服务器不同的目录地址]
upload /home/test/a.txt 192.168.100.2@192.168.100.2

[上传本地文件到服务器, 个别服务器路径与其他不同]
upload /home/test/a.txt ~/a.txt 192.168.3@~/home/a.txt

[将远程服务器中的文件上传到除原始服务器之外的其他服务器上]
upload 192.168.100.2@~/a.txt ~/a.txt
`,
		Completer: func(prefix string, args []string) (res []string) {
			res = remoteServerNameListByPrefixAndExclude(prefix, args)
			res = slice.Map(res, func(t1 string) string {
				return t1 + "@"
			})
			argsLen := len(args)
			if argsLen > 0 {
				return
			}

			if prefix == "" {
				prefix = server.CurrentPath()
			}

			if strings.ContainsRune(prefix, '@') {
				return
			}

			prefix = getRealFilePath(prefix)

			dirname := filepath.Dir(prefix)
			stat, err := os.Stat(dirname)
			if err != nil || !stat.IsDir() {
				return
			}

			filename := filepath.Base(prefix)
			if strings.HasSuffix(prefix, "/") {
				filename = ""
			}
			childFiles, err := os.ReadDir(dirname)
			if err != nil {
				return
			}

			for i := range childFiles {
				f := childFiles[i]
				if strings.HasPrefix(f.Name(), filename) && f.Name() != filename {
					res = append(res, f.Name()[len(filename):])
				}
			}

			return
		},
		Flags: func(f *grumble.Flags) {
			f.Uint("t", "type", 0, "上传的通道类型: 0: ssh+ftp")
			f.Bool("r", "recursive", false, "递归上传")
			f.String("s", "includeServer", "", "要上传到的服务器名称,可以是服务器的ip或别名, 多个服务ip器或别名使用逗号隔开, 优先级大于排除")
			f.String("e", "excludeServer", "", "排除的服务器名称, 可以是服务器的ip或别名, 多个服务器ip或别名使用逗号分割")
		},
		Args: func(a *grumble.Args) {
			a.String("sourceFile", "原始文件或目录, 格式为: /home/text.txt 或 192.168.100.218@/home/test/a.txt, 可以是文件或目录, 当为目录时必须携带 -r 参数")
			a.StringList("targetFile", "目标文件名称", grumble.Min(1))
		},
		Run: func(c *grumble.Context) (err error) {
			var (
				uploadType    = serverclientcommon.UploadType(uint8(c.Flags.Uint("type")))
				recursive     = c.Flags.Bool("recursive")
				includeServer []string
				excludeServer []string

				targetFileOrDirList = c.Args.StringList("targetFile")
				sourceFileOrDir     = c.Args.String("sourceFile")

				sourceAddr     *serverclientcommon.UploadAddrInfo
				targetAddrList []*serverclientcommon.UploadAddrInfo

				serverListExchangeData serverclientcommon.ExchangeData
				serverList             []*serverclientcommon.ServerInfo
			)

			includeServerFlag := c.Flags.String("includeServer")
			excludeServerFlag := c.Flags.String("excludeServer")
			if includeServerFlag != "" {
				includeServer = strings.Split(includeServerFlag, ",")
			}

			if excludeServerFlag != "" {
				excludeServer = strings.Split(excludeServerFlag, ",")
			}

			if uploadType < serverclientcommon.UploadTypeSSHFtp || uploadType >= serverclientcommon.UploadUnknown {
				return fmt.Errorf("不支持的传输协议类型")
			}

			if sourceAddr, err = uploadPathConvertToUploadAddrInfo(sourceFileOrDir); err != nil {
				return
			}

			if strings.HasSuffix(sourceAddr.Path, "/") {
				return fmt.Errorf("不支持路径以 / 结尾")
			}

			if len(targetFileOrDirList) == 0 {
				return fmt.Errorf("缺失目标路径")
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return
			}
			defer stream.WriteEndMsg()

			if serverListExchangeData, err = serverclientcommon.CmdRemoteServerList.ExchangeWithData(nil, stream); err != nil {
				return
			}

			if err = serverListExchangeData.UnmarshalJson(&serverList); err != nil {
				return
			}

			serverListLen := len(serverList)
			if serverListLen == 0 {
				return fmt.Errorf("未查询到已存在的服务器信息, 请先添加服务器信息")
			}

			serverNameList := make([]string, 0, serverListLen)
			for i := range serverList {
				serverInfo := serverList[i]
				serverNameList = append(serverNameList, serverInfo.Id)
				serverNameList = append(serverNameList, serverInfo.Alias...)
			}

			if sourceAddr.Server != "" && !slices.Contains(serverNameList, sourceAddr.Server) {
				return fmt.Errorf("不存在的服务器[%s]", sourceAddr.Server)
			}

			isHaveInclude := len(excludeServer) > 0
			globalUploadPathList := make([]string, 0, 8)
			targetAddrList = make([]*serverclientcommon.UploadAddrInfo, 0, serverListLen)
			for _, targetPath := range targetFileOrDirList {
				if strings.HasSuffix(targetPath, "/") {
					return fmt.Errorf("不支持路径以 / 结尾")
				}
				if !strings.Contains(targetPath, "@") {
					globalUploadPathList = append(globalUploadPathList, targetPath)
					continue
				}

				targetFlags := strings.Split(targetPath, "@")
				if len(targetFlags) != 2 {
					return fmt.Errorf("错误的目标地址格式: %s", targetPath)
				}

				if !slices.Contains(serverNameList, targetFlags[0]) {
					return fmt.Errorf("不存在的服务器IP或名称: %s", targetFlags[0])
				}

				if isHaveInclude && !slices.Contains(includeServer, targetFlags[0]) {
					return fmt.Errorf("地址[%s]不再白名单内", targetPath)
				}

				serverList = slice.Filter(serverList, func(s *serverclientcommon.ServerInfo) bool {
					return s.Id == targetFlags[0] || slices.Contains(s.Alias, targetFlags[0])
				})

				targetAddrList = append(targetAddrList, &serverclientcommon.UploadAddrInfo{
					Server: targetFlags[0],
					Path:   targetFlags[1],
				})
			}

			isHaveExclude := len(excludeServer) > 0
			if len(globalUploadPathList) > 0 {
				for i := range serverList {
					serverInfo := serverList[i]

					if isHaveInclude && slices.IndexFunc(includeServer, func(n string) bool {
						return n == serverInfo.Id || slices.Contains(includeServer, n)
					}) == 0 {
						continue
					}

					if isHaveExclude && slices.IndexFunc(includeServer, func(n string) bool {
						return n == serverInfo.Id || slices.Contains(includeServer, n)
					}) > 0 {
						continue
					}

					for _, p := range globalUploadPathList {
						targetAddrList = append(targetAddrList, &serverclientcommon.UploadAddrInfo{
							Server: serverInfo.Id,
							Path:   p,
						})
					}
				}
			}

			if len(targetAddrList) == 0 {
				return fmt.Errorf("缺失目标地址")
			}

			if err = serverclientcommon.CmdRemoteServerFileUpload.SendCommand(stream); err != nil {
				return
			}

			if err = stream.WriteJsonMsg(&serverclientcommon.RemoteServerUploadRequest{
				Recursive:      recursive,
				Include:        includeServer,
				Exclude:        excludeServer,
				SourceAddr:     sourceAddr,
				TargetAddrList: targetAddrList,
				UploadType:     uploadType,
			}); err != nil {
				return
			}

			if _, err = stream.ReceiveMsg(); err != nil {
				return
			}

			if err = stream.WriteMsg(nil, transportstream.MsgFlagSuccess); err != nil {
				return
			}

			var (
				nowUploadFilename serverclientcommon.ExchangeData
			)

			for {
				if nowUploadFilename, err = stream.ReceiveMsg(); err != nil {
					if err == transportstream.StreamIsEnd {
						return nil
					}
					return fmt.Errorf("获取正在上传的文件信息失败")
				}

				if err = stream.WriteMsg(nil, transportstream.MsgFlagSuccess); err != nil {
					return
				}

				if err = uploadProgress(nowUploadFilename, stream); err != nil {
					return
				}
			}
		},
	}

	// remoteServerDownloadCmd 远程服务器文件或目录下载
	remoteServerDownloadCmd = &grumble.Command{
		Name: "download",
		Help: "下载远程文件到本地",
		Completer: func(prefix string, args []string) (res []string) {
			argsLen := len(args)
			if argsLen == 0 {
				if prefix == "" {
					prefix = server.CurrentPath() + "/"
				}

				prefix = getRealFilePath(prefix)

				dirname := filepath.Dir(prefix)
				stat, err := os.Stat(dirname)
				if err != nil || !stat.IsDir() {
					return
				}

				filename := filepath.Base(prefix)
				if strings.HasSuffix(prefix, "/") {
					filename = ""
				}
				childFiles, err := os.ReadDir(dirname)
				if err != nil {
					return
				}

				for i := range childFiles {
					f := childFiles[i]
					if strings.HasPrefix(f.Name(), filename) && f.Name() != filename {
						res = append(res, f.Name()[len(filename):])
					}
				}
				return
			}

			res = remoteServerNameListByPrefixAndExclude(prefix, args)
			res = slice.Map(res, func(t1 string) string {
				return t1 + "@"
			})
			return
		},
		Args: func(a *grumble.Args) {
			a.String("saveDir", "保存到本地的哪个目录, 如果不存在该目录将会创建")
			a.StringList("remoteAddr", "远程服务器及文件地址, 格式: 服务器IP/别名@路径", grumble.Min(1))
		},
		Flags: func(f *grumble.Flags) {
			f.Uint("t", "type", 0, "上传的通道类型: 0: ssh+ftp")
		},
		Run: func(c *grumble.Context) error {
			var (
				saveDir           = c.Args.String("saveDir")
				remoteAddrStrList = c.Args.StringList("remoteAddr")

				protoType = serverclientcommon.UploadType(uint8(c.Flags.Uint("type")))
			)

			if protoType < serverclientcommon.UploadTypeSSHFtp || protoType >= serverclientcommon.UploadUnknown {
				return fmt.Errorf("不支持的传输协议类型")
			}

			saveDir = getRealFilePath(saveDir)
			if strings.HasSuffix(saveDir, "/") {
				return fmt.Errorf("不支持路径以 / 结尾")
			}

			unixServerInfo := server.NewUnixServerInfo()
			defer unixServerInfo.Close()

			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			_ = os.MkdirAll(saveDir, 0755)

			for i := range remoteAddrStrList {
				addStr := remoteAddrStrList[i]
				if err = remoteDownloadHandle(addStr, protoType, saveDir, stream); err != nil {
					return err
				}
			}

			return nil
		},
	}
)

func remoteDownloadHandle(addr string, protoType serverclientcommon.UploadType, savedir string, stream *transportstream.Stream) (err error) {
	defer stream.WriteEndMsg()

	var (
		tBytes serverclientcommon.ExchangeData
	)

	if err = serverclientcommon.CmdRemoteServerFileDownload.SendCommand(stream); err != nil {
		return err
	}

	if tBytes, err = transportstream.IntToBytes[uint8](uint8(protoType)); err != nil {
		return err
	} else if err = stream.WriteMsgStream(tBytes, transportstream.MsgFlagSuccess).
		WriteMsgStream([]byte(addr), transportstream.MsgFlagSuccess).
		Error(); err != nil {
		return err
	}

	filenameExchangeData, err := stream.ReceiveMsg()
	if err != nil {
		return err
	}

	relativePathExchangeData, err := stream.ReceiveMsg()
	if err != nil {
		return err
	}

	fileSizeExchangeData, err := stream.ReceiveMsg()
	if err != nil {
		return err
	}

	filesize, err := transportstream.BytesToInt[int64](fileSizeExchangeData)
	if err != nil {
		return fmt.Errorf("转换文件的长度失败: %s", err.Error())
	}

	if err = stream.WriteMsg(nil, transportstream.MsgFlagSuccess); err != nil {
		return err
	}

	filename := string(filenameExchangeData)

	relativePath := string(relativePathExchangeData)
	if relativePath == "" {
		relativePath = filepath.Base(filename)
	}

	targetPath := filepath.Join(savedir, relativePath)
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return fmt.Errorf("创建本地文件[%s]失败: %s", targetPath, err.Error())
	}
	defer f.Close()

	p, _ := pterm.DefaultProgressbar.WithTotal(100).WithTitle(fmt.Sprintf("正在下载文件: %s", filepath.Join(addr, string(relativePathExchangeData)))).WithRemoveWhenDone(true).Start()
	defer p.Stop()

	var (
		d       serverclientcommon.ExchangeData
		okCount = 0
	)
	for {
		d, err = stream.ReceiveMsg()
		if err == transportstream.StreamIsEnd {
			break
		}

		if err == io.EOF {
			return fmt.Errorf("传输流异常关闭")
		}

		if err != nil {
			return fmt.Errorf("文件传输发生错误: %s", err.Error())
		}

		okCount += len(d)
		targetPercentage := int(math.Floor(float64(okCount) / float64(filesize) * 100))
		addPercentage := targetPercentage - p.Current

		if p.Current+addPercentage > 99 {
			addPercentage = 99 - p.Current
		}

		_ = p.Add(addPercentage)

		if _, err = f.Write(d); err != nil {
			return fmt.Errorf("写出内容到文件[%s]失败: %s", targetPath, err.Error())
		}
	}
	pterm.Success.Printfln("服务器文件[%s]成功保存至: %s", addr, targetPath)
	return nil
}

func uploadProgress(filename serverclientcommon.ExchangeData, stream *transportstream.Stream) (err error) {
	var uploadResponse *serverclientcommon.RemoteServerUploadResponse

	p, _ := pterm.DefaultProgressbar.WithTotal(100).WithTitle(fmt.Sprintf("正在上传文件: %s", filename)).WithRemoveWhenDone(true).Start()
	defer p.Stop()
	for {
		if err = stream.ReceiveJsonMsg(&uploadResponse); err != nil {
			return
		}

		if !uploadResponse.Success {
			pterm.Error.Printfln(uploadResponse.ErrMsg)
			continue
		}

		if uploadResponse.End {
			pterm.Success.Printfln("文件[%s]上传成功", filename)
			_, _ = p.Stop()
			break
		}

		progress := uploadResponse.Progress
		addCount := progress - p.Current
		if addCount <= 0 {
			continue
		}

		if p.Current+addCount >= 100 {
			addCount = 99 - p.Current
		}

		p.Add(addCount)
	}
	return nil
}

func uploadPathConvertToUploadAddrInfo(p string) (sourceAddr *serverclientcommon.UploadAddrInfo, err error) {
	if strings.Contains(p, "@") {
		serverAddrArr := strings.Split(p, "@")
		if len(serverAddrArr) != 2 {
			return nil, fmt.Errorf("不符合规范的地址, 如果是务器文件路径格式参照: 服务器IP/别名@服务器上的文件地址")
		}
		sourceAddr = &serverclientcommon.UploadAddrInfo{
			Server: serverAddrArr[0],
			Path:   serverAddrArr[1],
		}
	} else {
		sourceAddr = &serverclientcommon.UploadAddrInfo{
			Path: getRealFilePath(p),
		}
	}
	return sourceAddr, nil
}

func getRealFilePath(p string) string {
	if strings.ContainsRune(p, '~') {
		currentUser := user.Current()
		homeDir := currentUser.HomeDir
		p = strings.ReplaceAll(p, "~", homeDir)
	}

	if p[0] != '/' {
		currentPath := server.CurrentPath()
		p = filepath.Join(currentPath, p)
	}

	return p
}

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
