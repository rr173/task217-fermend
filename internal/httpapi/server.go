// Package httpapi 提供 JSON HTTP 接口，路由统一以 /api 为前缀。
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
)

// Server HTTP 服务。
type Server struct {
	svc *service.Service
}

// New 构造 HTTP 服务。
func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

// Handler 注册全部路由并返回 http.Handler。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 批次
	mux.HandleFunc("POST /api/batches", s.createBatch)
	mux.HandleFunc("GET /api/batches", s.listBatches)
	mux.HandleFunc("GET /api/batches/{id}", s.getBatch)
	mux.HandleFunc("POST /api/batches/{id}/transition", s.transitionBatch)
	mux.HandleFunc("POST /api/batches/{id}/seal", s.sealBatch)

	// 通道
	mux.HandleFunc("POST /api/batches/{id}/channels", s.registerChannel)
	mux.HandleFunc("GET /api/batches/{id}/channels", s.listChannels)
	mux.HandleFunc("POST /api/channels/{id}/exclude", s.excludeChannel)

	// 阶段
	mux.HandleFunc("POST /api/batches/{id}/stages", s.addStage)
	mux.HandleFunc("GET /api/batches/{id}/stages", s.listStages)

	// 采样
	mux.HandleFunc("POST /api/batches/{id}/channels/{channelID}/samples", s.ingestSamples)
	mux.HandleFunc("GET /api/channels/{id}/samples", s.listSamples)

	// 传感器段与校正
	mux.HandleFunc("POST /api/channels/{id}/segments", s.markSegment)
	mux.HandleFunc("GET /api/channels/{id}/segments", s.listSegments)
	mux.HandleFunc("POST /api/channels/{id}/estimate-lag", s.estimateLag)
	mux.HandleFunc("POST /api/segments/{id}/apply", s.applyCorrection)
	mux.HandleFunc("POST /api/segments/{id}/exclude", s.excludeSegment)

	// 诊断
	mux.HandleFunc("POST /api/batches/{id}/diagnose", s.diagnose)
	mux.HandleFunc("GET /api/batches/{id}/endpoints", s.listEndpoints)
	mux.HandleFunc("POST /api/batches/{id}/resolve", s.resolveWithSample)
	mux.HandleFunc("POST /api/endpoints/{id}/confirm", s.confirmEndpoint)
	mux.HandleFunc("POST /api/endpoints/{id}/veto", s.vetoEndpoint)

	// 快照
	mux.HandleFunc("POST /api/batches/{id}/snapshots", s.createSnapshot)
	mux.HandleFunc("GET /api/batches/{id}/snapshots", s.listSnapshots)
	mux.HandleFunc("POST /api/snapshots/{id}/publish", s.publishSnapshot)
	mux.HandleFunc("POST /api/snapshots/{id}/supersede", s.supersedeSnapshot)

	// 自检
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/stats", s.stats)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
}

func statusOf(err error) int {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, model.ErrConflict),
		errors.Is(err, model.ErrDuplicate),
		errors.Is(err, model.ErrSealed),
		errors.Is(err, model.ErrSnapshotSealed),
		errors.Is(err, model.ErrEndpointConfirmed),
		errors.Is(err, model.ErrLateSample),
		errors.Is(err, model.ErrChannelExcluded),
		errors.Is(err, model.ErrNoActiveDiagnosis):
		return http.StatusConflict
	case errors.Is(err, model.ErrInvalidArgument),
		errors.Is(err, model.ErrUnitMismatch),
		errors.Is(err, model.ErrStageOrder),
		errors.Is(err, model.ErrUnknownBatch),
		errors.Is(err, model.ErrUnknownChannel):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func decodeBody(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func pathID(r *http.Request, name string) int64 {
	v := r.PathValue(name)
	id, _ := strconv.ParseInt(v, 10, 64)
	return id
}

func jsonUnmarshalInt(s string, out *int) {
	if v, err := strconv.Atoi(s); err == nil {
		*out = v
	}
}
