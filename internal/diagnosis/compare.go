package diagnosis

import (
	"math"

	"task217-fermend/internal/model"
)

// ComparisonResult 描述预测终点与离线采样终点的比对结论。
type ComparisonResult struct {
	PredictedT      int64   `json:"predicted_t_unix"`
	SampledT        int64   `json:"sampled_t_unix"`
	DeviationSecs   int64   `json:"deviation_secs"`
	WithinTolerance bool    `json:"within_tolerance"`
	Verdict         string  `json:"verdict"` // confirmed / conflict
}

// Compare 比较预测终点与实际采样终点，返回偏差与结论。
func (d *Diagnosis) Compare(predictedT, sampledT int64) ComparisonResult {
	dev := int64(math.Abs(float64(predictedT - sampledT)))
	within := dev <= d.ToleranceSecs
	verdict := "conflict"
	if within {
		verdict = "confirmed"
	}
	return ComparisonResult{
		PredictedT:      predictedT,
		SampledT:        sampledT,
		DeviationSecs:   dev,
		WithinTolerance: within,
		Verdict:         verdict,
	}
}

// ResolveWithSample 以离线采样终点为准，把综合预测候选判定为确认或冲突。
// 返回被更新的候选与比对结论。
func (d *Diagnosis) ResolveWithSample(batchID int64, version int, sampledT int64) (*model.EndpointCandidate, ComparisonResult, error) {
	eps, err := d.store.ListEndpoints(batchID, version)
	if err != nil {
		return nil, ComparisonResult{}, err
	}
	var combined *model.EndpointCandidate
	for _, e := range eps {
		if e.Source == "combined" {
			combined = e
			break
		}
	}
	res := d.Compare(combined.TUnix, sampledT)
	status := model.EndpointConflict
	if res.WithinTolerance {
		status = model.EndpointConfirmed
	}
	if err := d.store.UpdateEndpointStatus(combined.ID, status, res.Verdict); err != nil {
		return nil, ComparisonResult{}, err
	}
	combined.Status = status
	combined.Reason = res.Verdict
	return combined, res, nil
}

// ConfirmEndpoint 人工确认一个终点候选。
func (d *Diagnosis) ConfirmEndpoint(endpointID int64) (*model.EndpointCandidate, error) {
	if err := d.store.UpdateEndpointStatus(endpointID, model.EndpointConfirmed, "manual confirm"); err != nil {
		return nil, err
	}
	return d.store.GetEndpoint(endpointID)
}

// VetoEndpoint 否决一个终点候选。
func (d *Diagnosis) VetoEndpoint(endpointID int64, reason string) (*model.EndpointCandidate, error) {
	if err := d.store.UpdateEndpointStatus(endpointID, model.EndpointVetoed, reason); err != nil {
		return nil, err
	}
	return d.store.GetEndpoint(endpointID)
}
