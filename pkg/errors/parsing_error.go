package errors

import "errors"

type PipelineError string

const (
	ErrOpen            PipelineError = "open"
	ErrCorrupted       PipelineError = "corrupted"
	ErrEncrypted       PipelineError = "encrypted"
	ErrTimeout         PipelineError = "timeout"
	ErrExtract         PipelineError = "extract"
	ErrEmpty           PipelineError = "empty"
	ErrStorageDownload PipelineError = "storage_download"
	ErrStorageUpload   PipelineError = "storage_upload"
	ErrDB              PipelineError = "db"
	ErrUnknown         PipelineError = "unknown"
)

type PipeError struct {
	Err  error
	Code PipelineError
}

func NewPipeError(code PipelineError, err error) PipeError {
	if code == "" {
		code = ErrUnknown
	}

	return PipeError{
		Err:  err,
		Code: code,
	}
}

func (e PipeError) Error() string {
	return e.Err.Error()
}

func (e PipeError) Unwrap() error {
	return e.Err
}

func ToPipeError(err error) PipeError {
	if err == nil {
		return NewPipeError(ErrUnknown, nil)
	}

	if pipeError, ok := errors.AsType[*PipeError](err); ok {
		if pipeError.Code == "" {
			pipeError.Code = ErrUnknown
		}

		return *pipeError
	}

	return NewPipeError(ErrUnknown, err)
}
