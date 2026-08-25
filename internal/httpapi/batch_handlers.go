package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

type createBatchReq struct {
	Name       string `json:"name"`
	Strain     string `json:"strain"`
	Bioreactor string `json:"bioreactor"`
}

func (s *Server) createBatch(w http.ResponseWriter, r *http.Request) {
	var req createBatchReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	if req.Name == "" {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	b, err := s.svc.CreateBatch(req.Name, req.Strain, req.Bioreactor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.svc.ListBatches(200, 0)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	b, err := s.svc.GetBatch(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

type transitionReq struct {
	Target model.BatchStatus `json:"target"`
}

func (s *Server) transitionBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	var req transitionReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	b, err := s.svc.TransitionBatch(id, req.Target)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) sealBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	b, err := s.svc.SealBatch(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}
