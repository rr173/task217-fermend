package store

import (
	"task217-fermend/internal/model"
)

// CreateSnapshot 写入一个诊断快照。
func (s *Store) CreateSnapshot(sn *model.DiagnosisSnapshot) (*model.DiagnosisSnapshot, error) {
	res, err := s.db.Exec(
		`INSERT INTO snapshots (batch_id, version, status, evidence, endpoint_t_unix, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		sn.BatchID, sn.Version, string(sn.Status), sn.Evidence, sn.EndpointT, encodeTime(sn.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetSnapshot(id)
}

// GetSnapshot 按 id 读取快照。
func (s *Store) GetSnapshot(id int64) (*model.DiagnosisSnapshot, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, version, status, evidence, endpoint_t_unix, created_at
		 FROM snapshots WHERE id = ?`, id)
	var sn model.DiagnosisSnapshot
	var created string
	if err := row.Scan(&sn.ID, &sn.BatchID, &sn.Version, &sn.Status, &sn.Evidence, &sn.EndpointT, &created); err != nil {
		return nil, mapErr(err)
	}
	sn.CreatedAt, _ = decodeTime(created)
	return &sn, nil
}

// GetSnapshotByVersion 按批次+版本读取快照。
func (s *Store) GetSnapshotByVersion(batchID int64, version int) (*model.DiagnosisSnapshot, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, version, status, evidence, endpoint_t_unix, created_at
		 FROM snapshots WHERE batch_id = ? AND version = ?`, batchID, version)
	var sn model.DiagnosisSnapshot
	var created string
	if err := row.Scan(&sn.ID, &sn.BatchID, &sn.Version, &sn.Status, &sn.Evidence, &sn.EndpointT, &created); err != nil {
		return nil, mapErr(err)
	}
	sn.CreatedAt, _ = decodeTime(created)
	return &sn, nil
}

// ListSnapshots 按版本倒序列出批次快照。
func (s *Store) ListSnapshots(batchID int64) ([]*model.DiagnosisSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, version, status, evidence, endpoint_t_unix, created_at
		 FROM snapshots WHERE batch_id = ? ORDER BY version ASC`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.DiagnosisSnapshot
	for rows.Next() {
		var sn model.DiagnosisSnapshot
		var created string
		if err := rows.Scan(&sn.ID, &sn.BatchID, &sn.Version, &sn.Status, &sn.Evidence, &sn.EndpointT, &created); err != nil {
			return nil, err
		}
		sn.CreatedAt, _ = decodeTime(created)
		out = append(out, &sn)
	}
	return out, rows.Err()
}

// MaxSnapshotVersion 返回批次最大快照版本。
func (s *Store) MaxSnapshotVersion(batchID int64) (int, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM snapshots WHERE batch_id = ?`, batchID).Scan(&n)
	return n, err
}

// UpdateSnapshotStatus 更新快照状态（发布→替代等）。
func (s *Store) UpdateSnapshotStatus(id int64, status model.SnapshotStatus) error {
	res, err := s.db.Exec(
		`UPDATE snapshots SET status = ? WHERE id = ?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
