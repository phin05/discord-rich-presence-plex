package exc

const (
	codeInternal  = "INTERNAL_ERROR"
	codeMalformed = "PARSING_ERROR"
	codeInvalid   = "VALIDATION_ERROR"
)

type exc struct {
	err     error
	Code    string   `json:"code"`
	Message string   `json:"message,omitempty"`
	Details []string `json:"details,omitempty"`
}

func newExc(err error, code string, message string, details ...string) *exc {
	return &exc{
		err:     err,
		Code:    code,
		Message: message,
		Details: details,
	}
}

func (exc *exc) Error() string {
	return exc.Message
}

func internal(err error) *exc {
	return newExc(err, codeInternal, "")
}

func Malformed(message string, details ...string) *exc {
	return newExc(nil, codeMalformed, message, details...)
}

func Invalid(message string, details ...string) *exc {
	return newExc(nil, codeInvalid, message, details...)
}
