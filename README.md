# aidbt - AI Database Tool

`aidbt` 是一个 Go 编写的命令行数据库助手。你可以用中文描述数据库需求，AI 会生成一条 SQL，工具在本地完成 SQL 分类、风险提示、强确认、执行和结果表格展示。

## 功能特性

- 自然语言生成 SQL：AI 返回固定 JSON，工具本地二次分类。
- 双入口：`aidbt` 进入 REPL，`aidbt "查最近10个用户"` 执行一次性查询。
- 多 AI provider：Claude、OpenAI、DeepSeek、Ollama、自定义 OpenAI 兼容接口。
- 多数据库：MySQL、Oracle、达梦。
- 安全确认：所有 SQL 执行前确认，高危 DDL 需要输入确认短语。
- Schema 上下文：连接后自动探测当前库/schema 的表结构摘要。
- 本地日志：记录问题、SQL、执行状态、耗时和影响行数，不记录查询结果。

## 构建

```powershell
go build ./cmd/aidbt
```

项目当前目标 Go 版本为 `1.25.0`。如果本机 Go 版本较低，Go toolchain 会按需下载新版本。

## 快速开始

```powershell
.\aidbt.exe init
.\aidbt.exe
```

或一次性执行：

```powershell
.\aidbt.exe "查询最近10个用户"
```

## CLI 命令

| 命令 | 说明 |
| --- | --- |
| `aidbt` | 进入 REPL |
| `aidbt "自然语言问题"` | 一次性生成并执行 SQL |
| `aidbt init` | 初始化 AI provider 和数据库连接 |
| `aidbt conn add` | 添加数据库连接 |
| `aidbt conn edit [name]` | 编辑数据库连接 |
| `aidbt conn delete [name]` | 删除数据库连接 |
| `aidbt conn list` | 列出连接，密码脱敏 |
| `aidbt conn test [name]` | 测试连接 |
| `aidbt config show` | 脱敏显示配置 |

## REPL 命令

| 命令 | 说明 |
| --- | --- |
| `/help` | 显示帮助 |
| `/reset` | 清空当前对话历史 |
| `/schema refresh` | 重新探测当前连接表结构 |
| `/exit` / `/quit` | 退出 |

## 配置

配置目录为 `.aidbt`：

- 便携模式：可执行文件同目录存在 `.aidbt/` 时使用该目录。
- 安装模式：否则使用用户主目录下的 `.aidbt/`。

示例：

```toml
default_provider = "openai"
default_connection = "dev"

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "<your-api-key>"
model = "gpt-4o"

[connections.dev]
type = "mysql"
host = "127.0.0.1"
port = 3306
username = "root"
password = "<your-db-password>"
database = "app"
schema = ""
include = ["users", "orders*"]
exclude = ["audit_*"]

[ui]
stream = true
color = "auto"
max_rows = 100
```

## 环境变量覆盖

| 环境变量 | 说明 |
| --- | --- |
| `AIDBT_PROVIDER` | 覆盖默认 AI provider |
| `AIDBT_CONNECTION` | 覆盖默认数据库连接 |
| `AIDBT_<PROVIDER>_API_KEY` | 覆盖 provider API Key |
| `AIDBT_<PROVIDER>_MODEL` | 覆盖 provider 模型 |
| `AIDBT_<PROVIDER>_BASE_URL` | 覆盖 provider Base URL |
| `AIDBT_<CONNECTION>_PASSWORD` | 覆盖连接密码 |
| `AIDBT_<CONNECTION>_DSN` | 覆盖连接 DSN |

## 数据库驱动说明

- MySQL：`github.com/go-sql-driver/mysql`
- Oracle：`github.com/sijms/go-ora/v2`，纯 Go，不依赖 Oracle Client/CGO。
- 达梦：`github.com/godoes/gorm-dameng/dm8`，基于达梦官方 Go 驱动源码整理，可通过 Go module 引入。

达梦原计划候选 `github.com/fengzehao/dm-go-driver` 当前版本要求 Go 1.25；项目已升级到 Go 1.25，但实现仍采用 `godoes/gorm-dameng/dm8` 以获得更明确的可模块化路径。

## 安全边界

- 本地 SQL classifier 不信任 AI 自报的 `statement_type` 和 `risk`。
- 首版禁止多语句一次执行。
- `DROP`、`TRUNCATE`、`ALTER` 归为 critical，需要输入 `确认执行`。
- `UPDATE`、`DELETE` 未检测到 `WHERE` 时归为 high。
- v1 不提供自动事务确认/回滚，不提供执行失败后的自动修正循环。

## License

Apache License 2.0
