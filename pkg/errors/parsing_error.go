package errors

type ParsingErrorCode string

const (
	ParsingErrOpen            ParsingErrorCode = "open"
	ParsingErrCorrupted       ParsingErrorCode = "corrupted"
	ParsingErrEncrypted       ParsingErrorCode = "encrypted"
	ParsingErrTimeout         ParsingErrorCode = "timeout"
	ParsingErrExtract         ParsingErrorCode = "extract"
	ParsingErrStorageDownload ParsingErrorCode = "storage_download"
	ParsingErrStorageUpload   ParsingErrorCode = "storage_upload"
	ParsingErrDB              ParsingErrorCode = "db"
	ParsingErrUnknown         ParsingErrorCode = "unknown"
)

type ParsingError struct {
	Err  error
	Code ParsingErrorCode
}

func NewParsingError(code ParsingErrorCode, err error) *ParsingError {
	if code == "" {
		code = ParsingErrUnknown
	}

	return &ParsingError{
		Err:  err,
		Code: code,
	}
}

func (e *ParsingError) Error() string {
	if e == nil || e.Err == nil {
		return "parsing error"
	}

	return e.Err.Error()
}

func (e *ParsingError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}
