// Package accesskratosslog is the request log middleware of Kratos, used in place of logging.Server
// It keeps the message, fields and field sequence of logging.Server, with two logging adjustments:
// a request the client canceled logs at DEBUG, and a 4xx logs at WARN without stack
//
// accesskratosslog 是 Kratos 的请求日志中间件，用来替代 logging.Server
// 消息名、字段和字段顺序跟 logging.Server 一样，只在日志级别上不同：
// 客户端取消的请求记 DEBUG，4xx 记 WARN 不带 stack
package accesskratosslog

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"github.com/go-kratos/kratos/v3/transport/http/status"
	"google.golang.org/grpc/codes"
)

// Redacter matches the interface in the logging package of Kratos
// A request implementing it decides the text written to the log
//
// Redacter 跟 Kratos logging 包的同名接口一致
// 请求实现了它，日志里就写它给出的文字
type Redacter interface {
	Redact() string
}

// Option sets up the middleware
//
// Option 配置中间件
type Option func(*config)

type config struct {
	attrsFunc func(ctx context.Context) []slog.Attr
}

// WithAttrs appends the fields that attrsFunc returns to each log, such as the trace ID
//
// WithAttrs 把 attrsFunc 返回的字段追加到每条日志，比如 trace 编号
func WithAttrs(attrsFunc func(ctx context.Context) []slog.Attr) Option {
	return func(c *config) {
		c.attrsFunc = attrsFunc
	}
}

// Server is the middleware that logs each request, used in place of logging.Server in Kratos
// Passing nil as the slog instance means slog.Default(), the same as Kratos
//
// Server 是服务端请求日志中间件，用来替代 Kratos 的 logging.Server
// logger 传 nil 时使用 slog.Default()，跟 Kratos 一样
func Server(logger *slog.Logger, options ...Option) middleware.Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	cfg := &config{}
	for _, option := range options {
		option(cfg)
	}
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (res any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)

			code = int32(status.FromGRPCCode(codes.OK))

			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			res, err = next(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractCause(ctx, err, code)
			attrs := []slog.Attr{
				slog.String("kind", "server"),
				slog.String("component", kind),
				slog.String("operation", operation),
				slog.String("args", extractArgs(req)),
				slog.Int64("code", int64(code)),
				slog.String("reason", reason),
				slog.Float64("latency", time.Since(startTime).Seconds()),
			}
			if cfg.attrsFunc != nil {
				attrs = append(attrs, cfg.attrsFunc(ctx)...)
			}
			if err != nil {
				attrs = append(attrs, slog.Any("error", err))
				if stack != "" {
					attrs = append(attrs, slog.String("stack", stack))
				}
			}
			logger.LogAttrs(ctx, level, "server request", attrs...)
			return
		}
	}
}

// extractArgs gives the text of the request, the same as the logging package of Kratos
//
// extractArgs 给出请求的文字，跟 Kratos logging 包一样
func extractArgs(req any) string {
	if v, ok := req.(Redacter); ok {
		return v.Redact()
	}
	if v, ok := req.(fmt.Stringer); ok {
		return v.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractCause chooses the slog.Level and stack text, matching Kratos logging except in two cases:
//   - the client canceled the request: DEBUG with no stack, since the service is not at fault
//   - a 4xx: WARN with no stack, since the service rejected the request and works as designed
//
// extractCause 给出日志级别和 stack 文字，跟 Kratos logging 包一样，只有两种情况不同：
//   - 客户端取消了请求：DEBUG 不带 stack，因为那不是服务的故障
//   - 4xx：WARN 不带 stack，因为请求被拒绝，服务本身正常
func extractCause(ctx context.Context, err error, code int32) (slog.Level, string) {
	if err != nil {
		if stderrors.Is(ctx.Err(), context.Canceled) {
			return slog.LevelDebug, ""
		}
		if code >= 400 && code < 500 {
			return slog.LevelWarn, ""
		}
		return slog.LevelError, fmt.Sprintf("%+v", err)
	}
	return slog.LevelInfo, ""
}
