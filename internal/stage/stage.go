// Package stage 维护工艺阶段时间窗，拒绝阶段逆序。
package stage

import (
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// Stage 阶段模块。
type Stage struct {
	store *store.Store
}

// New 构造阶段模块。
func New(s *store.Store) *Stage {
	return &Stage{store: s}
}

// AddStage 登记一个工艺阶段。相邻阶段允许共享边界（半开窗 [start, end)），
// 故后一阶段的 start 可以等于上一阶段的 end；end（若非 0）必须大于 start。
func (st *Stage) AddStage(batchID int64, name string, startUnix, endUnix int64) (*model.Stage, error) {
	if endUnix != 0 && endUnix <= startUnix {
		return nil, model.ErrStageOrder
	}
	lastEnd, err := st.store.LastStageEndUnix(batchID)
	if err != nil {
		return nil, err
	}
	// 边界时刻归入后一阶段：start 只需不早于上一阶段 end（相等即拼接）。
	if startUnix < lastEnd {
		return nil, model.ErrStageOrder
	}
	seq, err := st.store.MaxStageSeq(batchID)
	if err != nil {
		return nil, err
	}
	s := &model.Stage{
		BatchID:   batchID,
		Name:      name,
		StartUnix: startUnix,
		EndUnix:   endUnix,
		Seq:       seq + 1,
		CreatedAt: time.Now().UTC(),
	}
	return st.store.CreateStage(s)
}

// ListStages 列出批次阶段。
func (st *Stage) ListStages(batchID int64) ([]*model.Stage, error) {
	return st.store.ListStages(batchID)
}
