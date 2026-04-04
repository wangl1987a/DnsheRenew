# 通知模块配置

本文档只说明通知模块的配置方式，不涉及 DNSHE 主流程配置。

当前通知配置统一由 `internal/config` 读取，通知模块本身不直接读取环境变量。

## 启用规则

每个通知渠道都遵循同一套规则：

- 某个渠道相关环境变量全部为空：视为未启用
- 某个渠道配置了一部分但不完整：程序启动时直接报错
- 某个渠道配置完整：视为启用

## 当前已实现的通知渠道

- Console
- Telegram
- Lark
- Webhook

## 已预留配置但尚未接入发送逻辑的渠道

- Mail

## 通用说明

多值环境变量使用与主配置一致的分隔规则，支持：

- `,`
- `;`
- 换行
- 制表符

目前会用到这个规则的字段：

- `DNSHE_NOTIFY_MAIL_TO`
- `DNSHE_NOTIFY_LARK_RECEIVERS`

## Console

控制台通知沿用调试开关。

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_DEBUG` | 是否启用控制台调试通知；支持 `1`、`true`、`yes`、`on` | 否 | `true` |

启用方式：

- `DNSHE_DEBUG=true`

说明：

- 当 `DNSHE_DEBUG` 为 `1`、`true`、`yes`、`on` 时启用
- 其他值视为未启用

最小示例：

```bash
DNSHE_DEBUG=true
```

## Telegram

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN` | Telegram Bot Token | 是 | `123456:abcdef` |
| `DNSHE_NOTIFY_TELEGRAM_CHAT_ID` | Telegram 目标聊天 ID | 是 | `-1001234567890` |
| `DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID` | Telegram topic / thread ID | 否 | `7` |
| `DNSHE_NOTIFY_TELEGRAM_PARSE_MODE` | 消息解析模式 | 否 | `HTML` |

约束：

- `DNSHE_NOTIFY_TELEGRAM_CHAT_ID` 必须是整数
- `DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID` 如果设置，必须是正整数

最小示例：

```bash
DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN=123456:abcdef
DNSHE_NOTIFY_TELEGRAM_CHAT_ID=-1001234567890
```

带 topic / parse mode 的示例：

```bash
DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN=123456:abcdef
DNSHE_NOTIFY_TELEGRAM_CHAT_ID=-1001234567890
DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID=7
DNSHE_NOTIFY_TELEGRAM_PARSE_MODE=HTML
```

## Mail

当前状态：

- 已完成配置解析和校验
- 尚未接入实际发送逻辑

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_NOTIFY_MAIL_FROM` | 发件人地址 | 是 | `bot@example.com` |
| `DNSHE_NOTIFY_MAIL_SMTP_HOST` | SMTP 主机地址 | 是 | `smtp.example.com` |
| `DNSHE_NOTIFY_MAIL_SMTP_PORT` | SMTP 端口 | 是 | `587` |
| `DNSHE_NOTIFY_MAIL_SMTP_IDENTITY` | SMTP identity，可为空 | 否 | `bot@example.com` |
| `DNSHE_NOTIFY_MAIL_SMTP_USERNAME` | SMTP 用户名 | 否 | `bot@example.com` |
| `DNSHE_NOTIFY_MAIL_SMTP_PASSWORD` | SMTP 密码 | 否 | `your-password` |
| `DNSHE_NOTIFY_MAIL_TO` | 收件人列表，支持多值分隔 | 是 | `alice@example.com,bob@example.com` |

约束：

- `DNSHE_NOTIFY_MAIL_SMTP_PORT` 必须是正整数
- 如果设置了认证相关字段中的任意一个，则 `DNSHE_NOTIFY_MAIL_SMTP_USERNAME` 和 `DNSHE_NOTIFY_MAIL_SMTP_PASSWORD` 必须同时存在
- `DNSHE_NOTIFY_MAIL_SMTP_IDENTITY` 可为空

最小示例：

```bash
DNSHE_NOTIFY_MAIL_FROM=bot@example.com
DNSHE_NOTIFY_MAIL_SMTP_HOST=smtp.example.com
DNSHE_NOTIFY_MAIL_SMTP_PORT=25
DNSHE_NOTIFY_MAIL_TO=alice@example.com,bob@example.com
```

带 SMTP 认证的示例：

```bash
DNSHE_NOTIFY_MAIL_FROM=bot@example.com
DNSHE_NOTIFY_MAIL_SMTP_HOST=smtp.example.com
DNSHE_NOTIFY_MAIL_SMTP_PORT=587
DNSHE_NOTIFY_MAIL_SMTP_USERNAME=bot@example.com
DNSHE_NOTIFY_MAIL_SMTP_PASSWORD=your-password
DNSHE_NOTIFY_MAIL_TO=alice@example.com,bob@example.com
```

## Lark

当前实现说明：

- 使用 `nikoksr/notify/service/lark`
- 发送内容为纯文本摘要，标题固定为 `DNSHE 续期摘要`

Lark 支持两种模式：

- webhook 模式
- custom app 模式

这两种模式不能混用。

### Lark Webhook

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_NOTIFY_LARK_WEBHOOK_URL` | Lark / 飞书群机器人 webhook 地址 | 是 | `https://open.feishu.cn/open-apis/bot/v2/hook/xxx` |

