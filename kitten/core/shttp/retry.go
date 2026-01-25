package shttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"

	"github.com/Kittengarten/KittenCore/kitten/core/times/retry"
)

// Retryable 实现 retry.Error 判断错误是否可重试
func (e *Error) Retryable() bool {
	if e.StatusCode < 400 {
		// 1xx, 2xx, 3xx 不重试
		return false
	}
	// 4xx 分情况重试
	switch e.StatusCode {
	case http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusTooManyRequests:
		return true
	case http.StatusConflict:
		return isIdempotentMethod(e.Method)
	}
	if e.StatusCode >= 500 {
		// 5xx 重试
		return true
	}
	// 4xx 其它不重试
	return false
}

// 判断方法是否幂等
func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

// CanRetryURLError 判断 *url.Error 是否可重试
func CanRetryURLError(err error) bool {
	if err == nil ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	uerr, ok := errors.AsType[*url.Error](err)
	if !ok {
		return false
	}
	switch {
	case uerr.Timeout(),
		errors.Is(uerr.Err, net.ErrClosed),
		isRetryableNetErr(uerr.Err):
		return true
	}
	if nerr, ok := errors.AsType[net.Error](uerr.Err); ok && nerr.Timeout() {
		return true
	}
	return false
}

func isRetryableNetErr(err error) bool {
	_, ok := errors.AsType[*net.OpError](err)
	if ok {
		return true
	}
	switch {
	case
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.ETIMEDOUT),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNABORTED),
		errors.Is(err, syscall.EPIPE):
		return true
	}
	return false
}

// Retry 使用重试机制执行 HTTP 请求
func Retry(ctx context.Context, c *http.Client, req *http.Request) (res *http.Response, err error) {
	err = retry.Do(ctx, retry.Default(), func() error {
		res, err = c.Do(req)
		return retry.NewError(err, CanRetryURLError(err))
	})
	return res, err
}
