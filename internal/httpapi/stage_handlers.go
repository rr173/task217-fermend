package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

type addStageReq struct {
	Name      string `json:"name"`
	StartUnix int64  `json:"start_unix"`
	EndUnix   int64  `json:"end_unix"`
}

func (s *Server) addStage(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	var req addStageReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	if req.Name == "" {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	st, err := s.svc.AddStage(batchID, req.Name, req.StartUnix, req.EndUnix)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (s *Server) listStages(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	stages, err := s.svc.ListStages(batchID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stages)
}
