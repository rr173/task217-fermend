package store

import (
	"task217-fermend/internal/model"
)

// CreateStage 写入一个工艺阶段；seq 由调用方（stage 模块）确定。
func (s *Store) CreateStage(st *model.Stage) (*model.Stage, error) {
	res, err := s.db.Exec(
		`INSERT INTO stages (batch_id, name, start_unix, end_unix, seq, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		st.BatchID, st.Name, st.StartUnix, st.EndUnix, st.Seq, encodeTime(st.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetStage(id)
}

// GetStage 按 id 读取阶段。
func (s *Store) GetStage(id int64) (*model.Stage, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, name, start_unix, end_unix, seq, created_at
		 FROM stages WHERE id = ?`, id)
	var st model.Stage
	var created string
	if err := row.Scan(&st.ID, &st.BatchID, &st.Name, &st.StartUnix, &st.EndUnix, &st.Seq, &created); err != nil {
		return nil, mapErr(err)
	}
	st.CreatedAt, _ = decodeTime(created)
	return &st, nil
}

// ListStages 按 seq 升序列出批次阶段。
func (s *Store) ListStages(batchID int64) ([]*model.Stage, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, name, start_unix, end_unix, seq, created_at
		 FROM stages WHERE batch_id = ? ORDER BY seq`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Stage
	for rows.Next() {
		var st model.Stage
		var created string
		if err := rows.Scan(&st.ID, &st.BatchID, &st.Name, &st.StartUnix, &st.EndUnix, &st.Seq, &created); err != nil {
			return nil, err
		}
		st.CreatedAt, _ = decodeTime(created)
		out = append(out, &st)
	}
	return out, rows.Err()
}

// MaxStageSeq 返回批次已登记的最大阶段序号。
func (s *Store) MaxStageSeq(batchID int64) (int, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(seq), 0) FROM stages WHERE batch_id = ?`, batchID).Scan(&n)
	return n, err
}

// LastStageEndUnix 返回最后一个阶段的结束时间窗（用于逆序拒绝）。
func (s *Store) LastStageEndUnix(batchID int64) (int64, error) {
	var v int64
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(start_unix), 0) FROM stages WHERE batch_id = ?`, batchID).Scan(&v)
	return v, err
}
