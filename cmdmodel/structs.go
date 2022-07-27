package cmdmodel

import (
	"fmt"
	"github.com/desertbit/grumble"
)

type ModelName string

func (m ModelName) Convert() error {
	modelInfo, ok := modelMap[m]
	if !ok {
		return fmt.Errorf("未配置的模式选项")
	}

	currentModelInfo = modelInfo
	return nil
}

const (
	ModelBypt   ModelName = "bypt"
	ModelServer ModelName = "server"
)

type ModelInfo struct {
	*ServerList
	Name                ModelName
	Commands            []*grumble.Command
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
