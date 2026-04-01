package dnshe

import (
	"context"
	"net/http"
)

type getQuotaResponse struct {
	baseResponse
	Quota Quota `json:"quota"`
}

// GetQuota 查询当前账号免费域名额度。
func (c *Client) GetQuota(ctx context.Context) (Quota, error) {
	var out getQuotaResponse
	if err := c.requestJSON(ctx, http.MethodGet, "quota", "", nil, nil, &out); err != nil {
		return Quota{}, err
	}
	if err := ensureSuccess("get quota", out.baseResponse); err != nil {
		return Quota{}, err
	}
	return out.Quota, nil
}
