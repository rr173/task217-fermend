package service

import (
	"task217-fermend/internal/diagnosis"
	"task217-fermend/internal/model"
)

// Diagnose 运行终点诊断：识别拐点、生成候选，并把批次推进到待诊断。
func (svc *Service) Diagnose(batchID int64) ([]*model.EndpointCandidate, int, error) {
	eps, version, err := svc.Diagnosis.InferEndpoints(batchID)
	if err != nil {
		return nil, 0, err
	}
	// 诊断成功后将批次推进到待诊断（若处于运行中）。
	_ = svc.Store.UpdateBatchStatus(batchID, model.BatchRunning, model.BatchPendingDiagnosis)
	return eps, version, nil
}

// ListEndpoints 列出批次某版本终点候选。
func (svc *Service) ListEndpoints(batchID int64, version int) ([]*model.EndpointCandidate, error) {
	return svc.Diagnosis.ListEndpoints(batchID, version)
}

// ResolveWithSample 用离线采样终点判定综合预测候选。
func (svc *Service) ResolveWithSample(batchID int64, version int, sampledT int64) (*model.EndpointCandidate, diagnosis.ComparisonResult, error) {
	return svc.Diagnosis.ResolveWithSample(batchID, version, sampledT)
}

// ConfirmEndpoint 人工确认终点候选。
func (svc *Service) ConfirmEndpoint(endpointID int64) (*model.EndpointCandidate, error) {
	return svc.Diagnosis.ConfirmEndpoint(endpointID)
}

// VetoEndpoint 否决终点候选。
func (svc *Service) VetoEndpoint(endpointID int64, reason string) (*model.EndpointCandidate, error) {
	return svc.Diagnosis.VetoEndpoint(endpointID, reason)
}

// Compare 比较两个终点时刻。
func (svc *Service) Compare(predictedT, sampledT int64) diagnosis.ComparisonResult {
	return svc.Diagnosis.Compare(predictedT, sampledT)
}
