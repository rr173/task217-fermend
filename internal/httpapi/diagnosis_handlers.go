package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

func (s *Server) diagnose(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	eps, version, err := s.svc.Diagnose(batchID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version":    version,
		"candidates": eps,
	})
}

func (s *Server) listEndpoints(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	version := 0
	if v := r.URL.Query().Get("version"); v != "" {
		jsonUnmarshalInt(v, &version)
	}
	if version == 0 {
		var err error
		version, err = s.svc.Store.MaxEndpointVersion(batchID)
		if err != nil {
			writeErr(w, err)
			return
		}
	}
	eps, err := s.svc.ListEndpoints(batchID, version)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eps)
}

type resolveReq struct {
	Version  int   `json:"version"`
	SampledT int64 `json:"sampled_t_unix"`
}

func (s *Server) resolveWithSample(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	var req resolveReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	if req.Version == 0 {
		var err error
		req.Version, err = s.svc.Store.MaxEndpointVersion(batchID)
		if err != nil {
			writeErr(w, err)
			return
		}
	}
	ep, cmp, err := s.svc.ResolveWithSample(batchID, req.Version, req.SampledT)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"endpoint": ep, "comparison": cmp})
}

func (s *Server) confirmEndpoint(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	ep, err := s.svc.ConfirmEndpoint(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ep)
}

type vetoReq struct {
	Reason string `json:"reason"`
}

func (s *Server) vetoEndpoint(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	var req vetoReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	ep, err := s.svc.VetoEndpoint(id, req.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ep)
}
