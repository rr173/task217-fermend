package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type statsResponse struct {
	Batches  int64 `json:"batches"`
	Channels int64 `json:"channels"`
	Samples  int64 `json:"samples"`
	Segments int64 `json:"segments"`
	Endpoints int64 `json:"endpoints"`
	Snapshots int64 `json:"snapshots"`
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	var out statsResponse
	if err := s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM batches`).Scan(&out.Batches); err != nil {
		writeErr(w, model.ErrNotFound)
		return
	}
	_ = s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM channels`).Scan(&out.Channels)
	_ = s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM samples`).Scan(&out.Samples)
	_ = s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM segments`).Scan(&out.Segments)
	_ = s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM endpoints`).Scan(&out.Endpoints)
	_ = s.svc.Store.DB().QueryRow(`SELECT COUNT(*) FROM snapshots`).Scan(&out.Snapshots)
	writeJSON(w, http.StatusOK, out)
}
