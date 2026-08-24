package snapshot

import (
	"encoding/json"

	"task217-fermend/internal/model"
)

// Evidence 描述快照冻结的证据摘要。
type Evidence struct {
	EndpointT     int64    `json:"endpoint_t_unix"`
	Verdict       string   `json:"verdict"`
	DeviationSecs int64    `json:"deviation_secs"`
	ExcludedChans []int64  `json:"excluded_channels"`
	Notes         string   `json:"notes"`
}

// EncodeEvidence 序列化证据为 JSON 字符串。
func EncodeEvidence(e Evidence) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodeEvidence 反序列化证据。
func DecodeEvidence(s string) (*Evidence, error) {
	var e Evidence
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// LatestPublished 返回批次最新已发布（非 superseded）的快照。
func (sn *Snapshot) LatestPublished(batchID int64) (*model.DiagnosisSnapshot, error) {
	list, err := sn.store.ListSnapshots(batchID)
	if err != nil {
		return nil, err
	}
	for _, s := range list {
		if s.Status == model.SnapshotPublished {
			return s, nil
		}
	}
	return nil, model.ErrNotFound
}
