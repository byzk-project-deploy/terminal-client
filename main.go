package main

import "github.com/byzk-project-deploy/terminal-client/cmd"

func main() {
	// for {
	cmd.Run()
	// }

	// t, err := tcell.NewDevTty()
	// if err != nil {
	// 	panic(err)
	// }
	// defer t.Stop()
	// defer t.Drain()
	// t.Start()
	// t.Write([]byte("你好亚"))
	// time.Sleep(10 * time.Second)

	// s, _ := tcell.NewTerminfoScreenFromTty(t)
	// defer func() {
	// 	s.Fini()
	// }()
	// s.Init()

	// t.Write([]byte("你好亚"))

	// time.Sleep(5 * time.Second)
	// app := tview.NewApplication()
	// form := tview.NewForm()

	// form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
	// 	if ev.Key() == tcell.KeyRune {
	// 		if ev.Rune() == 'q' {
	// 			app.Stop()
	// 			os.Exit(0)
	// 		}
	// 	}
	// 	return ev
	// })

	// form.AddButton("Edit with vim", func() {
	// 	app.Suspend(func() {
	// 		cmd := exec.Command("vim", "/tmp/foo")
	// 		cmd.Stdout = os.Stdout
	// 		cmd.Stdin = os.Stdin
	// 		_ = cmd.Run()
	// 	})
	// })

	// form.SetBorder(true).SetTitle("Launch vim & go back to main screen after exit, Press 'q' to exit")

	// if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
	// 	panic(err)
	// }
}
