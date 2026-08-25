package store

import (
	"time"

	"task217-fermend/internal/model"
)

const timeLayout = time.RFC3339Nano

func encodeTime(t time.Time) string { return t.UTC().Format(timeLayout) }
func decodeTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}

// CreateBatch 登记一个发酵批次。
func (s *Store) CreateBatch(b *model.Batch) (*model.Batch, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`INSERT INTO batches (name, strain, bioreactor, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		b.Name, b.Strain, b.Bioreactor, string(b.Status),
		encodeTime(b.CreatedAt), encodeTime(b.UpdatedAt),
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetBatch(id)
}

// GetBatch 按 id 读取批次。
func (s *Store) GetBatch(id int64) (*model.Batch, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	row := s.db.QueryRow(
		`SELECT id, name, strain, bioreactor, status, created_at, updated_at
		 FROM batches WHERE id = ?`, id)
	var b model.Batch
	var created, updated string
	if err := row.Scan(&b.ID, &b.Name, &b.Strain, &b.Bioreactor, &b.Status, &created, &updated); err != nil {
		return nil, mapErr(err)
	}
	b.CreatedAt, _ = decodeTime(created)
	b.UpdatedAt, _ = decodeTime(updated)
	return &b, nil
}

// ListBatches 按创建时间倒序列出批次。
func (s *Store) ListBatches(limit, offset int) ([]*model.Batch, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(
		`SELECT id, name, strain, bioreactor, status, created_at, updated_at
		 FROM batches ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Batch
	for rows.Next() {
		var b model.Batch
		var created, updated string
		if err := rows.Scan(&b.ID, &b.Name, &b.Strain, &b.Bioreactor, &b.Status, &created, &updated); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = decodeTime(created)
		b.UpdatedAt, _ = decodeTime(updated)
		out = append(out, &b)
	}
	return out, rows.Err()
}

// UpdateBatchStatus 原子地更新批次状态，仅在当前状态匹配时成功。
func (s *Store) UpdateBatchStatus(id int64, from, to model.BatchStatus) error {
	if err := s.check(); err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE batches SET status = ?, updated_at = ? WHERE id = ? AND status = ?`,
		string(to), encodeTime(time.Now().UTC()), id, string(from))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrConflict
	}
	return nil
}

// CountBatches 返回批次总数，供自检与统计使用。
func (s *Store) CountBatches() (int64, error) {
	if err := s.check(); err != nil {
		return 0, err
	}
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM batches`).Scan(&n)
	return n, err
}
