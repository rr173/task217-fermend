// Package sampling 负责采样接收：通道登记、曲线写入与单位校验。
package sampling

import (
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// ExpectedUnit 返回某物理量类型允许的单位集合，用于拒绝单位混用。
func ExpectedUnit(kind model.ChannelKind) []model.ChannelUnit {
	switch kind {
	case model.KindDissolvedOxygen:
		return []model.ChannelUnit{model.UnitPercent}
	case model.KindPH:
		return []model.ChannelUnit{model.UnitPH}
	case model.KindFeed, model.KindMetabolite:
		return []model.ChannelUnit{model.UnitGramPerL}
	case model.KindTemperature:
		return []model.ChannelUnit{model.UnitCelsius}
	case model.KindAgitation:
		return []model.ChannelUnit{model.UnitRPM}
	default:
		return []model.ChannelUnit{model.UnitCount}
	}
}

// UnitAllowed 判断单位是否被该类型允许。
func UnitAllowed(kind model.ChannelKind, unit model.ChannelUnit) bool {
	for _, u := range ExpectedUnit(kind) {
		if u == unit {
			return true
		}
	}
	return false
}

// Sampling 采样模块，持有持久化入口。
type Sampling struct {
	store *store.Store
}

// New 构造采样模块。
func New(s *store.Store) *Sampling {
	return &Sampling{store: s}
}

// RegisterChannel 登记通道，校验单位与类型匹配。
func (sm *Sampling) RegisterChannel(batchID int64, name string, kind model.ChannelKind, unit model.ChannelUnit) (*model.Channel, error) {
	if !UnitAllowed(kind, unit) {
		return nil, model.ErrUnitMismatch
	}
	c := &model.Channel{
		BatchID:   batchID,
		Name:      name,
		Kind:      kind,
		Unit:      unit,
		Status:    model.ChannelActive,
		CreatedAt: time.Now().UTC(),
	}
	return sm.store.CreateChannel(c)
}

// IngestPoints 幂等接收一批采样点，按 (batch_id, channel_id, seq) 去重。
func (sm *Sampling) IngestPoints(batchID, channelID int64, points []model.SamplePoint) error {
	ptrs := make([]*model.SamplePoint, 0, len(points))
	for i := range points {
		points[i].BatchID = batchID
		points[i].ChannelID = channelID
		if points[i].CreatedAt.IsZero() {
			points[i].CreatedAt = time.Now().UTC()
		}
		ptrs = append(ptrs, &points[i])
	}
	return sm.store.InsertSamplesTx(ptrs)
}

// ListChannels 列出批次全部通道。
func (sm *Sampling) ListChannels(batchID int64) ([]*model.Channel, error) {
	return sm.store.ListChannels(batchID)
}
