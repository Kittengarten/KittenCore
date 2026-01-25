package retry

// Error 可重试错误接口
type Error interface {
	error
	Retryable() bool
}

// 可重试错误
type retryableErr struct {
	error
	retryable bool
}

// Unwrap 实现包装错误，返回底层错误
func (e *retryableErr) Unwrap() error { return e.error }

// Retryable 是否可重试
func (e *retryableErr) Retryable() bool {
	return e.retryable
}

// NewError 创建一个可重试错误
func NewError(err error, retryable bool) error {
	if err == nil {
		return nil
	}
	return &retryableErr{error: err, retryable: retryable}
}
