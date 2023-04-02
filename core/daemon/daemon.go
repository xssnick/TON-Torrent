package daemon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Run(root, path string, listen, controlPort string, onFinish func(error)) (*os.Process, error) {
	args := []string{"-v", "1", "-C", path + "/global.config.json", "-I", listen, "-p", controlPort, "-D", root + "/storage-db"}

	log.Println("starting daemon with args:", strings.Join(args, " "))
	name := "storage-daemon"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	errLogs := &bytes.Buffer{}

	cmd := exec.Command(path+"/"+name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, errLogs)

	if err := cmd.Start(); err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing:", err)
		case *exec.ExitError:
			fmt.Println("command exit rc =", e.ExitCode())
		default:
			return nil, err
		}
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			err = errors.New(err.Error() + "\n\n" + errLogs.String())
		}
		onFinish(err)
	}()

	return cmd.Process, nil
}
