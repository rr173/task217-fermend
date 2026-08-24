package sampling

import (
	"task217-fermend/internal/model"
)

// ExcludeChannel 剔除一个通道，使其不再参与诊断。
func (sm *Sampling) ExcludeChannel(channelID int64) error {
	return sm.store.ExcludeChannel(channelID)
}

// GetChannel 读取通道详情。
func (sm *Sampling) GetChannel(channelID int64) (*model.Channel, error) {
	return sm.store.GetChannel(channelID)
}

// ListSamples 读取某通道采样点（诊断与校正复用）。
func (sm *Sampling) ListSamples(channelID int64, limit int) ([]*model.SamplePoint, error) {
	return sm.store.ListSamples(channelID, limit)
}

// ListSamplesRange 读取某通道时间窗内采样点。
func (sm *Sampling) ListSamplesRange(channelID int64, startUnix, endUnix int64) ([]*model.SamplePoint, error) {
	return sm.store.ListSamplesRange(channelID, startUnix, endUnix)
}
