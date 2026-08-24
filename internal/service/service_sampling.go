package service

import (
	"task217-fermend/internal/model"
)

// RegisterChannel 登记通道（单位校验）。
func (svc *Service) RegisterChannel(batchID int64, name string, kind model.ChannelKind, unit model.ChannelUnit) (*model.Channel, error) {
	return svc.Sampling.RegisterChannel(batchID, name, kind, unit)
}

// ListChannels 列出批次通道。
func (svc *Service) ListChannels(batchID int64) ([]*model.Channel, error) {
	return svc.Sampling.ListChannels(batchID)
}

// GetChannel 读取通道。
func (svc *Service) GetChannel(channelID int64) (*model.Channel, error) {
	return svc.Sampling.GetChannel(channelID)
}

// IngestSamples 幂等接收采样点。
func (svc *Service) IngestSamples(batchID, channelID int64, points []model.SamplePoint) error {
	return svc.Sampling.IngestPoints(batchID, channelID, points)
}

// ExcludeChannel 剔除通道。
func (svc *Service) ExcludeChannel(channelID int64) error {
	return svc.Sampling.ExcludeChannel(channelID)
}

// ListSamples 读取通道采样点。
func (svc *Service) ListSamples(channelID int64, limit int) ([]*model.SamplePoint, error) {
	return svc.Sampling.ListSamples(channelID, limit)
}
