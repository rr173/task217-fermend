package service

import (
	"task217-fermend/internal/model"
)

// CreateBatch 登记批次并置为准备态。
func (svc *Service) CreateBatch(name, strain, bioreactor string) (*model.Batch, error) {
	b := model.NewBatch(name, strain, bioreactor)
	return svc.Store.CreateBatch(b)
}

// GetBatch 读取批次。
func (svc *Service) GetBatch(id int64) (*model.Batch, error) {
	return svc.Store.GetBatch(id)
}

// ListBatches 列出批次。
func (svc *Service) ListBatches(limit, offset int) ([]*model.Batch, error) {
	return svc.Store.ListBatches(limit, offset)
}

// TransitionBatch 推进批次状态机（如 准备→运行中→待诊断→已确认→封存）。
func (svc *Service) TransitionBatch(id int64, target model.BatchStatus) (*model.Batch, error) {
	b, err := svc.Store.GetBatch(id)
	if err != nil {
		return nil, err
	}
	if !b.CanTransition(target) {
		return nil, model.ErrConflict
	}
	if err := svc.Store.UpdateBatchStatus(id, b.Status, target); err != nil {
		return nil, err
	}
	return svc.Store.GetBatch(id)
}

// SealBatch 封存批次（只允许从 confirmed / running 封存）。
func (svc *Service) SealBatch(id int64) (*model.Batch, error) {
	return svc.TransitionBatch(id, model.BatchSealed)
}
