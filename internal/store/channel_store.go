package store

import (
	"task217-fermend/internal/model"
)

// CreateChannel 登记一个传感器通道。
func (s *Store) CreateChannel(c *model.Channel) (*model.Channel, error) {
	res, err := s.db.Exec(
		`INSERT INTO channels (batch_id, name, kind, unit, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		c.BatchID, c.Name, string(c.Kind), string(c.Unit), string(c.Status),
		encodeTime(c.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetChannel(id)
}

// GetChannel 按 id 读取通道。
func (s *Store) GetChannel(id int64) (*model.Channel, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, name, kind, unit, status, created_at
		 FROM channels WHERE id = ?`, id)
	var c model.Channel
	var created string
	if err := row.Scan(&c.ID, &c.BatchID, &c.Name, &c.Kind, &c.Unit, &c.Status, &created); err != nil {
		return nil, mapErr(err)
	}
	c.CreatedAt, _ = decodeTime(created)
	return &c, nil
}

// ListChannels 列出某批次的全部通道。
func (s *Store) ListChannels(batchID int64) ([]*model.Channel, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, name, kind, unit, status, created_at
		 FROM channels WHERE batch_id = ? ORDER BY id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Channel
	for rows.Next() {
		var c model.Channel
		var created string
		if err := rows.Scan(&c.ID, &c.BatchID, &c.Name, &c.Kind, &c.Unit, &c.Status, &created); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = decodeTime(created)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ListActiveChannels 列出未被剔除的通道。
func (s *Store) ListActiveChannels(batchID int64) ([]*model.Channel, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, name, kind, unit, status, created_at
		 FROM channels WHERE batch_id = ? AND status = ? ORDER BY id`,
		batchID, string(model.ChannelActive))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Channel
	for rows.Next() {
		var c model.Channel
		var created string
		if err := rows.Scan(&c.ID, &c.BatchID, &c.Name, &c.Kind, &c.Unit, &c.Status, &created); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = decodeTime(created)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ExcludeChannel 把通道标记为剔除（仅当当前为 active）。
func (s *Store) ExcludeChannel(id int64) error {
	res, err := s.db.Exec(
		`UPDATE channels SET status = ? WHERE id = ? AND status = ?`,
		string(model.ChannelActive), id, string(model.ChannelActive))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrConflict
	}
	return nil
}