最小示例：

```bash
DNSHE_NOTIFY_LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxx
```

约束：

- 如果设置了 `DNSHE_NOTIFY_LARK_WEBHOOK_URL`，就不能再设置 custom app 相关字段

### Lark Custom App

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_NOTIFY_LARK_APP_ID` | Lark 自建应用 App ID | 是 | `cli_xxx` |
| `DNSHE_NOTIFY_LARK_APP_SECRET` | Lark 自建应用 App Secret | 是 | `yyy` |
| `DNSHE_NOTIFY_LARK_RECEIVER_TYPE` | 接收人 ID 类型 | 是 | `chat_id` |
| `DNSHE_NOTIFY_LARK_RECEIVERS` | 接收人列表，支持多值分隔 | 是 | `oc_xxx,oc_yyy` |

`DNSHE_NOTIFY_LARK_RECEIVER_TYPE` 当前支持：

- `open_id`
- `user_id`
- `union_id`
- `email`
- `chat_id`

最小示例：

```bash
DNSHE_NOTIFY_LARK_APP_ID=cli_xxx
DNSHE_NOTIFY_LARK_APP_SECRET=yyy
DNSHE_NOTIFY_LARK_RECEIVER_TYPE=chat_id
DNSHE_NOTIFY_LARK_RECEIVERS=oc_xxx,oc_yyy
```

按邮箱收件人的示例：

```bash
DNSHE_NOTIFY_LARK_APP_ID=cli_xxx
DNSHE_NOTIFY_LARK_APP_SECRET=yyy
DNSHE_NOTIFY_LARK_RECEIVER_TYPE=email
DNSHE_NOTIFY_LARK_RECEIVERS=alice@example.com,bob@example.com
```

## Webhook

| 变量名 | 参数说明 | 是否必须 | 示例 |
| --- | --- | --- | --- |
| `DNSHE_NOTIFY_WEBHOOK_URL` | 目标 webhook 地址 | 是 | `https://example.com/hooks/dnshe` |
| `DNSHE_NOTIFY_WEBHOOK_TOKEN` | 追加到 `Authorization: Bearer` 的令牌 | 否 | `token-123` |

约束：

- 如果设置了 `DNSHE_NOTIFY_WEBHOOK_TOKEN`，则必须同时设置 `DNSHE_NOTIFY_WEBHOOK_URL`

最小示例：

```bash
DNSHE_NOTIFY_WEBHOOK_URL=https://example.com/hooks/dnshe
```

带 Bearer Token 的示例：

```bash
DNSHE_NOTIFY_WEBHOOK_URL=https://example.com/hooks/dnshe
DNSHE_NOTIFY_WEBHOOK_TOKEN=token-123
```

## 常见配置错误

以下情况会在配置加载阶段直接报错：

- Telegram 只填了 `DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN`，没填 `DNSHE_NOTIFY_TELEGRAM_CHAT_ID`
- Mail 只填了 SMTP 地址，但没填 `DNSHE_NOTIFY_MAIL_TO`
- Mail 设置了 `DNSHE_NOTIFY_MAIL_SMTP_USERNAME`，但没设置 `DNSHE_NOTIFY_MAIL_SMTP_PASSWORD`
- Lark 同时设置了 webhook 和 custom app 两套字段
- Lark custom app 模式缺少 `DNSHE_NOTIFY_LARK_RECEIVER_TYPE`
- Webhook 只填了 `DNSHE_NOTIFY_WEBHOOK_TOKEN`，没填 `DNSHE_NOTIFY_WEBHOOK_URL`

## 组合示例

下面是一个同时启用 Console、Telegram 和 Webhook 的示例：

```bash
DNSHE_DEBUG=true

DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN=123456:abcdef
DNSHE_NOTIFY_TELEGRAM_CHAT_ID=-1001234567890
DNSHE_NOTIFY_TELEGRAM_PARSE_MODE=HTML

DNSHE_NOTIFY_WEBHOOK_URL=https://example.com/hooks/dnshe
DNSHE_NOTIFY_WEBHOOK_TOKEN=token-123
```

下面是一个同时启用 Mail 和 Lark Webhook 的示例：

```bash
DNSHE_NOTIFY_MAIL_FROM=bot@example.com
DNSHE_NOTIFY_MAIL_SMTP_HOST=smtp.example.com
DNSHE_NOTIFY_MAIL_SMTP_PORT=587
DNSHE_NOTIFY_MAIL_SMTP_USERNAME=bot@example.com
DNSHE_NOTIFY_MAIL_SMTP_PASSWORD=your-password
DNSHE_NOTIFY_MAIL_TO=alice@example.com,bob@example.com

DNSHE_NOTIFY_LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxx
```
