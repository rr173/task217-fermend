// Package calibration 负责传感器延迟校正：段标记、滞后估计、校正应用与通道剔除。
package calibration

import (
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// Calibration 校正模块。
type Calibration struct {
	store *store.Store
}

// New 构造校正模块。
func New(s *store.Store) *Calibration {
	return &Calibration{store: s}
}

// MarkSegment 人工标记一段传感器区间（滞后/缺口/待校准），可指定滞后秒数。
func (c *Calibration) MarkSegment(batchID, channelID, startUnix, endUnix int64, status model.SegmentStatus, lagSecs float64) (*model.SensorSegment, error) {
	if status == model.SegmentValid && lagSecs == 0 {
		// 无滞后修正时不允许直接标 valid，防止误标。
		status = model.SegmentPendingCalibration
	}
	seg := &model.SensorSegment{
		BatchID:   batchID,
		ChannelID: channelID,
		StartUnix: startUnix,
		EndUnix:   endUnix,
		Status:    status,
		LagSecs:   lagSecs,
		Gain:      1,
		Offset:    0,
		CreatedAt: time.Now().UTC(),
	}
	return c.store.CreateSegment(seg)
}

// ListSegments 列出通道的段。
func (c *Calibration) ListSegments(channelID int64) ([]*model.SensorSegment, error) {
	return c.store.ListSegments(channelID)
}

// EstimateAndMarkLag 以参考通道的谷点/拐点为基准，估计目标通道滞后并落段。
// 返回估计出的滞后秒数与创建的段。
func (c *Calibration) EstimateAndMarkLag(batchID, refChannelID, targetChannelID int64, refPoints, targetPoints []*model.SamplePoint) (*model.SensorSegment, error) {
	refT, _, ok := FindMinTime(refPoints)
	if !ok {
		return nil, model.ErrNotFound
	}
	tgtT, _, ok := FindMinTime(targetPoints)
	if !ok {
		return nil, model.ErrNotFound
	}
	lag := EstimateLag(refT, tgtT)
	if lag < 0 {
		lag = 0
	}
	seg := &model.SensorSegment{
		BatchID:   batchID,
		ChannelID: targetChannelID,
		StartUnix: targetPoints[0].TUnix,
		EndUnix:   targetPoints[len(targetPoints)-1].TUnix,
		Status:    model.SegmentLagging,
		LagSecs:   lag,
		Gain:      1,
		Offset:    0,
		CreatedAt: time.Now().UTC(),
	}
	return c.store.CreateSegment(seg)
}

// ApplyCorrection 应用校正：把 lagging 段标记为 valid 并保留滞后参数，供诊断阶段平移时间。
func (c *Calibration) ApplyCorrection(segmentID int64) (*model.SensorSegment, error) {
	seg, err := c.store.GetSegment(segmentID)
	if err != nil {
		return nil, err
	}
	if err := c.store.UpdateSegmentStatus(segmentID, model.SegmentValid, seg.LagSecs, seg.Gain, seg.Offset); err != nil {
		return nil, err
	}
	return c.store.GetSegment(segmentID)
}

// ExcludeSegment 剔除一段传感器区间。
func (c *Calibration) ExcludeSegment(segmentID int64) error {
	return c.store.UpdateSegmentStatus(segmentID, model.SegmentExcluded, 0, 1, 0)
}

// ExcludeChannel 剔除整个通道。
func (c *Calibration) ExcludeChannel(channelID int64) error {
	return c.store.ExcludeChannel(channelID)
}
