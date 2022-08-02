package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/byzk-project-deploy/terminal-client/cmdmodel"
	"github.com/byzk-project-deploy/terminal-client/loading"
	"github.com/desertbit/grumble"
)

var (
	historyFile = filepath.Join(os.TempDir(), "bypt.hist")
)

type ContextWrapper struct {
	*grumble.Context
	haveErr         bool
	successCallback func() error
}

func (c *ContextWrapper) PrintError(err error) {
	loading.Spinner().Stop()
	c.haveErr = true
	c.Context.App.PrintError(err)
}

func (c *ContextWrapper) Success() bool {
	return !c.haveErr
}

func (c *ContextWrapper) SuccessCallback(fn func() error) {
	c.successCallback = fn
}

func cmdErrRunWrapper(f func(c *ContextWrapper) error) func(c *grumble.Context) error {
	return func(c *grumble.Context) error {
		contextWrapper := &ContextWrapper{
			Context: c,
		}

		if err := f(contextWrapper); err != nil {
			return err
		}

		if contextWrapper.haveErr {
			return fmt.Errorf("命令均已在对应服务器上部执行, 但执行过程中发生异常")
		}

		if contextWrapper.successCallback != nil {
			return contextWrapper.successCallback()
		}

		return nil
	}
}

// func initCmd() {
// 	// initHistoryCmd()
// 	// initSystemCmd()
// 	modelConvert(modelBypt)
// 	current.SetNoFindCommandHandler(noFindCommandHandle)
// }

func init() {
	cmdmodel.Registry(cmdmodel.ModelDefault, &cmdmodel.ModelInfo{
		Commands:            defaultCommands,
		NoFindCommandHandle: noFindCommandHandle,
	})

	cmdmodel.Registry(cmdmodel.ModelPlugin, &cmdmodel.ModelInfo{
		Commands: pluginCommands,
	})
}

func Run() {
	cmdmodel.InitApp(cmdmodel.ModelDefault, historyFile)
}
