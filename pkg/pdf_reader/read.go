package pdf_reader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func Read(ctx context.Context, path string) (ReadResponse, error) {
	if !filepath.IsAbs(path) {
		return ReadResponse{}, fmt.Errorf("path must be absolute")
	}

	fi, err := os.Lstat(path)
	if err != nil {
		return ReadResponse{}, fmt.Errorf("stat failed: %w", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return ReadResponse{}, fmt.Errorf("symlinks not allowed")
	}

	cmd := exec.CommandContext(
		ctx,
		"systemd-run",
		"--user",
		"--scope",
		"--quiet",
		"-p", "MemoryMax=100M",
		"-p", "MemoryHigh=90M",
		"-p", "CPUQuota=25%",
		"-p", "TasksMax=100",
		"/home/gosha/Documents/projects/essentia/build/pdf_reader",
		path,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	std, err := cmd.CombinedOutput()

	done := make(chan struct {
		out []byte
		err error
	}, 1)

	go func() {
		done <- struct {
			out []byte
			err error
		}{std, err}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			return ReadResponse{}, fmt.Errorf("process failed: %w; output: %s", res.err, string(res.out))
		}

		if len(res.out) == 0 {
			return ReadResponse{}, fmt.Errorf("empty output")
		}

		var resp ReadResponse
		err := json.Unmarshal(res.out, &resp)
		if err != nil {
			return ReadResponse{}, fmt.Errorf("failed to unmarshal reader output: %w; output: %s", err, string(res.out))
		}

		if len(resp.Error) != 0 {
			return ReadResponse{}, fmt.Errorf("reader error: %s", resp.Error)
		}

		return resp, nil

	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return ReadResponse{}, fmt.Errorf("timeout killed")
	}
}
