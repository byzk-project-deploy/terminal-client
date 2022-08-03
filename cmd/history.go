package cmd

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/byzk-project-deploy/grumble"
)

var historyCmd = &grumble.Command{
	Name:     "history",
	Help:     "查看历史命令",
	LongHelp: "查看在本应用内使用过的命令",
	Flags: func(f *grumble.Flags) {
		f.BoolL("clear", false, "清除历史命令")
	},
	Run: func(c *grumble.Context) error {
		isClear := c.Flags.Bool("clear")
		if isClear {
			if err := ioutil.WriteFile(historyFile, []byte(""), 0655); err != nil {
				return errors.New("清空历史命令失败: " + err.Error())
			}
			return nil
		}

		f, err := os.OpenFile(historyFile, os.O_RDONLY, 0655)
		if err != nil {
			return errors.New("打开命令历史文件失败: " + err.Error())
		}
		defer f.Close()

		bufReader := bufio.NewReader(f)
		index := 0
		for {
			index += 1
			line, _, err := bufReader.ReadLine()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return errors.New("读取命令历史内容失败: " + err.Error())
			}

			_, _ = c.App.Println(index, "\t", string(line))

		}
	},
}

// func initHistoryCmd() {
// 	current.AddCommand(&grumble.Command{
// 		Name:     "history",
// 		Help:     "查看历史命令",
// 		LongHelp: "查看在本应用内使用过的命令",
// 		Flags: func(f *grumble.Flags) {
// 			f.BoolL("clear", false, "清除历史命令")
// 		},
// 		Run: func(c *grumble.Context) error {
// 			isClear := c.Flags.Bool("clear")
// 			if isClear {
// 				if err := ioutil.WriteFile(historyFile, []byte(""), 0655); err != nil {
// 					return errors.New("清空历史命令失败: " + err.Error())
// 				}
// 				return nil
// 			}

// 			f, err := os.OpenFile(historyFile, os.O_RDONLY, 0655)
// 			if err != nil {
// 				return errors.New("打开命令历史文件失败: " + err.Error())
// 			}
// 			defer f.Close()

// 			bufReader := bufio.NewReader(f)
// 			index := 0
// 			for {
// 				index += 1
// 				line, _, err := bufReader.ReadLine()
// 				if err == io.EOF {
// 					return nil
// 				}

// 				if err != nil {
// 					return errors.New("读取命令历史内容失败: " + err.Error())
// 				}

// 				_, _ = c.App.Println(index, "\t", string(line))

// 			}
// 		},
// 	})
// }
