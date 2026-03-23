package pdf_reader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

func Read(ctx context.Context, path string) (ReadResponse, error) {
	cmd := exec.CommandContext(
		ctx,
		"cgexec",
		"-g", "memory,cpu:pdf-limit",
		"./pdf-reader",
		path,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ReadResponse{}, err
	}

	if err := cmd.Start(); err != nil {
		return ReadResponse{}, err
	}

	done := make(chan struct {
		out []byte
		err error
	}, 1)

	go func() {
		out, err := io.ReadAll(stdout)
		done <- struct {
			out []byte
			err error
		}{out, err}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			return ReadResponse{}, res.err
		}

		var resp ReadResponse
		err := json.Unmarshal(res.out, &resp)
		if err != nil {
			return ReadResponse{}, fmt.Errorf("failed to unmarshal reader output: %w", err)
		}

		if len(resp.Error) != 0 {
			return ReadResponse{}, fmt.Errorf("failed to unmarshal reader output: %s", resp.Error)
		}

		return resp, nil

	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return ReadResponse{}, fmt.Errorf("timeout killed")
	}
}
