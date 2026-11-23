package accesskratosslog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/stretchr/testify/require"
	"github.com/yylego/kratos-access-slog/accesskratosslog"
)

// captureLog runs one request and returns the log fields without latency
// It checks that the request result and error come back unchanged
//
// captureLog 执行一次请求，返回去掉 latency 的日志字段
// 同时检查请求结果和错误是否原样返回
func captureLog(t *testing.T, ctx context.Context, requestErr error, options ...accesskratosslog.Option) map[string]any {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	handleRequest := accesskratosslog.Server(logger, options...)(func(ctx context.Context, req any) (any, error) {
		return "result", requestErr
	})
	result, err := handleRequest(ctx, "abc")
	require.Equal(t, "result", result)
	require.Equal(t, requestErr, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &fields))
	require.IsType(t, float64(0), fields["latency"])
	delete(fields, "latency")
	return fields
}

func TestServerLogOutput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		require.Equal(t, map[string]any{
			"level": "INFO", "msg": "server request",
			"kind": "server", "component": "", "operation": "", "args": "abc", "code": float64(200), "reason": "",
		}, captureLog(t, context.Background(), nil))
	})

	t.Run("rejected_request", func(t *testing.T) {
		cause := errors.New(401, "TOKEN_EXPIRED", "token expired")
		require.Equal(t, map[string]any{
			"level": "WARN", "msg": "server request",
			"kind": "server", "component": "", "operation": "", "args": "abc", "code": float64(401), "reason": "TOKEN_EXPIRED",
			"error": cause.Error(),
		}, captureLog(t, context.Background(), cause))
	})

	t.Run("failed_request", func(t *testing.T) {
		cause := errors.New(500, "DB_ERROR", "select fails")
		require.Equal(t, map[string]any{
			"level": "ERROR", "msg": "server request",
			"kind": "server", "component": "", "operation": "", "args": "abc", "code": float64(500), "reason": "DB_ERROR",
			"error": cause.Error(), "stack": fmt.Sprintf("%+v", cause),
		}, captureLog(t, context.Background(), cause))
	})

	t.Run("client_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		// Cancellation appears in the message without a cause chain, as when formatting with %v
		cause := errors.New(500, "SESSION_QUERY_FAILURE", "session lookup: context canceled")
		require.Equal(t, map[string]any{
			"level": "DEBUG", "msg": "server request",
			"kind": "server", "component": "", "operation": "", "args": "abc", "code": float64(500), "reason": "SESSION_QUERY_FAILURE",
			"error": cause.Error(),
		}, captureLog(t, ctx, cause))
	})
}

func TestServerWithAttrs(t *testing.T) {
	require.Equal(t, map[string]any{
		"level": "INFO", "msg": "server request",
		"kind": "server", "component": "", "operation": "", "args": "abc", "code": float64(200), "reason": "",
		"trace": "t-123",
	}, captureLog(t, context.Background(), nil, accesskratosslog.WithAttrs(func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("trace", "t-123")}
	})))
}
