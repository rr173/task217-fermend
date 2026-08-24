package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

type markSegmentReq struct {
	StartUnix int64              `json:"start_unix"`
	EndUnix   int64              `json:"end_unix"`
	Status    model.SegmentStatus `json:"status"`
	LagSecs   float64            `json:"lag_secs"`
}

func (s *Server) markSegment(w http.ResponseWriter, r *http.Request) {
	channelID := pathID(r, "id")
	var req markSegmentReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	ch, err := s.svc.GetChannel(channelID)
	if err != nil {
		writeErr(w, err)
		return
	}
	seg, err := s.svc.MarkSegment(ch.BatchID, channelID, req.StartUnix, req.EndUnix, req.Status, req.LagSecs)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, seg)
}

func (s *Server) listSegments(w http.ResponseWriter, r *http.Request) {
	channelID := pathID(r, "id")
	segs, err := s.svc.ListSegments(channelID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, segs)
}

type estimateLagReq struct {
	RefChannelID int64 `json:"ref_channel_id"`
}

func (s *Server) estimateLag(w http.ResponseWriter, r *http.Request) {
	channelID := pathID(r, "id")
	var req estimateLagReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	ch, err := s.svc.GetChannel(channelID)
	if err != nil {
		writeErr(w, err)
		return
	}
	seg, err := s.svc.EstimateLag(ch.BatchID, req.RefChannelID, channelID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, seg)
}

func (s *Server) applyCorrection(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	seg, err := s.svc.ApplyCorrection(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, seg)
}

func (s *Server) excludeSegment(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	if err := s.svc.ExcludeSegment(id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"excluded": true})
}
