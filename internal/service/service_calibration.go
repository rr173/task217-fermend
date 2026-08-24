package service

import (
	"task217-fermend/internal/model"
)

// AddStage 登记工艺阶段（拒绝逆序）。
func (svc *Service) AddStage(batchID int64, name string, startUnix, endUnix int64) (*model.Stage, error) {
	return svc.Stage.AddStage(batchID, name, startUnix, endUnix)
}

// ListStages 列出批次阶段。
func (svc *Service) ListStages(batchID int64) ([]*model.Stage, error) {
	return svc.Stage.ListStages(batchID)
}

// MarkSegment 人工标记传感器段。
func (svc *Service) MarkSegment(batchID, channelID, startUnix, endUnix int64, status model.SegmentStatus, lagSecs float64) (*model.SensorSegment, error) {
	return svc.Calibration.MarkSegment(batchID, channelID, startUnix, endUnix, status, lagSecs)
}

// ListSegments 列出通道段。
func (svc *Service) ListSegments(channelID int64) ([]*model.SensorSegment, error) {
	return svc.Calibration.ListSegments(channelID)
}

// EstimateLag 以参考通道为基准估计目标通道滞后并落段。
func (svc *Service) EstimateLag(batchID, refChannelID, targetChannelID int64) (*model.SensorSegment, error) {
	refPoints, err := svc.Sampling.ListSamples(refChannelID, 100000)
	if err != nil {
		return nil, err
	}
	targetPoints, err := svc.Sampling.ListSamples(targetChannelID, 100000)
	if err != nil {
		return nil, err
	}
	return svc.Calibration.EstimateAndMarkLag(batchID, refChannelID, targetChannelID, refPoints, targetPoints)
}

// ApplyCorrection 应用段校正。
func (svc *Service) ApplyCorrection(segmentID int64) (*model.SensorSegment, error) {
	return svc.Calibration.ApplyCorrection(segmentID)
}

// ExcludeSegment 剔除段。
func (svc *Service) ExcludeSegment(segmentID int64) error {
	return svc.Calibration.ExcludeSegment(segmentID)
}
