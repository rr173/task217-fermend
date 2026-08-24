package store

import (
	"database/sql"
	"errors"

	"task217-fermend/internal/model"
)

// 通用错误归一化：把 sql.ErrNoRows 映射为 model.ErrNotFound。
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNotFound
	}
	return err
}
