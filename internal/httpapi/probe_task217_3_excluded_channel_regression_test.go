package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"task217-fermend/internal/httpapi"
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
)

func TestExcludedChannelDoesNotEnterDiagnosis(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/exclude.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	ch, err := svc.RegisterChannel(b.ID, "do", model.KindDissolvedOxygen, model.UnitPercent)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.IngestSamples(b.ID, ch.ID, []model.SamplePoint{{Seq: 1, TUnix: 10, Value: 1}}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	httpapi.New(svc).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/channels/1/exclude", bytes.NewReader(nil)))
	if rec.Code != http.StatusOK {
		t.Fatalf("exclude response = %d %s", rec.Code, rec.Body.String())
	}
	_, _, err = svc.Diagnose(b.ID)
	if err == nil {
		t.Fatal("diagnosis succeeded after the only channel was excluded")
	}
}
