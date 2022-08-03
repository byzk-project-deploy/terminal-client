package cmd

import (
	"fmt"
	rpcinterfaces "github.com/byzk-project-deploy/base-interface"
	"github.com/byzk-project-deploy/grumble"
	packaging_plugin "github.com/byzk-project-deploy/packaging-plugin"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	// pluginConvertCmd 插件模式转换命令
	pluginConvertCmd = &grumble.Command{
		Name: "plugin",
		Help: "切换为插件模式",
		Run: func(c *grumble.Context) error {
			return cmdmodel.ModelPlugin.Convert()
		},
	}

	// pluginPackCmd 插件打包命令
	pluginPackCmd = &grumble.Command{
		Name: "pack",
		Help: "打包插件",
		Args: func(a *grumble.Args) {
			a.String("plugin source file", "插件的原始文件地址")
			a.String("plugin dest file", "插件打包之后的存放地址, 默认为: ($sourceFile)_pack", grumble.Default(""))
		},
		Flags: func(f *grumble.Flags) {
			f.Bool("d", "directory", false, "是否打包目录中的所有文件")
			f.Bool("r", "recursive", false, "是否递归目录中的所有子目录，只有开启[directory]参数时, 该参数才会生效")
		},
		Run: func(c *grumble.Context) error {
			rangeDir := c.Flags.Bool("directory")
			recursive := c.Flags.Bool("recursive")

			sourceFilePath := c.Args.String("plugin source file")
			destFilePath := c.Args.String("plugin dest file")
			if destFilePath == "" {
				destFilePath = sourceFilePath + "_pack"
			}

			if rangeDir {

				stat, err := os.Stat(destFilePath)
				if err != nil || !stat.IsDir() {
					if err := os.MkdirAll(destFilePath, 0755); err != nil {
						return fmt.Errorf("创建目标目录[%s]失败: %s", destFilePath, err.Error())
					}
				}

				fileLen := rangeDirGetFilePathList(sourceFilePath, recursive, func(filename string) {
					if err := pluginPack(filename, filepath.Join(destFilePath, filepath.Base(filename)+"_pack")); err != nil {
						c.App.PrintError(fmt.Errorf("打包文件[%s]失败: %s", filename, err.Error()))
					}
				})

				if fileLen == 0 {
					return fmt.Errorf("未在目录[%s]中找到插件文件", sourceFilePath)
				}
			} else {
				return pluginPack(sourceFilePath, destFilePath)
			}
			return nil
		},
	}

	// pluginInstallCmd 插件安装命令
	pluginInstallCmd = &grumble.Command{
		Name: "install",
		Help: "安装插件",
		Args: func(a *grumble.Args) {
			a.String("plugin file or dir path", "插件文件或目录地址")
		},
		Flags: func(f *grumble.Flags) {
			f.Bool("d", "directory", false, "是否安装目录中的所有文件")
			f.Bool("r", "recursive", false, "是否递归目录中的所有子目录，只有开启[directory]参数时, 该参数才会生效")
		},
		Run: func(c *grumble.Context) error {
			unixServer := server.NewUnixServerInfo()
			stream, err := unixServer.ConnToStream()
			if err != nil {
				return fmt.Errorf("连接Unix服务失败: %s", err.Error())
			}

			rangeDir := c.Flags.Bool("directory")
			recursive := c.Flags.Bool("recursive")

			var pluginInfo *serverclientcommon.DbPluginInfo
			pluginFileOrDir := c.Args.String("plugin file or dir path")
			if rangeDir {
				l := len(pluginFileOrDir)
				total := rangeDirGetFilePathList(pluginFileOrDir, recursive, func(filename string) {
					pluginFileName := filename[l+1:]
					if res, err := serverclientcommon.CmdPluginInstall.ExchangeWithData(filename, stream); err != nil {
						c.App.PrintError(fmt.Errorf("插件包[%s]安装失败: %s", pluginFileName, err.Error()))
					} else {
						if err = res.UnmarshalJson(&pluginInfo); err != nil {
							c.App.PrintError(fmt.Errorf("插件文件[%s]安装失败: %s", pluginFileName, err.Error()))
							return
						}
						color.Green("插件[%s]安装成功", pluginInfo.Name)
					}
				})
				if total == 0 {
					return fmt.Errorf("未在目录[%s]中找到插件文件", pluginFileOrDir)
				}
			} else {
				if res, err := serverclientcommon.CmdPluginInstall.ExchangeWithData(pluginFileOrDir, stream); err != nil {
					return err
				} else {
					if err = res.UnmarshalJson(&pluginInfo); err != nil {
						return fmt.Errorf("转换安装之后的插件数据失败: %s", err.Error())
					}
					color.Green("插件[%s]安装成功", pluginInfo.Name)
				}
			}
			_, _ = c.App.Println()
			return nil
		},
	}

	// pluginListCmd 插件列表命令
	pluginListCmd = &grumble.Command{
		Name:     "ls",
		Help:     "查看插件列表",
		LongHelp: "查看已安装的插件列表以及插件的当前状态",
		Run: func(c *grumble.Context) error {
			unixServer := server.NewUnixServerInfo()
			stream, err := unixServer.ConnToStream()
			if err != nil {
				return err
			}

			res, err := serverclientcommon.CmdPluginList.Exchange(stream)
			if err != nil {
				return fmt.Errorf("获取插件列表失败: %s", err.Error())
			}

			var pluginStatusInfoList []*serverclientcommon.PluginStatusInfo

			if err = res.UnmarshalJson(&pluginStatusInfoList); err != nil {
				return fmt.Errorf("转换插件列表数据失败: %s", err.Error())
			}

			table := uitable.New()
			table.MaxColWidth = 30
			table.Wrap = true
			table.AddRow("ID", "名称", "描述", "当前状态", "错误信息")
			for i := range pluginStatusInfoList {
				pluginStatusInfo := pluginStatusInfoList[i]
				id := pluginStatusInfo.Id[0:30]
				status := "未运行"
				switch pluginStatusInfo.Status {
				case serverclientcommon.PluginStatusOk:
					status = color.GreenString("正在运行")
				case serverclientcommon.PluginStatusRebooting:
					status = color.YellowString("正在重启")
				case serverclientcommon.PluginStatusErr:
					status = color.RedString("发生错误")
				}
				table.AddRow(id, pluginStatusInfo.Name, pluginStatusInfo.ShortDesc, status)
			}
			_, _ = c.App.Println(table)
			return nil
		},
	}

	// pluginInfoCmd 插件信息查询命令
	pluginInfoCmd = &grumble.Command{
		Name:  "info",
		Help:  "查看插件的详细信息",
		Usage: "info pluginIdOrName",
		Completer: func(prefix string, args []string) (res []string) {
			res = make([]string, 0)

			if len(args) > 0 {
				return
			}

			unixServerInfo := server.NewUnixServerInfo()
			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return
			}

			serverData, err := serverclientcommon.CmdPluginInfoPromptList.ExchangeWithData(prefix, stream)
			if err != nil {
				return
			}

			var pluginInfoList []*serverclientcommon.PluginStatusInfo
			_ = serverData.UnmarshalJson(&pluginInfoList)
			for i := range pluginInfoList {
				pluginInfo := pluginInfoList[i]
				if strings.Contains(pluginInfo.Name, " ") {
					res = append(res, "\""+pluginInfo.Name+"\"")
					continue
				}
				res = append(res, pluginInfo.Name)
			}
			return
		},
		Args: func(a *grumble.Args) {
			a.String("id", "插件的ID或插件名称")
		},
		Run: func(c *grumble.Context) error {
			unixServerInfo := server.NewUnixServerInfo()
			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			serverData, err := serverclientcommon.CmdPluginInfo.ExchangeWithData(c.Args.String("id"), stream)
			if err != nil {
				return err
			}

			var pluginInfo *serverclientcommon.PluginStatusInfo
			if err = serverData.UnmarshalJson(&pluginInfo); err != nil {
				return fmt.Errorf("转换服务器数据失败: %s", err.Error())
			}

			table := uitable.New()

			pluginType := ""
			if pluginInfo.Type.Is(rpcinterfaces.PluginTypeCmd) {
				pluginType += "终端 "
			}

			if pluginInfo.Type.Is(rpcinterfaces.PluginTypeWeb) {
				pluginType += "WEB "
			}

			errMsg := ""
			if pluginInfo.Msg != "" {
				errMsg = color.RedString(pluginInfo.Msg)
			}

			status := "未运行"
			switch pluginInfo.Status {
			case serverclientcommon.PluginStatusOk:
				status = color.GreenString("正在运行")
			case serverclientcommon.PluginStatusRebooting:
				status = color.YellowString("正在重启")
			case serverclientcommon.PluginStatusErr:
				status = color.RedString("发生错误")
			}

			var enableText string
			if pluginInfo.Enable {
				enableText = color.YellowString("未启用")
			} else {
				enableText = color.GreenString("已启用")
			}

			table.AddRow("ID:", pluginInfo.Id)
			table.AddRow("名称:", pluginInfo.Name)
			table.AddRow("作者:", pluginInfo.Author)
			table.AddRow("类型:", pluginType)
			table.AddRow("描述:", pluginInfo.ShortDesc)
			table.AddRow("详细说明:", pluginInfo.Desc)
			table.AddRow("启用状态:", enableText)
			table.AddRow("运行状态:", status)
			table.AddRow("安装时间:", utils.TimeFormat(pluginInfo.InstallTime))
			table.AddRow("创建时间:", utils.TimeFormat(pluginInfo.CreateTime))
			table.AddRow("启动时间:", utils.TimeFormat(pluginInfo.StartTime))
			table.AddRow("停止时间:", utils.TimeFormat(pluginInfo.StopTime))
			table.AddRow("错误信息:", errMsg)

			_, _ = c.App.Println(table, "\n")

			return nil
		},
	}
)

func rangeDirGetFilePathList(dirName string, recursive bool, callback func(filename string)) int {

	stat, err := os.Stat(dirName)
	if err != nil || !stat.IsDir() {
		return 0
	}

	dirs, err := ioutil.ReadDir(dirName)
	if err != nil {
		return 0
	}

	res := 0
	for i := range dirs {
		dir := dirs[i]
		if dir.IsDir() {
			if recursive {
				res += rangeDirGetFilePathList(filepath.Join(dirName, dir.Name()), recursive, callback)
			}
			continue
		}
		callback(filepath.Join(dirName, dir.Name()))
		res += 1
	}
	return res
}

func pluginPack(srcFilePath, targetFilePath string) error {
	if err := packaging_plugin.Packing(srcFilePath, targetFilePath); err != nil {
		_ = os.RemoveAll(targetFilePath)
		return err
	}
	return nil
}
