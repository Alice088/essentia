package pdf_parser

import (
	errs "Alice088/essentia/pkg/errors"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func (r *Parser) Parse(ctx context.Context) (ReadResponse, error) {
	timeout, cancel := context.WithTimeout(ctx, r.Config.ReaderContextTimeout)
	defer cancel()

	if !filepath.IsAbs(r.TMP.Path()) {
		return ReadResponse{}, fmt.Errorf("path must be absolute")
	}

	cmd := exec.CommandContext(
		timeout,
		"systemd-run",
		"--user",
		"--scope",
		"--quiet",
		"-p", "MemoryMax=100M",
		"-p", "MemoryHigh=90M",
		"-p", "CPUQuota=25%",
		"-p", "TasksMax=100",
		filepath.Join(".", "build", "pdf_reader"),
		r.TMP.Path(),
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
		out := struct {
			out []byte
			err error
		}{std, err}
		select {
		case done <- out:
		case <-timeout.Done():
			return
		}
	}()

	select {
	case res := <-done:
		var resp ReadResponse
		if len(res.out) > 0 {
			err := json.Unmarshal(res.out, &resp)
			if err != nil {
				return ReadResponse{}, errs.NewPipeError(
					errs.ErrUnknown,
					fmt.Errorf("failed to unmarshal parser output: %w", err),
				)
			}

			if len(resp.Error) != 0 {
				return ReadResponse{}, errs.NewPipeError(codeFromResponse(resp.ErrorCode), fmt.Errorf("reader error: %s", resp.Error))
			}
		}

		if res.err != nil {
			return ReadResponse{}, errs.NewPipeError(
				errs.ErrUnknown,
				fmt.Errorf("process failed: %w", res.err),
			)
		}

		if len(res.out) == 0 {
			return ReadResponse{}, errs.NewPipeError(errs.ErrUnknown, fmt.Errorf("empty output"))
		}

		return resp, nil

	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return ReadResponse{}, errs.NewPipeError(errs.ErrTimeout, fmt.Errorf("timeout killed"))
	}
}

func codeFromResponse(code string) errs.PipelineError {
	switch errs.PipelineError(strings.ToLower(code)) {
	case errs.ErrOpen,
		errs.ErrCorrupted,
		errs.ErrEncrypted,
		errs.ErrTimeout,
		errs.ErrExtract,
		errs.ErrEmpty,
		errs.ErrStorageDownload,
		errs.ErrStorageUpload,
		errs.ErrDB,
		errs.ErrUnknown:
		return errs.PipelineError(strings.ToLower(code))
	default:
		return errs.ErrUnknown
	}
}
