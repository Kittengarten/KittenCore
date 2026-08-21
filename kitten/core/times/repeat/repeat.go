package repeat

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/times"
)

// Config 重复执行配置，不提供默认配置
type Config struct {
	n        int           // 重复执行次数
	shortest time.Duration // 最小执行间隔
	longest  time.Duration // 最大执行间隔
}

// New 创建重复执行配置
func New(n int, shortest, longest time.Duration) Config {
	return Config{
		n:        n,
		shortest: shortest,
		longest:  longest,
	}
}

// Do 重复执行操作
func Do(ctx context.Context, cfg Config, fn func() error) error {
	cfg = Config{
		n:        max(0, cfg.n),
		shortest: max(0, cfg.shortest),
		longest:  max(0, cfg.longest),
	}
	t := time.NewTimer(0) // Go 1.23+，立即 Reset 不会残留一次信号
	for i := range cfg.n {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := fn(); err != nil {
			return fmt.Errorf(`停止执行喵！执行第 %d 次错误：%w`, i+1, err)
		}
		t.Reset(times.RandDurationRange(cfg.shortest, cfg.longest))
		select {
		case <-t.C:
			slog.Debug(`执行`, slog.Int(`次数`, i+1))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// IterS 重复执行操作（遍历切片）
func IterS[E any](ctx context.Context, cfg Config, s []E, fn func(i int, e E) error) error {
	cfg.shortest = max(0, cfg.shortest)
	cfg.longest = max(0, cfg.longest)
	t := time.NewTimer(0) // Go 1.23+，立即 Reset 不会残留一次信号
	for i, e := range s {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := fn(i, e); err != nil {
			return fmt.Errorf(`停止遍历喵！遍历第 %d 个元素错误：%w`, i+1, err)
		}
		t.Reset(times.RandDurationRange(cfg.shortest, cfg.longest))
		select {
		case <-t.C:
			slog.Debug(`遍历`, slog.Int(`下标`, i), slog.Any(`元素`, e))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// IterM 重复执行操作（遍历映射）
func IterM[K comparable, V any](ctx context.Context, cfg Config, m map[K]V, fn func(k K, v V) error) error {
	cfg.shortest = max(0, cfg.shortest)
	cfg.longest = max(0, cfg.longest)
	t := time.NewTimer(0) // Go 1.23+，立即 Reset 不会残留一次信号
	for k, v := range m {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := fn(k, v); err != nil {
			return fmt.Errorf(`停止遍历喵！遍历元素 %v 错误：%w`, k, err)
		}
		t.Reset(times.RandDurationRange(cfg.shortest, cfg.longest))
		select {
		case <-t.C:
			slog.Debug(`遍历`, slog.Any(`键`, k), slog.Any(`值`, v))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
