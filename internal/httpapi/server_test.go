package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
)

func TestHealthAndBatchRoutes(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/http.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := New(service.New(s)).Handler()

	health := httptest.NewRecorder()
	h.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK || !bytes.Contains(health.Body.Bytes(), []byte(`"status":"ok"`)) {
		t.Fatalf("health response = %d %s", health.Code, health.Body.String())
	}

	body, _ := json.Marshal(map[string]string{"name": "batch-1", "strain": "strain-1", "bioreactor": "BR-1"})
	created := httptest.NewRecorder()
	h.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewReader(body)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create batch status = %d, body = %s", created.Code, created.Body.String())
	}
	var batch struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &batch); err != nil || batch.ID == 0 {
		t.Fatalf("created batch = %s, err = %v", created.Body.String(), err)
	}

	badChannel, _ := json.Marshal(map[string]string{"name": "bad-do", "kind": "do", "unit": "pH"})
	bad := httptest.NewRecorder()
	h.ServeHTTP(bad, httptest.NewRequest(http.MethodPost, "/api/batches/1/channels", bytes.NewReader(badChannel)))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad channel status = %d, want 400; body = %s", bad.Code, bad.Body.String())
	}
}
