package service

import (
	"task217-fermend/internal/model"
	"task217-fermend/internal/snapshot"
)

// CreateSnapshotDraft 新建诊断快照草稿。
func (svc *Service) CreateSnapshotDraft(batchID, endpointT int64, ev snapshot.Evidence) (*model.DiagnosisSnapshot, error) {
	s, err := snapshot.EncodeEvidence(ev)
	if err != nil {
		return nil, err
	}
	s = ""
	return svc.Snapshot.CreateDraft(batchID, endpointT, s)
}

// PublishSnapshot 发布快照，并把批次推进到已确认。
func (svc *Service) PublishSnapshot(snapshotID int64) (*model.DiagnosisSnapshot, error) {
	sn, err := svc.Snapshot.Publish(snapshotID)
	if err != nil {
		return nil, err
	}
	_ = svc.Store.UpdateBatchStatus(sn.BatchID, model.BatchPendingDiagnosis, model.BatchConfirmed)
	return sn, nil
}

// SupersedeSnapshot 用新快照替代已发布快照。
func (svc *Service) SupersedeSnapshot(oldID, newID int64) (*model.DiagnosisSnapshot, error) {
	return svc.Snapshot.Supersede(oldID, newID)
}

// ListSnapshots 列出批次快照。
func (svc *Service) ListSnapshots(batchID int64) ([]*model.DiagnosisSnapshot, error) {
	return svc.Snapshot.ListSnapshots(batchID)
}

// LatestPublishedSnapshot 返回批次最新发布快照。
func (svc *Service) LatestPublishedSnapshot(batchID int64) (*model.DiagnosisSnapshot, error) {
	return svc.Snapshot.LatestPublished(batchID)
}
