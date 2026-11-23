<!-- TEMPLATE (ZH) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-access-slog/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-access-slog/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-access-slog)](https://pkg.go.dev/github.com/yylego/kratos-access-slog)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-access-slog/main.svg)](https://coveralls.io/github/yylego/kratos-access-slog?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.26%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-access-slog.svg)](https://github.com/yylego/kratos-access-slog/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-access-slog)](https://goreportcard.com/report/github.com/yylego/kratos-access-slog)

<!-- TEMPLATE (ZH) CLOSE: BADGES -->

# kratos-access-slog

Kratos 的请求日志中间件，用来替代 `logging.Server`，客户端取消记 DEBUG，被拒绝的请求记 WARN。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 核心特性

- 📝 **日志形态不变**：消息名、字段和字段顺序跟 `logging.Server` 一样，换过来以后原有的日志查询照常可用
- 🔕 **客户端取消记 DEBUG**：客户端取消的请求（关掉页面、切走）记 DEBUG，不带 stack
- ⚠️ **4xx 记 WARN**：被拒绝的请求（参数不合法、token 过期）记 WARN，不带 stack
- 🚨 **其余失败记 ERROR**：不属于取消或 4xx 情况的失败请求记 ERROR，带 stack，跟 `logging.Server` 一样
- 🧩 **额外字段**：`WithAttrs` 往每条日志追加字段，比如 trace 编号

## 安装

```bash
go get github.com/yylego/kratos-access-slog
```

## 设计意图

Kratos 的 `logging.Server` 对失败的请求记 ERROR 并附上 stack 文字，这条级别规则没法配置。客户端在请求途中关掉页面时，net/http 会取消请求上下文，处理函数返回 `context canceled`，这个请求就成了一条 ERROR。token 过期这类被拒绝的请求也一样记成 ERROR。这些都不是服务的故障，却把真故障淹没在里面。

这个中间件沿用 `logging.Server` 的写法，调整以下日志规则：

| 请求                   | `logging.Server` | `accesskratosslog.Server` |
| ---------------------- | ---------------- | ------------------------- |
| 没有错误               | INFO             | INFO                      |
| 请求失败且上下文已取消 | ERROR 带 stack   | DEBUG                     |
| 4xx                    | ERROR 带 stack   | WARN                      |
| 其余失败               | ERROR 带 stack   | ERROR 带 stack            |

客户端取消看的是请求上下文，不看返回的错误：处理函数常用 `%v` 包装原因，错误链会断，但 net/http 照样会取消请求上下文。

## 使用

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

logger 传 nil 时使用 `slog.Default()`，跟 `logging.Server` 一样。处理函数的结果和错误原样返回。

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->
