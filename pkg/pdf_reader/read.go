package pdf_reader

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

func Read(ctx context.Context, path string) (ReadResponse, error) {
	if !filepath.IsAbs(path) {
		return ReadResponse{}, fmt.Errorf("path must be absolute")
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
		var resp ReadResponse
		if len(res.out) > 0 {
			err := json.Unmarshal(res.out, &resp)
			if err != nil {
				return ReadResponse{}, errs.NewParsingError(
					errs.ParsingErrUnknown,
					fmt.Errorf("failed to unmarshal reader output: %w; output: %s", err, string(res.out)),
				)
			}

			if len(resp.Error) != 0 {
				return ReadResponse{}, errs.NewParsingError(codeFromResponse(resp.ErrorCode), fmt.Errorf("reader error: %s", resp.Error))
			}
		}

		if res.err != nil {
			return ReadResponse{}, errs.NewParsingError(
				errs.ParsingErrUnknown,
				fmt.Errorf("process failed: %w; output: %s", res.err, string(res.out)),
			)
		}

		if len(res.out) == 0 {
			return ReadResponse{}, errs.NewParsingError(errs.ParsingErrUnknown, fmt.Errorf("empty output"))
		}

		return resp, nil

	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return ReadResponse{}, errs.NewParsingError(errs.ParsingErrTimeout, fmt.Errorf("timeout killed"))
	}
}

func codeFromResponse(code string) errs.ParsingErrorCode {
	switch errs.ParsingErrorCode(strings.ToLower(code)) {
	case errs.ParsingErrOpen,
		errs.ParsingErrCorrupted,
		errs.ParsingErrEncrypted,
		errs.ParsingErrTimeout,
		errs.ParsingErrExtract,
		errs.ParsingErrEmpty,
		errs.ParsingErrStorageDownload,
		errs.ParsingErrStorageUpload,
		errs.ParsingErrDB,
		errs.ParsingErrUnknown:
		return errs.ParsingErrorCode(strings.ToLower(code))
	default:
		return errs.ParsingErrUnknown
	}
}
