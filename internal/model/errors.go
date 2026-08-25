package model

import "errors"

// 领域错误：HTTP 层据此映射状态码。
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrDuplicate          = errors.New("duplicate")
	ErrSealed             = errors.New("sealed immutable")
	ErrStageOrder         = errors.New("stage out of order")
	ErrUnitMismatch       = errors.New("unit mismatch")
	ErrUnknownBatch       = errors.New("unknown batch")
	ErrUnknownChannel     = errors.New("unknown channel")
	ErrLateSample         = errors.New("late sample")
	ErrNoActiveDiagnosis  = errors.New("no active diagnosis")
	ErrChannelExcluded    = errors.New("channel excluded")
	ErrSnapshotSealed     = errors.New("snapshot sealed")
	ErrEndpointConfirmed  = errors.New("endpoint already confirmed")
	ErrStoreUnavailable   = errors.New("store unavailable")
)
