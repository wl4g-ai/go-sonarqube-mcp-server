---
name: sonarqube-mcp-refactor-to-golang
description: '用于将内置 SonarQube MCP Java server 重构、迁移、翻译或重新实现为当前 Go/Golang sonarqube-mcp 项目。使用 ./resources/sonarqube-mcp-server-1.20.0.2929，并按 main.go、internal/helpers、internal/mcpserver、internal/mcptools 的现有结构产出惯用 Go 代码。'
argument-hint: '要迁移的 SonarQube MCP 功能、工具类别或 Java 类'
---

# SonarQube MCP 转 Go

## 目标

将内置 Java 版 SonarQube MCP Server 的行为迁移到当前 Go 项目，保留 `github.com/mark3labs/mcp-go` 与现有目录风格，并逐步替换从 Cyberflow 模板复制来的占位逻辑。

源码目录在：

`./resources/sonarqube-mcp-server-1.19.0.2785/`

动手前先读：

`./references/source-map.md`

## 基本原则

- 迁移行为和公开契约，不照搬 Java 写法，也不要复制大段 Java 代码或 license header。
- 现有 Checkmarx、Netsparker、scan、Cyberflow 风格工具默认视为脚手架，除非用户明确要求保留。
- 优先沿用当前 Go 结构：`main.go`、`internal/helpers`、`internal/mcpserver`、`internal/mcptools`。
- 新增抽象必须有实际价值：减少重复、隔离 SonarQube 客户端逻辑，或匹配已有项目风格。

## 目标结构

- `main.go`：启动参数、stdio/http transport、`/mcp`、日志、优雅退出。
- `internal/helpers`：配置、token、header 转发、URL 构造、请求/响应处理、SonarQube client 公共逻辑。
- `internal/mcpserver/server.go`：MCP server、middleware、工具注册。
- `internal/mcptools`：每个工具单独文件，提供 `New<Name>MCPTool` 和 `<Name>Handler`。

## 流程

1. **明确范围**

   整体迁移必须垂直切片推进；指定功能、类别或类名时，只迁移切片和必要公共依赖；范围不清时，从配置、HTTP client、server API helper、后端无关误报工具开始。

2. **读取上下文**

   先看 Go 目标结构，再看资源目录中对应 Java 类和测试；行为、边界条件、schema 不明确时，以 Java 测试为准。

3. **先建映射再改代码**

   配置/env 映射到 Go helper；Java API wrapper 映射到 Go request/client；Java `Tool` 映射到 Go schema、handler、注册代码。

4. **按 Go 风格实现**

   复用现有 helper、`mcp.NewToolWithRawSchema`、日志和响应处理；

   优先使用：

   - `SONARQUBE_URL`
   - `SONARQUBE_TOKEN`
   - `SONARQUBE_ORG`
   - `SONARQUBE_TOOLSETS`
   - `SONARQUBE_READ_ONLY`
   - `SONARQUBE_PROJECT_KEY`

   等 SonarQube 契约。

5. **注册与清理**

   工具编译通过后再加入到 `internal/mcpserver/server.go`；

   只有等价 SonarQube 工具完成或用户确认后，才移除旧占位工具；

   用户可见命令、env、工具列表变化时同步文档。

6. **验证**

   运行：

   ```bash
   gofmt
   go test ./...
   go build ./...
   ```

   或：

   ```bash
   make
   ```

   transport 变更要 smoke test stdio 与 HTTP `/mcp`；

   工具变更要覆盖 URL、必填参数、默认 project key、token/header 转发、只读限制和错误响应。

## 决策点

- Java 中 Jetty、Reactor、Servlet、Apache async HTTP 等框架细节，只迁移观察行为到 Go/http 与 `mcp-go`。
- IDE bridge、local analyzer、plugin sync、proxied MCP/CAG、telemetry、HTTPS/auth 属于大模块，核心 API 工具稳定后再分阶段迁移。
- mutation 工具必须先检查 `SONARQUBE_READ_ONLY`。
- SonarQube Cloud 与自托管差异要显式处理 URL、org、token。
- 没有现成 schema 时，从 Java 工具定义和测试推导最小 JSON schema，保持工具名和描述稳定。

## 完成标准

- 名称和产物统一为 `sonarqube-mcp`。
- **当前项目存量代码是示例 Confluence Swagger to MCP 的脚手架（Confluence page/space/comment/attachment/label 等工具），必须全部删除，最终只保留 SonarQube MCP 的代码。** 包括但不限于：`Addlabels`、`Createchildpage`、`Createcomment`、`Createpage`、`Deletepage`、`Downloadattachment`、`GetAVeryLongOperationId...`、`Getspacecontent`、`Listspaces`、`Removelabel`、`Searchcontent`、`Updatepage`、`Updatepagebody`、`Uploadattachment` 等所有 Confluence 占位工具，以及 `MCP_UPSTREAM_ENDPOINT` / `MCP_UPSTREAM_TOKEN` / Confluence REST API 相关的 upstream 转发逻辑。
- 除非刻意保留，用户可见内容不再残留 Checkmarx、Netsparker、scan、Cyberflow、Confluence 占位痕迹。
- stdio 与 HTTP transport 仍可用，变动不会被普通日志打扰。
- 修改过的 Go 文件已格式化、可编译，并有测试或 smoke check。
- 最终说明已迁移内容、暂缓内容和通过的验证命令。
