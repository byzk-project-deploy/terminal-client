package cmdmodel

import "github.com/desertbit/grumble"

type ModelName string

const (
	ModelBypt   ModelName = "bypt"
	ModelServer ModelName = "server"
)

type ModelInfo struct {
	Name                ModelName
	Commands            []*grumble.Command
	ServerList          *ServerList
	NoFindCommandHandle func(app *grumble.App, args []string) error
}

func (m *ModelInfo) Convert() {
	currentModelInfo = m
	if currentApp == nil {
		return
	}

	m.settingPrompt()

	currentApp.SetNoFindCommandHandler(m.NoFindCommandHandle)
	currentApp.Commands().RemoveAll()
	for i := range m.Commands {
		command := m.Commands[i]
		currentApp.AddCommand(command)
	}
}

func (m *ModelInfo) settingPrompt() {
	if currentApp == nil || currentModelInfo == nil {
		return
	}

	promptPrefix := string(m.Name)
	// if m.Name != "" {
	// 	promptPrefix += "@" + name
	// }
	promptStr := promptPrefix + " » "
	currentApp.SetPrompt(promptStr)

}
