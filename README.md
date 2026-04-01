# dnsherene

这是一个用于 DNSHE 免费域名月度续期的工具，项目分层比较明确：

- `pkg/dnshe`：SDK 层，封装 API 客户端和领域模型。
- `pkg/notifier`：统一通知接口。
- `internal/app`：续期编排逻辑。
- `cmd/dnsherene`：CLI 入口。

## 项目结构

```text
cmd/dnsherene/main.go        # 应用启动与环境变量接线
internal/config/config.go    # 环境配置加载与校验
internal/app/service.go      # 续期执行流程
pkg/dnshe/client.go          # DNSHE API SDK
pkg/dnshe/types.go           # SDK 模型定义
pkg/notifier/*.go            # 通知接口与实现
```

## DNSHE SDK 覆盖接口

`pkg/dnshe` 当前已覆盖文档中的全部接口分组：

- `subdomains`
  - `ListSubdomains`
  - `RegisterSubdomain`
  - `GetSubdomain`
  - `DeleteSubdomain`
  - `RenewSubdomain`
- `dns_records`
  - `ListDNSRecords`
  - `CreateDNSRecord`
  - `UpdateDNSRecord`
  - `DeleteDNSRecord`
- `keys`
  - `ListAPIKeys`
  - `CreateAPIKey`
  - `DeleteAPIKey`
  - `RegenerateAPIKey`
- `quota`
  - `GetQuota`

### SDK 说明

- HTTP 错误和 `success=false` 的业务错误都会以 `*dnshe.APIError` 返回。
- 如果接口返回限流字段，`APIError` 会保留 `limit`、`remaining`、`reset_at` 等信息。
- `RenewSubdomain` 已包含文档中的续期返回字段，例如 `message`、`renewed_at`、`never_expires`、`status`。
- 创建 DNS 记录前会校验支持的记录类型：`A`、`AAAA`、`CNAME`、`MX`、`TXT`。

## 配置说明

必填项：

- `DNSHE_API_KEYS`
- `DNSHE_API_SECRETS`

凭证规则：

- 始终使用上面这两个复数环境变量。
- 单账号场景也按列表处理，只填 1 组即可。
- `DNSHE_API_KEYS` 和 `DNSHE_API_SECRETS` 的项目数量必须一致。
- 列表分隔符支持 `,`、`;` 和换行。

可选项（执行行为）：

- `DNSHE_DRY_RUN`：填 `true` 或 `1` 时只做演练，不发送真实续期请求。
- `DNSHE_API_BASE_URL`：默认值为 `https://api005.dnshe.com/index.php`。

可选项（统一通知）：

- `NOTIFY_WEBHOOK_URL`：设置后会以结构化 JSON 形式把通知发送到 webhook。
- `NOTIFY_WEBHOOK_TOKEN`：可选的 Bearer Token。

默认通知行为：

- 公共日志只输出所有 API 凭证合计的续期成功数量。
- 通知模块统一接收一个 `internal/report.Info` 结构体，内部包含账号数组；每个账号项都会带上匹配数量、续期数量、失败数量、脱敏 API Key，以及成功/失败域名列表。
- 每组 API 的详细通知仅在配置了 `NOTIFY_WEBHOOK_URL` 时发送到 webhook。
- 如果账号下没有可续期的子域名，程序会按空操作成功处理，不会作为失败退出。

## 隐私说明

- GitHub Actions 的公开日志不会打印域名、到期时间、剩余天数、Webhook URL 或原始 API Key。
- 成功时的公开输出只有 `renewed_total=<number>`。
- 失败时会额外输出脱敏后的错误追踪，例如错误条数和分组后的失败原因摘要。
- 详细续期结果会通过 webhook 字段发送，包括：
  - 每组 API 凭证的成功数量
  - 每组 API 凭证的未成功数量
  - 已更新域名的新到期时间和剩余天数
  - 未更新域名列表及失败原因
  - 脱敏后的 API Key 标识

## 本地运行

```bash
DNSHE_API_KEYS="cfsd_xxx" \
DNSHE_API_SECRETS="yyy" \
go run ./cmd/dnsherene
```

调试模式：

```bash
DNSHE_API_KEYS="cfsd_xxx" \
DNSHE_API_SECRETS="yyy" \
DNSHERENEW_DEBUG=true \
go run ./cmd/dnsherene
```

开启 `DNSHERENEW_DEBUG=true` 后，详细通知事件会同步打印到标准输出，便于本地排查；默认模式下这些明细仍只通过 webhook 发送。

多组 API 凭证：

```bash
DNSHE_API_KEYS="key_1,key_2" \
DNSHE_API_SECRETS="secret_1,secret_2" \
go run ./cmd/dnsherene
```

演练模式：

```bash
DNSHE_API_KEYS="cfsd_xxx" \
DNSHE_API_SECRETS="yyy" \
DNSHE_DRY_RUN=true \
go run ./cmd/dnsherene
```

## GitHub Actions（月度执行）

工作流文件：`.github/workflows/monthly-renew.yml`

调度方式：

- `15 0 1 * *`（UTC），每月 1 日执行一次。
- 同时支持 `workflow_dispatch` 手动触发。

### 仓库设置

添加 secrets：

- `DNSHE_API_KEYS`
- `DNSHE_API_SECRETS`
- `NOTIFY_WEBHOOK_URL`（可选）
- `NOTIFY_WEBHOOK_TOKEN`（可选）

添加 variables（可选）：

- `DNSHE_API_BASE_URL`
