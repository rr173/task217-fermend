package model

import "time"

// BatchStatus 发酵批次状态。
type BatchStatus string

const (
	BatchPreparing        BatchStatus = "preparing"         // 准备
	BatchRunning          BatchStatus = "running"           // 运行中
	BatchPendingDiagnosis BatchStatus = "pending_diagnosis" // 待诊断
	BatchConfirmed        BatchStatus = "confirmed"         // 已确认
	BatchSealed           BatchStatus = "sealed"            // 封存
)

// ValidBatchTransitions 批次状态机：每步只能前进到允许的下一状态。
var ValidBatchTransitions = map[BatchStatus][]BatchStatus{
	BatchPreparing:        {BatchRunning},
	BatchRunning:          {BatchPendingDiagnosis, BatchSealed},
	BatchPendingDiagnosis: {BatchConfirmed, BatchSealed},
	BatchConfirmed:        {BatchSealed},
	BatchSealed:           {},
}

// ChannelKind 通道物理量类型。
type ChannelKind string

const (
	KindDissolvedOxygen ChannelKind = "do"          // 溶氧
	KindPH              ChannelKind = "ph"          // pH
	KindFeed            ChannelKind = "feed"        // 补料
	KindMetabolite      ChannelKind = "metabolite"  // 代谢物
	KindTemperature     ChannelKind = "temperature" // 温度
	KindAgitation       ChannelKind = "agitation"   // 搅拌
)

// ChannelUnit 通道物理单位，用于拒绝单位混用。
type ChannelUnit string

const (
	UnitPercent     ChannelUnit = "%"      // 溶氧饱和度
	UnitPH          ChannelUnit = "pH"     // 无量纲
	UnitGramPerL    ChannelUnit = "g/L"    // 补料/代谢物浓度
	UnitCelsius     ChannelUnit = "degC"   // 温度
	UnitRPM         ChannelUnit = "rpm"    // 搅拌
	UnitCount       ChannelUnit = "count"  // 计数
)

// ChannelStatus 通道状态。
type ChannelStatus string

const (
	ChannelActive   ChannelStatus = "active"
	ChannelExcluded ChannelStatus = "excluded"
)

// SegmentStatus 传感器段状态。
type SegmentStatus string

const (
	SegmentPendingCalibration SegmentStatus = "pending_calibration" // 待校准
	SegmentValid              SegmentStatus = "valid"               // 有效
	SegmentLagging            SegmentStatus = "lagging"             // 滞后
	SegmentGap                SegmentStatus = "gap"                 // 缺口
	SegmentExcluded           SegmentStatus = "excluded"            // 剔除
)

// EndpointStatus 终点候选状态。
type EndpointStatus string

const (
	EndpointPredicted EndpointStatus = "predicted" // 预测
	EndpointConflict  EndpointStatus = "conflict"  // 冲突
	EndpointConfirmed EndpointStatus = "confirmed" // 确认
	EndpointVetoed    EndpointStatus = "vetoed"    // 否决
)

// SnapshotStatus 诊断快照状态。
type SnapshotStatus string

const (
	SnapshotDraft      SnapshotStatus = "draft"      // 草稿
	SnapshotPublished  SnapshotStatus = "published"  // 发布
	SnapshotSuperseded SnapshotStatus = "superseded" // 替代
)

// Batch 发酵批次。
type Batch struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Strain      string      `json:"strain"`     // 菌株
	Bioreactor  string      `json:"bioreactor"` // 罐号
	Status      BatchStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Channel 传感器通道。
type Channel struct {
	ID       int64         `json:"id"`
	BatchID  int64         `json:"batch_id"`
	Name     string        `json:"name"`
	Kind     ChannelKind   `json:"kind"`
	Unit     ChannelUnit   `json:"unit"`
	Status   ChannelStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
}

// Stage 工艺阶段时间窗。
type Stage struct {
	ID        int64     `json:"id"`
	BatchID   int64     `json:"batch_id"`
	Name      string    `json:"name"`
	StartUnix int64     `json:"start_unix"` // 相对批次起点的秒
	EndUnix   int64     `json:"end_unix"`   // 相对批次起点的秒；0 表示进行中
	Seq       int       `json:"seq"`        // 阶段序号，用于拒绝逆序
	CreatedAt time.Time `json:"created_at"`
}

// SamplePoint 单条曲线采样点。
type SamplePoint struct {
	ID        int64     `json:"id"`
	BatchID   int64     `json:"batch_id"`
	ChannelID int64     `json:"channel_id"`
	Seq       int64     `json:"seq"`  // 设备序号，幂等键的一部分
	TUnix     int64     `json:"t_unix"` // 相对批次起点的秒
	Value     float64   `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// SensorSegment 传感器段：一段连续采样区间的质量标记。
type SensorSegment struct {
	ID        int64         `json:"id"`
	BatchID   int64         `json:"batch_id"`
	ChannelID int64         `json:"channel_id"`
	StartUnix int64         `json:"start_unix"`
	EndUnix   int64         `json:"end_unix"`
	Status    SegmentStatus `json:"status"`
	LagSecs   float64       `json:"lag_secs"`    // 滞后秒数（校正模块估计或人工指定）
	Gain      float64       `json:"gain"`        // 校正增益
	Offset    float64       `json:"offset"`      // 校正偏置
	CreatedAt time.Time     `json:"created_at"`
}

// EndpointCandidate 代谢终点候选。
type EndpointCandidate struct {
	ID        int64          `json:"id"`
	BatchID   int64          `json:"batch_id"`
	Version   int            `json:"version"`
	Source    string         `json:"source"` // 触发来源：knee_do / knee_ph / sample
	TUnix     int64          `json:"t_unix"`
	Value     float64        `json:"value"`
	Status    EndpointStatus `json:"status"`
	Reason    string         `json:"reason"`
	CreatedAt time.Time      `json:"created_at"`
}

// DiagnosisSnapshot 诊断快照。
type DiagnosisSnapshot struct {
	ID          int64          `json:"id"`
	BatchID     int64          `json:"batch_id"`
	Version     int            `json:"version"`
	Status      SnapshotStatus `json:"status"`
	Evidence    string         `json:"evidence"` // JSON 序列化的证据摘要
	EndpointT   int64          `json:"endpoint_t_unix"`
	CreatedAt   time.Time      `json:"created_at"`
}

// NewBatch 构造一个处于准备态的批次。
func NewBatch(name, strain, bioreactor string) *Batch {
	now := time.Now().UTC()
	return &Batch{
		Name:       name,
		Strain:     strain,
		Bioreactor: bioreactor,
		Status:     BatchPreparing,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CanTransition 校验批次状态机是否允许目标状态。
func (b *Batch) CanTransition(target BatchStatus) bool {
	if b.Status == target {
		return true
	}
	for _, allowed := range ValidBatchTransitions[b.Status] {
		if allowed == target {
			return true
		}
	}
	return false
}
