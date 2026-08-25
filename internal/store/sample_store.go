package store

import (
	"database/sql"

	"task217-fermend/internal/model"
)

// InsertSample 写入一条采样点。幂等键为 (batch_id, channel_id, seq)。
// 若已存在则不覆盖（保留第一次写入的时间与值），并返回 model.ErrDuplicate。
func (s *Store) InsertSample(p *model.SamplePoint) error {
	if err := s.check(); err != nil {
		return err
	}
	res, err := s.db.Exec(
		`INSERT OR IGNORE INTO samples (batch_id, channel_id, seq, t_unix, value, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		p.BatchID, p.ChannelID, p.Seq, p.TUnix, p.Value, encodeTime(p.CreatedAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrDuplicate
	}
	return nil
}

// InsertSamplesTx 在事务内批量写入采样点。
// 任一 (batch_id, channel_id, seq) 已存在则静默跳过（保留第一次写入的时间与值），
// 整批仍成功提交——即对重复序号幂等、不报错。
func (s *Store) InsertSamplesTx(points []*model.SamplePoint) error {
	if err := s.check(); err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, p := range points {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO samples (batch_id, channel_id, seq, t_unix, value, created_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			p.BatchID, p.ChannelID, p.Seq, p.TUnix, p.Value, encodeTime(p.CreatedAt)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListSamples 按时间升序列出某通道的采样点。
func (s *Store) ListSamples(channelID int64, limit int) ([]*model.SamplePoint, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(
		`SELECT id, batch_id, channel_id, seq, t_unix, value, created_at
		 FROM samples WHERE channel_id = ? ORDER BY t_unix ASC, seq ASC LIMIT ?`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSamples(rows)
}

// ListSamplesRange 按时间窗列出某通道采样点。
func (s *Store) ListSamplesRange(channelID int64, startUnix, endUnix int64) ([]*model.SamplePoint, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(
		`SELECT id, batch_id, channel_id, seq, t_unix, value, created_at
		 FROM samples WHERE channel_id = ? AND t_unix >= ? AND t_unix <= ?
		 ORDER BY t_unix ASC, seq ASC`, channelID, startUnix, endUnix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSamples(rows)
}

func scanSamples(rows *sql.Rows) ([]*model.SamplePoint, error) {
	var out []*model.SamplePoint
	for rows.Next() {
		var p model.SamplePoint
		var created string
		if err := rows.Scan(&p.ID, &p.BatchID, &p.ChannelID, &p.Seq, &p.TUnix, &p.Value, &created); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = decodeTime(created)
		out = append(out, &p)
	}
	return out, rows.Err()
}

// CountSamples 返回某通道采样点数量。
func (s *Store) CountSamples(channelID int64) (int64, error) {
	if err := s.check(); err != nil {
		return 0, err
	}
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM samples WHERE channel_id = ?`, channelID).Scan(&n)
	return n, err
}
