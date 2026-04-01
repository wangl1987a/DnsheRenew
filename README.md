# dnsherene

DNSHE free-domain monthly renew tool, with clear layering:

- `pkg/dnshe`: SDK layer (API client and domain models).
- `pkg/notifier`: unified notification API.
- `internal/app`: renewal orchestration logic.
- `cmd/dnsherene`: CLI entrypoint.

## Project layout

```text
cmd/dnsherene/main.go        # app bootstrap + env wiring
internal/config/config.go    # environment loading and validation
internal/app/service.go      # renew workflow
internal/app/selector.go     # target selection logic
pkg/dnshe/client.go          # DNSHE API SDK
pkg/dnshe/types.go           # SDK models
pkg/notifier/*.go            # notifier interface and implementations
```

## DNSHE SDK 覆盖接口

`pkg/dnshe` 当前已覆盖文档里的全部分组：

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

## Configuration

Required:

- `DNSHE_API_KEYS`
- `DNSHE_API_SECRETS`

Credential rule:

- Always use the plural variables above.
- Single account is represented as one item in each variable.
- `DNSHE_API_KEYS` and `DNSHE_API_SECRETS` must have the same item count.
- List separators support `,` `;` or newline.

Optional (target selection):

- `DNSHE_SUBDOMAIN_IDS`: comma-separated IDs, for example `1,2,5` (highest priority).
- `DNSHE_ROOTDOMAIN`: used only when `DNSHE_SUBDOMAIN_IDS` is empty.
- `DNSHE_SUBDOMAIN`: used only when `DNSHE_SUBDOMAIN_IDS` is empty.
- `DNSHE_DRY_RUN`: `true`/`1` to skip real renew requests.
- `DNSHE_API_BASE_URL`: default `https://api005.dnshe.com/index.php`.

Optional (unified notifications):

- `NOTIFY_WEBHOOK_URL`: if set, events are sent to webhook in JSON.
- `NOTIFY_WEBHOOK_TOKEN`: optional bearer token for webhook.

Default notifier behavior:

- Public logs only print the total renewed count across all API credentials.
- Detailed per-API notifications are sent only to webhook when `NOTIFY_WEBHOOK_URL` is set.

## Privacy

- GitHub Actions public logs do not print domain names, expiry times, remaining days, or raw API keys.
- The only public output is `renewed_total=<number>`.
- Detailed renewal results are sent through webhook fields, including:
  - per API credential updated count
  - per API credential not-updated count
  - updated domain expiry time and remaining days
  - not-updated domain list and failure reasons
  - masked API key identifier

## Run locally

```bash
DNSHE_API_KEYS="cfsd_xxx" \
DNSHE_API_SECRETS="yyy" \
DNSHE_SUBDOMAIN_IDS="1,2" \
go run ./cmd/dnsherene
```

Multi API:

```bash
DNSHE_API_KEYS="key_1,key_2" \
DNSHE_API_SECRETS="secret_1,secret_2" \
DNSHE_SUBDOMAIN_IDS="1,2" \
go run ./cmd/dnsherene
```

Dry run:

```bash
DNSHE_API_KEYS="cfsd_xxx" \
DNSHE_API_SECRETS="yyy" \
DNSHE_DRY_RUN=true \
go run ./cmd/dnsherene
```

## GitHub Actions (monthly)

Workflow file: `.github/workflows/monthly-renew.yml`

Schedule:

- `15 0 1 * *` (UTC), monthly on day 1.
- Manual run supported with `workflow_dispatch`.

### Repository settings

Add secrets:

- `DNSHE_API_KEYS`
- `DNSHE_API_SECRETS`
- `DNSHE_SUBDOMAIN_IDS` (optional, recommended)
- `NOTIFY_WEBHOOK_URL` (optional)
- `NOTIFY_WEBHOOK_TOKEN` (optional)

Add variables (optional):

- `DNSHE_ROOTDOMAIN`
- `DNSHE_SUBDOMAIN`
- `DNSHE_API_BASE_URL`
