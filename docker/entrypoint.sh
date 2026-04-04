#!/bin/sh

set -eu

# 容器内运行参数：crond 负责调度，业务命令以非 root 用户执行。
app_user="dnshe"
app_bin="/usr/local/bin/dnsherene"
crontab_file="/etc/crontabs/${app_user}"

# 校验必填环境变量，缺失时直接退出容器。
require_env() {
  name="$1"
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    echo "$name is required" >&2
    exit 1
  fi
}

# 将多行输入折叠为逗号分隔，便于写入 crontab 环境变量。
normalize_list() {
  printf "%s" "$1" | tr '\r\n' ',,'
}

# 以 shell 安全的方式写入 crontab 中的环境变量。
write_assignment() {
  name="$1"
  value="$2"
  escaped=$(printf "%s" "$value" | sed "s/'/'\\\\''/g")
  printf "%s='%s'\n" "$name" "$escaped" >> "$crontab_file"
}

# 根据 TZ 设置容器本地时区，cron 和手动执行都会共享该时区。
setup_timezone() {
  if [ -n "${TZ:-}" ] && [ -f "/usr/share/zoneinfo/${TZ}" ]; then
    ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime
    echo "${TZ}" > /etc/timezone
  fi
}

# 生成 crontab：写入运行环境，并把日志重定向到容器标准输出。
setup_crontab() {
  schedule="${DNSHE_CRON_SCHEDULE:-15 0 1 * *}"

  : > "$crontab_file"
  chmod 600 "$crontab_file"

  printf "SHELL=/bin/sh\n" >> "$crontab_file"
  printf "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n" >> "$crontab_file"
  printf "HOME=/home/%s\n" "$app_user" >> "$crontab_file"

  write_assignment "TZ" "${TZ:-UTC}"
  write_assignment "DNSHE_API_KEYS" "$(normalize_list "${DNSHE_API_KEYS:-}")"
  write_assignment "DNSHE_API_SECRETS" "$(normalize_list "${DNSHE_API_SECRETS:-}")"
  write_assignment "DNSHE_API_BASE_URL" "${DNSHE_API_BASE_URL:-}"
  write_assignment "DNSHE_DRY_RUN" "${DNSHE_DRY_RUN:-false}"
  write_assignment "DNSHE_DEBUG" "${DNSHE_DEBUG:-false}"
  write_assignment "DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN" "${DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN:-}"
  write_assignment "DNSHE_NOTIFY_TELEGRAM_CHAT_ID" "${DNSHE_NOTIFY_TELEGRAM_CHAT_ID:-}"
  write_assignment "DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID" "${DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID:-}"
  write_assignment "DNSHE_NOTIFY_TELEGRAM_PARSE_MODE" "${DNSHE_NOTIFY_TELEGRAM_PARSE_MODE:-}"
  write_assignment "DNSHE_NOTIFY_LARK_WEBHOOK_URL" "${DNSHE_NOTIFY_LARK_WEBHOOK_URL:-}"
  write_assignment "DNSHE_NOTIFY_LARK_APP_ID" "${DNSHE_NOTIFY_LARK_APP_ID:-}"
  write_assignment "DNSHE_NOTIFY_LARK_APP_SECRET" "${DNSHE_NOTIFY_LARK_APP_SECRET:-}"
  write_assignment "DNSHE_NOTIFY_LARK_RECEIVER_TYPE" "${DNSHE_NOTIFY_LARK_RECEIVER_TYPE:-}"
  write_assignment "DNSHE_NOTIFY_LARK_RECEIVERS" "$(normalize_list "${DNSHE_NOTIFY_LARK_RECEIVERS:-}")"
  write_assignment "DNSHE_NOTIFY_WEBHOOK_URL" "${DNSHE_NOTIFY_WEBHOOK_URL:-}"
  write_assignment "DNSHE_NOTIFY_WEBHOOK_TOKEN" "${DNSHE_NOTIFY_WEBHOOK_TOKEN:-}"

  printf "%s %s >> /proc/1/fd/1 2>> /proc/1/fd/2\n" "$schedule" "$app_bin" >> "$crontab_file"
  echo "installed cron schedule: $schedule"
}

# 立即执行一次主程序，常用于 `docker compose run --rm dnsherene run`。
run_once() {
  require_env "DNSHE_API_KEYS"
  require_env "DNSHE_API_SECRETS"
  exec su-exec "$app_user:$app_user" "$app_bin" "$@"
}

# 默认进入 cron 模式；传入 `run` 时立即执行一次；其他参数按原样执行。
case "${1:-cron}" in
  cron)
    require_env "DNSHE_API_KEYS"
    require_env "DNSHE_API_SECRETS"
    setup_timezone
    setup_crontab
    exec crond -f -l 2 -c /etc/crontabs
    ;;
  run)
    shift
    setup_timezone
    run_once "$@"
    ;;
  *)
    exec "$@"
    ;;
esac
