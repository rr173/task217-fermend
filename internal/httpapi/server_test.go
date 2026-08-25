package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TestExcludedChannelRejectedByDiagnosis 锁定剔除状态在接口、存储与诊断链路中的传播：
// 被剔除的通道不再参与代谢终点诊断；当批次唯一诊断通道被剔除时，诊断拒绝并返回
// 可处理的无有效通道结果（409 Conflict，ErrNoActiveDiagnosis），而非误报诊断成功。
func TestExcludedChannelRejectedByDiagnosis(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/excluded.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := New(service.New(s)).Handler()

	// 登记批次并推进到 running。
	body, _ := json.Marshal(map[string]string{"name": "excluded-batch", "strain": "S", "bioreactor": "BR"})
	created := httptest.NewRecorder()
	h.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewReader(body)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create batch status = %d, body = %s", created.Code, created.Body.String())
	}
	var batch struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(created.Body.Bytes(), &batch)
	tb, _ := json.Marshal(map[string]string{"target": "running"})
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(batch.ID, 10)+"/transition", bytes.NewReader(tb)))

	// 登记唯一的溶氧诊断通道并写入曲线。
	chBody, _ := json.Marshal(map[string]string{"name": "do", "kind": "do", "unit": "%"})
	chResp := httptest.NewRecorder()
	h.ServeHTTP(chResp, httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(batch.ID, 10)+"/channels", bytes.NewReader(chBody)))
	if chResp.Code != http.StatusCreated {
		t.Fatalf("register channel status = %d, body = %s", chResp.Code, chResp.Body.String())
	}
	var ch struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(chResp.Body.Bytes(), &ch)

	pts, _ := json.Marshal(map[string]any{"points": []map[string]any{
		{"seq": 0, "t_unix": 0, "value": 100.0},
		{"seq": 1, "t_unix": 100, "value": 30.0},
		{"seq": 2, "t_unix": 200, "value": 60.0},
	}})
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/batches/%d/channels/%d/samples", batch.ID, ch.ID), bytes.NewReader(pts)))

	// 剔除前：诊断应成功。
	ok := httptest.NewRecorder()
	h.ServeHTTP(ok, httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(batch.ID, 10)+"/diagnose", nil))
	if ok.Code != http.StatusOK {
		t.Fatalf("diagnose before exclude status = %d, body = %s", ok.Code, ok.Body.String())
	}

	// 剔除该通道：接口应反映剔除状态。
	exc := httptest.NewRecorder()
	h.ServeHTTP(exc, httptest.NewRequest(http.MethodPost, "/api/channels/"+strconv.FormatInt(ch.ID, 10)+"/exclude", nil))
	if exc.Code != http.StatusOK {
		t.Fatalf("exclude status = %d, body = %s", exc.Code, exc.Body.String())
	}
	if !bytes.Contains(exc.Body.Bytes(), []byte(`"excluded":true`)) {
		t.Fatalf("exclude response = %s, want excluded:true", exc.Body.String())
	}

	// 剔除后：诊断应拒绝并返回可处理的无有效通道结果（409 Conflict）。
	rej := httptest.NewRecorder()
	h.ServeHTTP(rej, httptest.NewRequest(http.MethodPost, "/api/batches/"+strconv.FormatInt(batch.ID, 10)+"/diagnose", nil))
	if rej.Code != http.StatusConflict {
		t.Fatalf("diagnose after exclude status = %d, want 409; body = %s", rej.Code, rej.Body.String())
	}
	if !bytes.Contains(rej.Body.Bytes(), []byte("no active diagnosis")) {
		t.Fatalf("diagnose after exclude body = %s, want no active diagnosis", rej.Body.String())
	}

	// 二次剔除同一通道应保持幂等且报已剔除冲突。
	again := httptest.NewRecorder()
	h.ServeHTTP(again, httptest.NewRequest(http.MethodPost, "/api/channels/"+strconv.FormatInt(ch.ID, 10)+"/exclude", nil))
	if again.Code != http.StatusConflict {
		t.Fatalf("re-exclude status = %d, want 409; body = %s", again.Code, again.Body.String())
	}
}

