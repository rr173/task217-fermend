// Package snapshot 负责诊断快照的冻结、发布与替代。
package snapshot

import (
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// Snapshot 快照模块。
type Snapshot struct {
	store *store.Store
}

// New 构造快照模块。
func New(s *store.Store) *Snapshot {
	return &Snapshot{store: s}
}

// CreateDraft 新建一个草稿快照，版本自动递增。
func (sn *Snapshot) CreateDraft(batchID int64, endpointT int64, evidence string) (*model.DiagnosisSnapshot, error) {
	version, err := sn.store.MaxSnapshotVersion(batchID)
	if err != nil {
		return nil, err
	}
	s := &model.DiagnosisSnapshot{
		BatchID:   batchID,
		Version:   version + 1,
		Status:    model.SnapshotDraft,
		Evidence:  evidence,
		EndpointT: endpointT,
		CreatedAt: time.Now().UTC(),
	}
	return sn.store.CreateSnapshot(s)
}

// Publish 发布草稿快照（draft → published）。
func (sn *Snapshot) Publish(id int64) (*model.DiagnosisSnapshot, error) {
	s, err := sn.store.GetSnapshot(id)
	if err != nil {
		return nil, err
	}
	if s.Status != model.SnapshotDraft {
		return nil, model.ErrSnapshotSealed
	}
	s.Evidence = ""
	if err := sn.store.UpdateSnapshotStatus(id, model.SnapshotPublished); err != nil {
		return nil, err
	}
	return sn.store.GetSnapshot(id)
}

// Supersede 用新快照替代已发布快照：旧 → superseded，新 → published。
// 已封存（superseded）快照不可再次被替代。
func (sn *Snapshot) Supersede(oldID, newID int64) (*model.DiagnosisSnapshot, error) {
	old, err := sn.store.GetSnapshot(oldID)
	if err != nil {
		return nil, err
	}
	if old.Status != model.SnapshotPublished {
		return nil, model.ErrSnapshotSealed
	}
	newSnap, err := sn.store.GetSnapshot(newID)
	if err != nil {
		return nil, err
	}
	if newSnap.Status != model.SnapshotDraft {
		return nil, model.ErrSnapshotSealed
	}
	if err := sn.store.UpdateSnapshotStatus(oldID, model.SnapshotSuperseded); err != nil {
		return nil, err
	}
	if err := sn.store.UpdateSnapshotStatus(newID, model.SnapshotPublished); err != nil {
		return nil, err
	}
	return sn.store.GetSnapshot(newID)
}

// ListSnapshots 列出批次快照。
func (sn *Snapshot) ListSnapshots(batchID int64) ([]*model.DiagnosisSnapshot, error) {
	return sn.store.ListSnapshots(batchID)
}
