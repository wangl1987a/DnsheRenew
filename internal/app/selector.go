package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"dnsherene/pkg/dnshe"
)

// selectTargets 根据输入配置筛选本次续期目标。
//
// 规则：
// 1. 当 idListRaw 非空时，按 ID 列表选取，且优先级高于过滤条件。
// 2. 当 idListRaw 为空时，使用 rootdomainFilter/subdomainFilter 过滤。
func selectTargets(
	all []dnshe.Subdomain,
	idListRaw string,
	rootdomainFilter string,
	subdomainFilter string,
) ([]dnshe.Subdomain, error) {
	byID := make(map[int]dnshe.Subdomain, len(all))
	for _, s := range all {
		byID[s.ID] = s
	}

	result := make([]dnshe.Subdomain, 0)
	seen := make(map[int]struct{})

	if strings.TrimSpace(idListRaw) != "" {
		// 约定：显式 ID 选择优先级最高，忽略名称过滤条件。
		ids, err := parseIDs(idListRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid DNSHE_SUBDOMAIN_IDS: %w", err)
		}

		for _, id := range ids {
			item, ok := byID[id]
			if !ok {
				// 兼容“直接按 ID 续期”的场景：即使列表接口没返回该 ID，也保留目标。
				item = dnshe.Subdomain{
					ID:         id,
					FullDomain: fmt.Sprintf("id-%d", id),
				}
			}
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			result = append(result, item)
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
		return result, nil
	}

	for _, s := range all {
		if !matchesFilter(s.Rootdomain, rootdomainFilter) {
			continue
		}
		if !matchesFilter(s.Subdomain, subdomainFilter) {
			continue
		}
		if _, ok := seen[s.ID]; ok {
			continue
		}
		seen[s.ID] = struct{}{}
		result = append(result, s)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

// parseIDs 解析逗号分隔的 ID 列表，返回去空格后的正整数切片。
func parseIDs(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not an integer", p)
		}
		if id <= 0 {
			return nil, fmt.Errorf("id must be positive: %d", id)
		}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no IDs found")
	}
	return out, nil
}

// matchesFilter 采用大小写不敏感的精确匹配；空过滤条件表示匹配全部。
func matchesFilter(value string, filter string) bool {
	if strings.TrimSpace(filter) == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(filter))
}
