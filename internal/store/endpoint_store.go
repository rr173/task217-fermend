package store

import (
	"task217-fermend/internal/model"
)

// CreateEndpoint 写入一个终点候选。幂等键 (batch_id, version, source)。
func (s *Store) CreateEndpoint(e *model.EndpointCandidate) (*model.EndpointCandidate, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`INSERT INTO endpoints (batch_id, version, source, t_unix, value, status, reason, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.BatchID, e.Version, e.Source, e.TUnix, e.Value, string(e.Status), e.Reason, encodeTime(e.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetEndpoint(id)
}

// GetEndpoint 按 id 读取终点候选。
func (s *Store) GetEndpoint(id int64) (*model.EndpointCandidate, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	row := s.db.QueryRow(
		`SELECT id, batch_id, version, source, t_unix, value, status, reason, created_at
		 FROM endpoints WHERE id = ?`, id)
	var e model.EndpointCandidate
	var created string
	if err := row.Scan(&e.ID, &e.BatchID, &e.Version, &e.Source, &e.TUnix, &e.Value, &e.Status, &e.Reason, &created); err != nil {
		return nil, mapErr(err)
	}
	e.CreatedAt, _ = decodeTime(created)
	return &e, nil
}

// ListEndpoints 列出某批次某版本的全部终点候选。
func (s *Store) ListEndpoints(batchID int64, version int) ([]*model.EndpointCandidate, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(
		`SELECT id, batch_id, version, source, t_unix, value, status, reason, created_at
		 FROM endpoints WHERE batch_id = ? AND version = ? ORDER BY source`, batchID, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.EndpointCandidate
	for rows.Next() {
		var e model.EndpointCandidate
		var created string
		if err := rows.Scan(&e.ID, &e.BatchID, &e.Version, &e.Source, &e.TUnix, &e.Value, &e.Status, &e.Reason, &created); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = decodeTime(created)
		out = append(out, &e)
	}
	return out, rows.Err()
}

// MaxEndpointVersion 返回批次最大终点版本。
func (s *Store) MaxEndpointVersion(batchID int64) (int, error) {
	if err := s.check(); err != nil {
		return 0, err
	}
	var n int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM endpoints WHERE batch_id = ?`, batchID).Scan(&n)
	return n, err
}

// UpdateEndpointStatus 更新终点状态（预测→确认/否决等）。
func (s *Store) UpdateEndpointStatus(id int64, status model.EndpointStatus, reason string) error {
	if err := s.check(); err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE endpoints SET status = ?, reason = ? WHERE id = ?`,
		string(status), reason, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
