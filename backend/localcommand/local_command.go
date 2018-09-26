package localcommand

import (
	//"os"
	//"os/exec"
	"syscall"
	"time"
  "fmt"
	//"unsafe"

  "github.com/iamacarpet/go-winpty"
	"github.com/pkg/errors"
)

const (
	DefaultCloseSignal  = syscall.SIGINT
	DefaultCloseTimeout = 10 * time.Second
)

type LocalCommand struct {
	command string
	argv    []string

	closeSignal  syscall.Signal
	closeTimeout time.Duration

	//cmd       *exec.Cmd
	// pty       *os.File
  pty       *winpty.WinPTY
	ptyClosed chan struct{}
}

func New(command string, argv []string, options ...Option) (*LocalCommand, error) {
	//cmd := exec.Command(command, argv...)

	// pty, err := pty.Start(cmd)
  fmt.Printf("new command %s" , command)
  fmt.Println("\n")
  new_command := command
  for i := 0; i < len(argv); i++ { // looping from 0 to the length of the array
        fmt.Printf("%d th element of a is %s\n", i, argv[i])
        new_command = new_command + " " + argv[i]
  }
  
  pty, err := winpty.Open("", new_command)
  fmt.Printf("full command list %s" , new_command)
	if err != nil {
		// todo close cmd?
		return nil, errors.Wrapf(err, "failed to start command `%s`", command)
	}
	ptyClosed := make(chan struct{})

	lcmd := &LocalCommand{
		command: command,
		argv:    argv,

		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,

		//cmd:       cmd,
		pty:       pty,
		ptyClosed: ptyClosed,
	}

	for _, option := range options {
		option(lcmd)
	}

	// When the process is closed by the user,
	// close pty so that Read() on the pty breaks with an EOF.
	/*go func() {
		defer func() {
			lcmd.pty.Close()
			close(lcmd.ptyClosed)
		}()

		//lcmd.cmd.Wait()
	}()*/

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	//return lcmd.pty.Read(p)
  return lcmd.pty.StdOut.Read(p)
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	//return lcmd.pty.Write(p)
  return lcmd.pty.StdIn.Write(p)
}

func (lcmd *LocalCommand) Close() error {
  lcmd.pty.Close()
  return nil
	/*if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		lcmd.cmd.Process.Signal(lcmd.closeSignal)
	}
	for {
		select {
		case <-lcmd.ptyClosed:
			return nil
		case <-lcmd.closeTimeoutC():
			lcmd.cmd.Process.Signal(syscall.SIGKILL)
		}
	}*/
}

func (lcmd *LocalCommand) WindowTitleVariables() map[string]interface{} {
	return map[string]interface{}{
		"command": lcmd.command,
		"argv":    lcmd.argv,
		"pid":     "-1", //lcmd.cmd.Process.Pid,
	}
}

func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	/*window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	} else {
		return nil
	}*/
  lcmd.pty.SetSize(uint32(width), uint32(height))
  return nil
}

func (lcmd *LocalCommand) closeTimeoutC() <-chan time.Time {
	if lcmd.closeTimeout >= 0 {
		return time.After(lcmd.closeTimeout)
	}

	return make(chan time.Time)
}
