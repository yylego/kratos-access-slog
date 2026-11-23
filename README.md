<!-- TEMPLATE (EN) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-access-slog/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-access-slog/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-access-slog)](https://pkg.go.dev/github.com/yylego/kratos-access-slog)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-access-slog/main.svg)](https://coveralls.io/github/yylego/kratos-access-slog?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.26%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-access-slog.svg)](https://github.com/yylego/kratos-access-slog/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-access-slog)](https://goreportcard.com/report/github.com/yylego/kratos-access-slog)

<!-- TEMPLATE (EN) CLOSE: BADGES -->

# kratos-access-slog

Request logging middleware of Kratos, used in place of `logging.Server`, with DEBUG on client cancellation and WARN on rejected requests.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## Main Features

- 📝 **Same Log Shape**: Same message, fields and field sequence as `logging.Server`, so switching keeps the existing log queries working
- 🔕 **Client Cancellation at DEBUG**: A request the client canceled (page closed, switched to a different page) logs at DEBUG with no stack
- ⚠️ **4xx at WARN**: A rejected request (invalid input, expired token) logs at WARN with no stack
- 🚨 **Remaining Failures at ERROR**: Failed requests outside the cancellation and 4xx cases log at ERROR with stack, the same as `logging.Server`
- 🧩 **Extra Fields**: `WithAttrs` appends fields such as the trace ID to each log

## Installation

```bash
go get github.com/yylego/kratos-access-slog
```

## Design Intent

`logging.Server` of Kratos logs each failed request at ERROR with stack text; this choice cannot be configured. When a client closes the page during a request, net/http cancels the request context, the request function returns `context canceled`, and that request becomes an ERROR. Rejected requests such as an expired token become ERROR too. The service is not at fault in those cases, yet those logs hide the true faults.

This middleware follows `logging.Server` with these logging adjustments:

| Request                                | `logging.Server` | `accesskratosslog.Server` |
| -------------------------------------- | ---------------- | ------------------------- |
| No error                               | INFO             | INFO                      |
| Failed request with a canceled context | ERROR with stack | DEBUG                     |
| 4xx                                    | ERROR with stack | WARN                      |
| Remaining failures                     | ERROR with stack | ERROR with stack          |

Client cancellation is judged on the request context, not on the returned error: endpoint functions often wrap the cause with `%v`, which breaks the chain of causes, yet net/http cancels the request context in that case too.

## Usage

```go
import (
    "context"
    "log/slog"

    "github.com/go-kratos/kratos/v3/transport/http"
    "github.com/yylego/kratos-access-slog/accesskratosslog"
    "github.com/yylego/kratos-trace/tracekratos"
)

func newHTTPServer(logger *slog.Logger) *http.Server {
    return http.NewServer(
        http.Middleware(
            accesskratosslog.Server(logger, accesskratosslog.WithAttrs(func(ctx context.Context) []slog.Attr {
                if traceID := tracekratos.GetTraceID(ctx); traceID != "" {
                    return []slog.Attr{slog.String("trace", traceID)}
                }
                return nil
            })),
        ),
    )
}
```

Passing nil as the logger means `slog.Default()`, the same as `logging.Server`. The result and error of the endpoint function pass through unchanged.

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->
