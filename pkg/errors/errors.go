package errors

type StatusCoder interface {
	StatusCode() int
}

type SafeMassager interface {
	SafeMessage() string
}
