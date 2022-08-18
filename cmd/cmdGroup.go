package cmd

import (
	"github.com/byzk-project-deploy/grumble"
)

var (
	// defaultCommands 默认的命令
	defaultCommands = []*grumble.Command{
		shellCommand,
		pwdCommand,
		cdCommand,
		callCommand,
		historyCmd,
		pluginConvertCmd,
		remoteServerConvertCmd,
		{
			Name: "exit",
			Help: "退出bypt工具",
			Run: func(c *grumble.Context) error {
				c.Stop()
				return nil
			},
		},
	}

	// pluginCommands  插件模式下的命令
	pluginCommands = []*grumble.Command{
		pluginPackCmd,
		pluginInstallCmd,
		pluginListCmd,
		pluginInfoCmd,
		modelExitCmd,
		historyCmd,
	}

	// serverCommands 远程服务命令
	serverCommands = []*grumble.Command{
		remoteServerJoinCmd,
		remoteServerListCmd,
		remoteServerInfoCmd,
		remoteServerPasswordUpdateCmd,
		remoteServerAliasUpdateCmd,
		remoteServerRemoveCmd,
		remoteServerRepairCmd,
		remoteServerUploadCmd,
		remoteServerDownloadCmd,
		modelExitCmd,
		historyCmd,
	}
)
