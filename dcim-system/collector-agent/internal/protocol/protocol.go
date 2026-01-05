package protocol

import (
	"context"
	"time"
)

// CollectMode 采集模式
type CollectMode string

const (
	// CollectModePull 主动拉取模式 - Agent主动向设备发起采集请求
	CollectModePull CollectMode = "pull"
	// CollectModePush 被动接收模式 - Agent被动接收设备推送的数据
	CollectModePush CollectMode = "push"
)

// DeviceData 设备采集数据
type DeviceData struct {
	DeviceID   string                 `json:"device_id"`   // 设备ID
	DeviceIP   string                 `json:"device_ip"`   // 设备IP
	DeviceType string                 `json:"device_type"` // 设备类型
	Timestamp  time.Time              `json:"timestamp"`   // 采集时间
	Metrics    map[string]interface{} `json:"metrics"`     // 采集指标
	Status     string                 `json:"status"`      // 采集状态: success/failed
	Error      string                 `json:"error"`       // 错误信息
}

// CollectTask 采集任务
type CollectTask struct {
	TaskID     string                 `json:"task_id"`     // 任务ID
	DeviceID   string                 `json:"device_id"`   // 设备ID
	DeviceIP   string                 `json:"device_ip"`   // 设备IP
	DeviceType string                 `json:"device_type"` // 设备类型
	Protocol   string                 `json:"protocol"`    // 采集协议
	Mode       CollectMode            `json:"mode"`        // 采集模式: pull/push
	Interval   int                    `json:"interval"`    // 采集间隔(秒) - 仅pull模式有效
	Metrics    []string               `json:"metrics"`     // 采集指标列表
	Config     map[string]interface{} `json:"config"`      // 协议配置参数
	CronExpr   string                 `json:"cron_expr"`   // Cron表达式(可选) - 仅pull模式有效
}

// Protocol 协议插件接口
type Protocol interface {
	// Name 返回协议名称
	Name() string

	// Collect 执行数据采集 (主动拉取模式)
	Collect(ctx context.Context, task *CollectTask) (*DeviceData, error)

	// Validate 验证配置参数
	Validate(config map[string]interface{}) error

	// SupportedModes 返回支持的采集模式
	SupportedModes() []CollectMode

	// Close 关闭连接
	Close() error
}

// ProtocolFactory 协议工厂
type ProtocolFactory interface {
	// Create 创建协议实例
	Create(config map[string]interface{}) (Protocol, error)
}
