package cmd

import "github.com/desertbit/grumble"

var (
	byptCommands = []*grumble.Command{
		shellCommand,
		pwdCommand,
		cdCommand,
		callCommand,
		historyCmd,
		{
			Name: "exit",
			Help: "exit the shell",
			Run: func(c *grumble.Context) error {
				c.Stop()
				return nil
			},
		},
	}
)
