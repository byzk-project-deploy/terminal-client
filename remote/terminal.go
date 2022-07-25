package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/stdio"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Terminal struct {
	user     string
	password string
	addr     string
	sshCli   *ssh.Client
	session  *ssh.Session

	stdout io.Reader
	stdin  io.Writer
	stderr io.Reader

	exitCh chan struct{}
}

func New(user, password, addr string) (*Terminal, error) {
	terminal := &Terminal{
		user:     user,
		password: password,
		addr:     addr,
	}

	if err := terminal.init(); err != nil {
		terminal.Close()
		return nil, err
	}

	terminal.exitCh = make(chan struct{}, 1)
	return terminal, nil
}

func (t *Terminal) Close() {
	defer func() {
		recover()
	}()

	if t.session != nil {
		t.session.Close()
	}

	if t.sshCli != nil {
		t.sshCli.Close()
	}

	close(t.exitCh)

}

func (t *Terminal) init() error {
	var err error
	sshConfig := &ssh.ClientConfig{
		User: t.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(t.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	t.sshCli, err = ssh.Dial("tcp", t.addr, sshConfig)
	if err != nil {
		return fmt.Errorf("创建命令转发通道失败: %s", err.Error())
	}

	t.session, err = t.sshCli.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %s", err.Error())
	}
	return nil
}

func (t *Terminal) updateTerminalSize() {

	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has
		// changed.
		sigwinchCh := make(chan os.Signal, 1)
		defer func() {
			signal.Stop(sigwinchCh)
			// 	close(sigwinchCh)
		}()
		signal.Notify(sigwinchCh, syscall.SIGWINCH)

		fd := int(os.Stdin.Fd())
		termWidth, termHeight, err := term.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		for {
			select {
			// The client updated the size of the local PTY. This change needs to occur
			// on the server side PTY as well.
			case sigwinch := <-sigwinchCh:
				if sigwinch == nil {
					return
				}
				currTermWidth, currTermHeight, err := term.GetSize(fd)

				// Terminal size has not changed, don't do anything.
				if currTermHeight == termHeight && currTermWidth == termWidth {
					continue
				}

				t.session.WindowChange(currTermHeight, currTermWidth)
				if err != nil {
					fmt.Printf("Unable to send window-change reqest: %s.", err)
					continue
				}

				termWidth, termHeight = currTermWidth, currTermHeight
			case <-t.exitCh:
				// fmt.Println("退出完事")
				return
			}
		}
	}()

}

func (t *Terminal) Run(cmd string) error {
	return t.RunWithOption(cmd, &serverclientcommon.CommandRunOption{
		WorkDir: server.CurrentPath(),
	})
}

func (t *Terminal) RunWithOption(cmd string, option *serverclientcommon.CommandRunOption) error {
	fd := int(os.Stdin.Fd())
	// make raw
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("转换原始终端失败: %s", err.Error())
	}
	defer term.Restore(fd, state)

	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		return err
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	err = t.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	t.updateTerminalSize()

	t.stdin, err = t.session.StdinPipe()
	if err != nil {
		return err
	}
	t.stdout, err = t.session.StdoutPipe()
	if err != nil {
		return err
	}
	t.stderr, err = t.session.StderrPipe()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go io.Copy(os.Stderr, t.stderr)
	go io.Copy(os.Stdout, t.stdout)
	go func() {
		buf := make([]byte, 128)
		for {

			n, err := stdio.Stdin.ReadWithContext(buf, ctx)
			if err == io.EOF {
				return
			}

			if err != nil {
				fmt.Println(err)
				return
			}

			if n > 0 {
				_, err = t.stdin.Write(buf[:n])
				if err != nil {
					return
				}
			}

		}
	}()

	defer func() {
		t.exitCh <- struct{}{}
	}()

	option.SystemCallOptionMarshal(t.session)

	t.session.Setenv("a", "")
	err = t.session.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}
