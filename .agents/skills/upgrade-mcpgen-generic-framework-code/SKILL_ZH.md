---
name: upgrade-mcpgen-generic-framework-code
description: '当 mcpgen-go 模板升级后，将其生成的新框架代码移植到当前项目。只改 Makefile、go.mod、internal/helpers/*、internal/mcpserver/*、main.go 等通用框架文件，不改 internal/mcptools/* 业务工具定义。'
argument-hint: '来自 sibling 项目的 git diff（mcpgen-go 升级前后的差异）'
---

# mcpgen-go Framework Upgrade

## 目标

将 mcpgen-go 通用模板生成器升级后产生的新框架代码，从参考 diff 移植到当前项目。**只迁移通用框架部分，不修改业务 tools 定义。**

## 当前项目固定约束（不可变）

这些是当前项目的固有特征，升级时 **必须保持不动**，不能用 diff 中的内容覆盖：

### 二进制名称
- Binary name: `sonarqube-mcp`
- go.mod module path: `sonarqube-mcp`
- 所有代码中 hardcoded 名称（如 `LoadConfig("sonarqube-mcp")`、`printDefaultConfigYAML()` 输出）必须使用 `sonarqube-mcp`，不可照抄 diff 中的 reference project 名称

### 业务 Tools — 绝对不可修改
- `internal/mcptools/*.go` — 所有 34 个 tools 定义文件
- `internal/mcptools/aggregate/*.go` — 自定义 `agg_*` 聚合工具
- `internal/mcptools/registry.go` — 工具注册表（名称映射是项目特有的）
- `internal/mcpcli/cli.go` — CLI 参数解析（GNU-style 已就绪）

### 项目特有 Client — 不可删除或被覆盖
- `internal/helpers/client.go` 中的 `SQClient` struct 及其方法（`DoGet`、`DoPost`、`DoPostWithBody`、`DoGetRaw`、`doRequest`）是 SonarQube 专用 API client，不可替换为 diff 中的 `ForwardRequest` 等通用 upstream proxy 逻辑
- 但 `client.go` 中可以 **新增** 通用工具函数（如 `LoadConfig`、`FormatAuthorizationHeader`），只要不删改现有 SQClient 相关代码

### 项目特有 Config — 不可修改
- `internal/helpers/config.go` — SonarQube 配置（`SONARQUBE_URL`、`SONARQUBE_TOKEN`、`SONARQUBE_ORG`、`SONARQUBE_PROJECT_KEY`、`SONARQUBE_TOOLSETS`、`SONARQUBE_READ_ONLY`、`IsCloud()`、`ToolSetEnabled()` 等）
- `internal/helpers/request_log.go` — 请求日志（`LogRequest`、`LogResponse`、verbosity 控制）
- `internal/helpers/token_*.go` — token 解析（keychain / wincred / file）

### 项目特有 main.go 逻辑 — 不可删除
- SonarQube URL/Cloud 启动检测和日志输出（`SONARQUBE_URL` 检查、`IsCloud()` 分支）
- `SONARQUBE_*` 环境变量说明（usage 中的 Environment 部分）
- CLI mode 的 `list` 子命令和 GNU-style 工具调用逻辑

---

## 框架文件清单（diff 可能涉及的文件）

| 文件 | 策略 |
|------|------|
| `Makefile` | 通用构建逻辑可整体替换，`BINARY_NAME` 须为 `sonarqube-mcp` |
| `go.mod` | 依赖变更可移植，module path 保持 `sonarqube-mcp`，indirect deps 由 `go mod tidy` 自动管理 |
| `internal/helpers/client.go` | **仅新增**通用函数/类型到文件末尾，不删改现有 `SQClient` 代码 |
| `internal/mcpserver/server.go` | 注册逻辑可增强（如 config 过滤），但已有的 Cloud/Server 条件注册必须保留 |
| `main.go` | 通用 flag/logging/SSE 增强可移植，但 SonarQube 特有启动逻辑必须保留 |
| `internal/mcpcli/cli.go` | 通常无需改动（GNU-style 已就绪） |

---

## 本次升级的具体变更

以下是本次需要移植的 diff 内容（由用户提供，每次升级时变化）：

{USER_INPUT}

---

## 移植流程

1. **读取当前项目文件** — 先 Read 每个将被修改的文件，理解现有结构
2. **分析 diff** — 区分：
   - **通用变更**：Makefile 构建系统、go.mod 依赖、`internal/helpers/client.go` 新增函数、`internal/mcpserver/server.go` 注册逻辑增强、`main.go` 通用 flag/logging/transport 增强
   - **项目特有变更**（需跳过）：工具名称映射、upstream endpoint 逻辑、业务 API client 替换
3. **逐文件移植**，关键替换规则：
   - diff 中的 reference binary name → `sonarqube-mcp`
   - diff 中的 reference module path → `sonarqube-mcp`
   - `BINARY_NAME` / `LoadConfig("...")` / `printDefaultConfigYAML()` 中的名称 → `sonarqube-mcp`
   - 遇到 diff 要删除/替换 `SQClient` 方法 → **跳过该条变更**
   - 遇到 diff 要修改 `internal/mcptools/` 下文件 → **跳过该条变更**
   - 遇到 diff 要修改 config / token / request_log → **跳过该条变更**
4. **验证**：
   ```bash
   go mod tidy
   go build ./...
   go vet ./...
   go test ./...
   ```
   确保所有 34 个 tools + agg_* 工具编译通过、测试通过。

## 跳过规则速查

以下内容出现在 diff 中时 **必须跳过、不做移植**：
- `internal/mcptools/` 路径下的任何变更
- `internal/helpers/config.go` 中的定制变更
- 删除/替换 `SQClient` 结构体或其方法的变更
- 删除 SonarQube 相关 env var 检查的变更
- registry.go 中工具名称映射的变更
