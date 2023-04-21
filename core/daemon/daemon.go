package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Run(ctx context.Context, root, path string, listen, controlPort string, onFinish func(error)) (*os.Process, error) {
	netConfigPath := root + "/global.config.json"
	_, err := os.Stat(netConfigPath)
	if os.IsNotExist(err) { // download network config if not exists
		resp, err := http.Get("https://ton.org/global.config.json")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		cfgData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(netConfigPath, cfgData, 0644)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	args := []string{"-v", "1", "-C", netConfigPath, "-I", listen, "-p", controlPort, "-D", root + "/storage-db"}

	log.Println("starting daemon with args:", strings.Join(args, " "))
	name := "storage-daemon"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	errLogs := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, path+"/"+name, args...)
	log.Println("command: ", cmd.String())

	cmd.SysProcAttr = daemonAttr()
	// cmd.Stdout = io.MultiWriter(os.Stdout, errLogs)
	// cmd.Stderr = io.MultiWriter(os.Stderr, errLogs)

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
			reason := errLogs.String()
			reason += " | Exit code: " + err.Error()
			err = errors.New(reason)
		}
		onFinish(err)
	}()

	return cmd.Process, nil
}
