package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

type registerChannelReq struct {
	Name string            `json:"name"`
	Kind model.ChannelKind `json:"kind"`
	Unit model.ChannelUnit `json:"unit"`
}

func (s *Server) registerChannel(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	var req registerChannelReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	if req.Name == "" {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	c, err := s.svc.RegisterChannel(batchID, req.Name, req.Kind, req.Unit)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	chs, err := s.svc.ListChannels(batchID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chs)
}

func (s *Server) excludeChannel(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	if err := s.svc.ExcludeChannel(id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"excluded": true})
}
