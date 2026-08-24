// Package service 编排各业务模块，暴露面向 HTTP 层的用例入口。
package service

import (
	"task217-fermend/internal/calibration"
	"task217-fermend/internal/diagnosis"
	"task217-fermend/internal/sampling"
	"task217-fermend/internal/snapshot"
	"task217-fermend/internal/stage"
	"task217-fermend/internal/store"
)

// Service 聚合所有模块，是 HTTP 层唯一依赖。
type Service struct {
	Store       *store.Store
	Sampling    *sampling.Sampling
	Stage       *stage.Stage
	Calibration *calibration.Calibration
	Diagnosis   *diagnosis.Diagnosis
	Snapshot    *snapshot.Snapshot
}

// New 装配服务及其子模块。
func New(s *store.Store) *Service {
	return &Service{
		Store:       s,
		Sampling:    sampling.New(s),
		Stage:       stage.New(s),
		Calibration: calibration.New(s),
		Diagnosis:   diagnosis.New(s),
		Snapshot:    snapshot.New(s),
	}
}
