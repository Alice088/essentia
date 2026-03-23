package errors

// StatusCoder defines an interface for errors or objects that can provide
// an associated HTTP status code.
//
// It is typically used at the transport layer (e.g. HTTP handlers) to map
// domain or application errors to appropriate HTTP responses, without
// coupling business logic to net/http.
type StatusCoder interface {
	StatusCode() int
}

// SafeMassager defines an interface for errors or objects that can expose
// a sanitized, user-safe message.
//
// implement should return a string that is safe to show to end users,
// without leaking internal details, stack traces, or sensitive data.
type SafeMassager interface {
	SafeMessage() string
}
