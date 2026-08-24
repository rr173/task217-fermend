package httpapi

import (
	"net/http"

	"task217-fermend/internal/model"
	"task217-fermend/internal/snapshot"
)

type createSnapshotReq struct {
	EndpointT    int64  `json:"endpoint_t_unix"`
	Verdict      string `json:"verdict"`
	DeviationSec int64  `json:"deviation_secs"`
	Notes        string `json:"notes"`
}

func (s *Server) createSnapshot(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	var req createSnapshotReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	ev := snapshot.Evidence{
		EndpointT:     req.EndpointT,
		Verdict:       req.Verdict,
		DeviationSecs: req.DeviationSec,
		Notes:         req.Notes,
	}
	sn, err := s.svc.CreateSnapshotDraft(batchID, req.EndpointT, ev)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sn)
}

func (s *Server) listSnapshots(w http.ResponseWriter, r *http.Request) {
	batchID := pathID(r, "id")
	snaps, err := s.svc.ListSnapshots(batchID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snaps)
}

func (s *Server) publishSnapshot(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	sn, err := s.svc.PublishSnapshot(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sn)
}

type supersedeReq struct {
	NewSnapshotID int64 `json:"new_snapshot_id"`
}

func (s *Server) supersedeSnapshot(w http.ResponseWriter, r *http.Request) {
	oldID := pathID(r, "id")
	var req supersedeReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	sn, err := s.svc.SupersedeSnapshot(oldID, req.NewSnapshotID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sn)
}
