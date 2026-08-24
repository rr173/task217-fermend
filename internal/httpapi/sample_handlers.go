package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
)

type ingestReq struct {
	Points []samplePointReq `json:"points"`
}

type samplePointReq struct {
	Seq   int64   `json:"seq"`
	TUnix int64   `json:"t_unix"`
	Value float64 `json:"value"`
}

func (s *Server) ingestSamples(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	channelID := pathID(r, "channelID")
	var req ingestReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	points := make([]model.SamplePoint, 0, len(req.Points))
	for _, p := range req.Points {
		points = append(points, model.SamplePoint{Seq: p.Seq, TUnix: p.TUnix, Value: p.Value})
	}
	if err := s.svc.IngestSamples(batchID, channelID, points); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"ingested": len(points)})
}

func (s *Server) listSamples(w http.ResponseWriter, r *http.Request) {
	channelID := pathID(r, "id")
	pts, err := s.svc.ListSamples(channelID, 100000)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pts)
}
