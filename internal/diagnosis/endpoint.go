package diagnosis

import (
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// Diagnosis 终点推断模块。
type Diagnosis struct {
	store *store.Store
	// ToleranceSecs 预测终点与离线采样终点的容差（秒）。
	ToleranceSecs int64
}

// New 构造诊断模块。
func New(s *store.Store) *Diagnosis {
	return &Diagnosis{store: s, ToleranceSecs: 60}
}

// ChannelLagSecs 返回某通道已应用校正段的滞后秒数（取有效段最大滞后）。
func (d *Diagnosis) ChannelLagSecs(channelID int64) float64 {
	segs, err := d.store.ListSegments(channelID)
	if err != nil {
		return 0
	}
	var lag float64
	for _, seg := range segs {
		if seg.Status == model.SegmentValid && seg.LagSecs != 0 {
			if seg.LagSecs > lag {
				lag = seg.LagSecs
			}
		}
	}
	return lag
}

// InferEndpoints 读取全部 active 通道，识别各自拐点并做滞后校正，生成终点候选。
// 返回写入的候选列表与新的诊断版本号。
func (d *Diagnosis) InferEndpoints(batchID int64) ([]*model.EndpointCandidate, int, error) {
	channels, err := d.store.ListActiveChannels(batchID)
	if err != nil {
		return nil, 0, err
	}
	if len(channels) == 0 {
		return nil, 0, model.ErrNotFound
	}

	version, err := d.store.MaxEndpointVersion(batchID)
	if err != nil {
		return nil, 0, err
	}
	version++

	var knees []Knee
	for _, ch := range channels {
		points, err := d.store.ListSamples(ch.ID, 100000)
		if err != nil {
			return nil, 0, err
		}
		lag := d.ChannelLagSecs(ch.ID)
		switch ch.Kind {
		case model.KindDissolvedOxygen:
			if t, v, ok := DoReboundTime(points); ok {
				t = t - int64(lag)
				knees = append(knees, Knee{Source: "knee_do", TUnix: t, Value: v, Channel: ch.ID})
			}
		case model.KindPH:
			if t, v, ok := PhKneeTime(points); ok {
				t = t - int64(lag)
				knees = append(knees, Knee{Source: "knee_ph", TUnix: t, Value: v, Channel: ch.ID})
			}
		case model.KindFeed:
			if t, v, ok := FeedKneeTime(points); ok {
				t = t - int64(lag)
				knees = append(knees, Knee{Source: "knee_feed", TUnix: t, Value: v, Channel: ch.ID})
			}
		default:
			// 温度/搅拌等非拐点通道不参与终点推断。
		}
	}
	if len(knees) == 0 {
		return nil, 0, model.ErrNoActiveDiagnosis
	}

	// 写各通道候选。
	var created []*model.EndpointCandidate
	for _, k := range knees {
		e := &model.EndpointCandidate{
			BatchID:   batchID,
			Version:   version,
			Source:    k.Source,
			TUnix:     k.TUnix,
			Value:     k.Value,
			Status:    model.EndpointPredicted,
			CreatedAt: time.Now().UTC(),
		}
		ce, err := d.store.CreateEndpoint(e)
		if err != nil {
			return nil, 0, err
		}
		created = append(created, ce)
	}

	// 综合候选：取各拐点时间的中位数。
	ts := make([]int64, 0, len(knees))
	for _, k := range knees {
		ts = append(ts, k.TUnix)
	}
	combined := &model.EndpointCandidate{
		BatchID:   batchID,
		Version:   version,
		Source:    "combined",
		TUnix:     MedianTime(ts),
		Value:     0,
		Status:    model.EndpointPredicted,
		Reason:    "median of channel knees",
		CreatedAt: time.Now().UTC(),
	}
	cc, err := d.store.CreateEndpoint(combined)
	if err != nil {
		return nil, 0, err
	}
	created = append(created, cc)

	return created, version, nil
}

// ListEndpoints 列出批次指定版本的终点候选。
func (d *Diagnosis) ListEndpoints(batchID int64, version int) ([]*model.EndpointCandidate, error) {
	return d.store.ListEndpoints(batchID, version)
}
