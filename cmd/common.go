package cmd

import (
	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/desertbit/grumble"
)

var (
	modelExitCmd = &grumble.Command{
		Name:     "exit",
		Help:     "退出当前模式",
		LongHelp: "退出当前模式到默认模式",
		Flags: func(f *grumble.Flags) {
			f.Bool("f", "force", false, "强制退出, 当为true时将退出整个程序而不是回到默认模式")
		},
		Run: func(c *grumble.Context) error {
			if c.Flags.Bool("force") {
				c.Stop()
				return nil
			}
			return cmdmodel.ModelDefault.Convert()
		},
	}
)
