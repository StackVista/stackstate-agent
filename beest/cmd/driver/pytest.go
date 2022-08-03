//go:build !windows

package driver

import (
	"beest/cmd/step"
	"beest/sut"
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

	AnsibleInventoryFlag = "--ansible-inventory"
	HostsFlag            = "--hosts"
	ConnectionFlag       = "--connection"

	ExpressionFlag = "-k"
)

type PyTestVerifier struct{}

func (pv *PyTestVerifier) Verify(step *step.VerificationStep, watch bool, selection string) error {
	cmd := buildPyTestCmd(step, watch, selection)
	log.Printf("Running Pytest cmd: %s", cmd)
	return runPyTestCmd(context.Background(), cmd)
}

func buildPyTestCmd(step *step.VerificationStep, watch bool, selection string) *exec.Cmd {
	var args []string

	if watch {
		args = append(args, "--")
	}

	args = append(args, "-rap")

	// Ansible module is only available with ansible connection backend
	args = append(args, fmt.Sprintf("%s=%s", ConnectionFlag, "ansible"))
	args = append(args, fmt.Sprintf("%s=%s", AnsibleInventoryFlag, step.Inventory()))

	if len(step.Hostnames()) > 0 {
		var hosts []string
		for _, hostName := range step.Hostnames() {
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

	cmd.Dir = step.TestsPath()

	pyPaths := sut.TestFrameworkPaths()
	if defPyPath := os.Getenv("PYTHONPATH"); defPyPath != "" {
		pyPaths = append(pyPaths, defPyPath)
	}
	pyPath := strings.Join(pyPaths, ":")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s", pyPath))

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
	}

	return nil
}
