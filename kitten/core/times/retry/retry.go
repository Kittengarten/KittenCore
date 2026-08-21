package retry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"
)

// Config 重试配置
type Config struct {
	n       int           // 最大重试次数
	base    time.Duration // 基础重试间隔
	longest time.Duration // 最大重试间隔
}

// 默认重试配置
var config = Config{
	n:       3,
	base:    100 * time.Millisecond,
	longest: time.Second,
}

// Default 获取默认重试配置
func Default() Config {
	return config
}

// New 创建重试配置
func New(n int, base, longest time.Duration) Config {
	return Config{
		n:       n,
		base:    base,
		longest: longest,
	}
}

// ErrMaxRetryN 达到最大重试次数
var ErrMaxRetryN = errors.New(`达到最大重试次数`)

// Do 执行重试操作
func Do(ctx context.Context, cfg Config, fn func() error) error {
	cfg = Config{
		n:       max(0, cfg.n),
		base:    max(0, cfg.base),
		longest: max(0, cfg.longest),
	}
	t := time.NewTimer(0) // Go 1.23+，立即 Reset 不会残留一次信号
	for i := range cfg.n {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err := fn()
		switch rerr, ok := err.(Error); {
		case ok && !rerr.Retryable():
			// 不能重试，返回
			return fmt.Errorf(`停止重试喵！执行第 %d 次错误：%w，该错误不能重试`, i+1, err)
		case i == cfg.n-1 && err != nil:
			// 最后一次失败，返回
			return fmt.Errorf("停止重试喵！%w\n%w：%d", err, ErrMaxRetryN, cfg.n)
		case err == nil:
			// 成功，返回
			return nil
		}
		// 非最后一次的失败，应当等待重试
		t.Reset(expBackoff(i, cfg.base, cfg.longest))
		select {
		case <-t.C:
			slog.Error(`重试`,
				slog.Any(`错误`, err),
				slog.Int(`次数`, i+1),
			)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// 指数退避算法计算等待时间，带随机抖动，次数 i 从 0 开始
func expBackoff(i int, base, longest time.Duration) time.Duration {
	d := min(base*(1<<i), longest)
	jitter := time.Duration(rand.N(d >> 1))
	return d>>1 + jitter
}
