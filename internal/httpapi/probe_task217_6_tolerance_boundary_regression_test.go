package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"task217-fermend/internal/httpapi"
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
	"testing"
)

func TestResolveAtToleranceIsConfirmed(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/compare.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateEndpoint(&model.EndpointCandidate{BatchID: b.ID, Version: 1, Source: "combined", TUnix: 500, Status: model.EndpointPredicted}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	httpapi.New(svc).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/batches/1/resolve", bytes.NewBufferString(`{"version":1,"sampled_t_unix":560}`)))
	if rec.Code != http.StatusOK || bytes.Contains(rec.Body.Bytes(), []byte(`"verdict":"conflict"`)) {
		t.Fatalf("resolve response = %d %s", rec.Code, rec.Body.String())
	}
}
