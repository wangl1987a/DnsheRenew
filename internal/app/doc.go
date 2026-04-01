// Package app 提供续期任务的应用层编排逻辑。
//
// 该包只负责“流程组织”，不直接关心底层 HTTP 调用细节：
// - DNSHE API 调用由 pkg/dnshe 提供。
// - 通知通道由 pkg/notifier 通过接口注入。
package app
