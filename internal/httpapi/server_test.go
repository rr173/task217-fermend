package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// TestSealBatch 推进 confirmed 批次封存，并校验 running 直接封存的合法路径。
func TestSealBatch(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/seal.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := New(service.New(s)).Handler()

	createBatch := func(name string) int64 {
		body, _ := json.Marshal(map[string]string{"name": name, "strain": "strain-1", "bioreactor": "BR-1"})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewReader(body)))
		if rec.Code != http.StatusCreated {
			t.Fatalf("create batch status = %d, body = %s", rec.Code, rec.Body.String())
		}
		var b struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || b.ID == 0 {
			t.Fatalf("created batch = %s, err = %v", rec.Body.String(), err)
		}
		return b.ID
	}
	transition := func(id int64, target string) (string, int) {
		body, _ := json.Marshal(map[string]string{"target": target})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(id, 10)+"/transition", bytes.NewReader(body)))
		var b struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &b)
		return b.Status, rec.Code
	}
	seal := func(id int64) (string, int) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(id, 10)+"/seal", nil))
		var b struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &b)
		return b.Status, rec.Code
	}

	// confirmed -> sealed 必须推进到 sealed。
	id1 := createBatch("confirmed-then-seal")
	transition(id1, "running")
	transition(id1, "pending_diagnosis")
	transition(id1, "confirmed")
	if status, code := seal(id1); status != "sealed" || code != http.StatusOK {
		t.Fatalf("confirmed seal = %q (HTTP %d), want sealed / 200", status, code)
	}

	// running -> sealed 直接封存的合法路径。
	id2 := createBatch("running-then-seal")
	transition(id2, "running")
	if status, code := seal(id2); status != "sealed" || code != http.StatusOK {
		t.Fatalf("running seal = %q (HTTP %d), want sealed / 200", status, code)
	}

	// 已封存再次封存保持幂等（仍为 sealed）。
	if status, code := seal(id2); status != "sealed" || code != http.StatusOK {
		t.Fatalf("reseal = %q (HTTP %d), want sealed / 200", status, code)
	}

	// 准备态不得直接封存。
	id3 := createBatch("preparing-no-seal")
	if _, code := seal(id3); code != http.StatusConflict {
		t.Fatalf("preparing seal status = %d, want 409", code)
	}
}
