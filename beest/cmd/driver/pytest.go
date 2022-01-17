//go:build !windows

package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	DefaultPyTestBinary      = "py.test"
	DefaultPyTestWatchBinary = "ptw"

	SshIdentityFileFlag  = "--ssh-identity-file"
	AnsibleInventoryFlag = "--ansible-inventory"
	HostsFlag            = "--hosts"
	ConnectionFlag       = "--connection"

	ExpressionFlag = "-k"
)

type PytestContext interface {
	ConnectionContext
	TestsPath() string
	Hostnames() []string
}

func PyTestRun(ctx PytestContext, watch bool, selection string) error {
	cmd := buildPyTestCmd(ctx, watch, selection)
	log.Printf("Running Pytest cmd: %s", cmd)
	return runPyTestCmd(context.Background(), cmd)
}

func buildPyTestCmd(ctx PytestContext, watch bool, selection string) *exec.Cmd {
	var args []string

	if watch {
		args = append(args, "--")
	}

	args = append(args, "-vs") //TODO debug
	args = append(args, "-rap")

	args = append(args, fmt.Sprintf("%s=%s", SshIdentityFileFlag, ctx.PrivateKey()))
	args = append(args, fmt.Sprintf("%s=%s", ConnectionFlag, "ansible")) //Ansible module is only available with ansible connection backend
	args = append(args, fmt.Sprintf("%s=%s", AnsibleInventoryFlag, ctx.Inventory()))

	if len(ctx.Hostnames()) > 0 {
		var hosts []string
		for _, hostName := range ctx.Hostnames() {
			hosts = append(hosts, fmt.Sprintf("ansible://%s", hostName))
		}
		args = append(args, fmt.Sprintf("%s=%s", HostsFlag, strings.Join(hosts, ",")))
	}

	if selection != "" {
		args = append(args, fmt.Sprintf("%s %s", ExpressionFlag, selection))
	}

	var cmd *exec.Cmd
	if watch {
		cmd = exec.Command(DefaultPyTestWatchBinary, args...)
	} else {
		cmd = exec.Command(DefaultPyTestBinary, args...)
	}

	// TODO env variables
	//cmd.Env = tf.buildEnv(mergeEnv)
	cmd.Dir = ctx.TestsPath()

	return cmd
}

func runPyTestCmd(ctx context.Context, cmd *exec.Cmd) error {
	var errBuf strings.Builder

	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// kill children if parent is dead
		Pdeathsig: syscall.SIGKILL,
		// set process group ID
		Setpgid: true,
	}

	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			if cmd != nil && cmd.Process != nil && cmd.ProcessState != nil {
				// send SIGINT to process group
				err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
				if err != nil {
					log.Printf("Error from SIGINT: %s", err)
				}
			}

			// TODO: send a kill if it doesn't respond for a bit?
		}
	}()

	// check for early cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := cmd.Run()
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	if err != nil {
		return errors.New(errBuf.String())

		//TODO parse pytest errors
		//return tf.wrapExitError(ctx, err, errBuf.String())
	}

	return nil
}
