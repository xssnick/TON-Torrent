package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

func Run() {
	args := strings.Split("-v 3 -C global.config.json -I :3333 -p 5555 -D storage-db", " ")

	cmd := exec.Command("./storage-daemon", args...)
	cmd.Dir = "/Users/xssnick/dev/ton/ton/build/storage/storage-daemon/"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing:", err)
		case *exec.ExitError:
			fmt.Println("command exit rc =", e.ExitCode())
		default:
			panic(err)
		}
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		cmd.Process.Kill()
	}()
}
