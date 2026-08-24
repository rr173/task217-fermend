package store

import (
	"task217-fermend/internal/model"
)

// CreateSegment 写入一个传感器段。
func (s *Store) CreateSegment(seg *model.SensorSegment) (*model.SensorSegment, error) {
	res, err := s.db.Exec(
		`INSERT INTO segments (batch_id, channel_id, start_unix, end_unix, status, lag_secs, gain, offset, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		seg.BatchID, seg.ChannelID, seg.StartUnix, seg.EndUnix, string(seg.Status),
		seg.LagSecs, seg.Gain, seg.Offset, encodeTime(seg.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetSegment(id)
}

// GetSegment 按 id 读取段。
func (s *Store) GetSegment(id int64) (*model.SensorSegment, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, channel_id, start_unix, end_unix, status, lag_secs, gain, offset, created_at
		 FROM segments WHERE id = ?`, id)
	var seg model.SensorSegment
	var created string
	if err := row.Scan(&seg.ID, &seg.BatchID, &seg.ChannelID, &seg.StartUnix, &seg.EndUnix,
		&seg.Status, &seg.LagSecs, &seg.Gain, &seg.Offset, &created); err != nil {
		return nil, mapErr(err)
	}
	seg.CreatedAt, _ = decodeTime(created)
	return &seg, nil
}

// ListSegments 按时间窗升序列出某通道的段。
func (s *Store) ListSegments(channelID int64) ([]*model.SensorSegment, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, channel_id, start_unix, end_unix, status, lag_secs, gain, offset, created_at
		 FROM segments WHERE channel_id = ? ORDER BY start_unix`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.SensorSegment
	for rows.Next() {
		var seg model.SensorSegment
		var created string
		if err := rows.Scan(&seg.ID, &seg.BatchID, &seg.ChannelID, &seg.StartUnix, &seg.EndUnix,
			&seg.Status, &seg.LagSecs, &seg.Gain, &seg.Offset, &created); err != nil {
			return nil, err
		}
		seg.CreatedAt, _ = decodeTime(created)
		out = append(out, &seg)
	}
	return out, rows.Err()
}

// UpdateSegmentStatus 更新段状态与校正参数。
func (s *Store) UpdateSegmentStatus(id int64, status model.SegmentStatus, lagSecs, gain, offset float64) error {
	res, err := s.db.Exec(
		`UPDATE segments SET status = ?, lag_secs = ?, gain = ?, offset = ? WHERE id = ?`,
		string(status), lagSecs, gain, offset, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
