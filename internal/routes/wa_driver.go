package routes

type errSender struct{ reason string }

func (e errSender) SendOTP(string, string) error {
	return errWithPrefix("whatsapp belum siap: ", e.reason)
}

func errWithPrefix(prefix, msg string) error {
	return &prefixedError{prefix: prefix, msg: msg}
}

type prefixedError struct {
	prefix string
	msg    string
}

func (e *prefixedError) Error() string { return e.prefix + e.msg }
