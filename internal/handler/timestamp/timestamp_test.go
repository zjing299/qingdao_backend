package timestamp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qingdao_backend/internal/client"
	"qingdao_backend/internal/config"
	logic "qingdao_backend/internal/logic/timestamp"
	"qingdao_backend/internal/types"
)

type fakeProvider struct {
	algorithm    string
	rid          string
	verifyResult types.VerifyResult
}

func (f *fakeProvider) RSA(_ context.Context, rid string) (*types.ApplyResult, error) {
	f.algorithm, f.rid = "SHA256", rid
	return &types.ApplyResult{RID: rid, Algorithm: f.algorithm, SN: "test", TS: 1, TSR: "AQ=="}, nil
}
func (f *fakeProvider) SM(_ context.Context, rid string) (*types.ApplyResult, error) {
	f.algorithm, f.rid = "SM3", rid
	return &types.ApplyResult{RID: rid, Algorithm: f.algorithm, SN: "test", TS: 1, TSR: "AQ=="}, nil
}
func (f *fakeProvider) Verify(_ context.Context, req types.VerifyRequest) (*types.VerifyResult, error) {
	f.rid = req.RID
	result := f.verifyResult
	return &result, nil
}

func post(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestApplyRoutesPassDigestWithoutHashing(t *testing.T) {
	for _, algorithm := range []string{"SHA256", "SM3"} {
		t.Run(algorithm, func(t *testing.T) {
			provider := &fakeProvider{}
			w := post(Apply(logic.New(provider), algorithm), `{"rid":"`+strings.Repeat("AB", 32)+`"}`)
			if w.Code != http.StatusOK || provider.algorithm != algorithm || provider.rid != strings.Repeat("ab", 32) {
				t.Fatalf("status=%d algorithm=%s rid=%s", w.Code, provider.algorithm, provider.rid)
			}
		})
	}
}

func TestApplyRejectsInvalidRequestsBeforeCallingUpstream(t *testing.T) {
	for _, body := range []string{
		`{"rid":"short"}`, `{"rid":"` + strings.Repeat("g", 64) + `"}`,
		`{"rid":"` + strings.Repeat("a", 64) + `","appsecret":"secret"}`,
		`{"rid":"` + strings.Repeat("a", 64) + `"} {}`, `null`, `[]`,
	} {
		provider := &fakeProvider{}
		w := post(Apply(logic.New(provider), "SHA256"), body)
		if w.Code != http.StatusBadRequest || provider.rid != "" {
			t.Fatalf("body=%s status=%d", body, w.Code)
		}
	}
}

func TestUnconfiguredAndPendingUpstreamAreUnavailable(t *testing.T) {
	for _, tc := range []struct {
		config config.TimestampConf
		code   string
	}{
		{config.TimestampConf{}, "TIMESTAMP_NOT_CONFIGURED"},
		{config.TimestampConf{BaseURL: "https://unused.invalid"}, "TIMESTAMP_INTEGRATION_PENDING"},
	} {
		service := logic.New(client.NewTimestampClient(tc.config))
		w := post(Apply(service, "SHA256"), `{"rid":"`+strings.Repeat("a", 64)+`"}`)
		var response types.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if w.Code != http.StatusServiceUnavailable || response.Code != tc.code {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestVerifyRequiresValidImprintAndChecksUpstreamDigest(t *testing.T) {
	rid := strings.Repeat("a", 64)
	for _, tc := range []struct {
		name     string
		upstream types.VerifyResult
		valid    bool
	}{
		{"match", types.VerifyResult{Valid: true, RID: rid, Algorithm: "SHA256"}, true},
		{"different digest", types.VerifyResult{Valid: true, RID: strings.Repeat("b", 64), Algorithm: "SHA256"}, false},
		{"different algorithm", types.VerifyResult{Valid: true, RID: rid, Algorithm: "SM3"}, false},
		{"invalid signature", types.VerifyResult{Valid: false, RID: rid, Algorithm: "SHA256"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &fakeProvider{verifyResult: tc.upstream}
			w := post(Verify(logic.New(provider)), `{"rid":"`+rid+`","algorithm":"SHA256","tsr":"AQ=="}`)
			var response struct {
				Data types.VerifyResult `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusOK || response.Data.Valid != tc.valid || provider.rid != rid {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	for _, body := range []string{
		`{"rid":"` + rid + `","algorithm":"SHA256","tsr":"not base64"}`,
		`{"rid":"` + rid + `","algorithm":"MD5","tsr":"AQ=="}`,
		`{"rid":"` + rid + `","algorithm":"SM3","tsr":""}`,
	} {
		provider := &fakeProvider{}
		w := post(Verify(logic.New(provider)), body)
		if w.Code != http.StatusBadRequest || provider.rid != "" {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestRejectsFileUploadsAndOversizedRequests(t *testing.T) {
	handler := Apply(logic.New(&fakeProvider{}), "SHA256")
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("file bytes"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status=%d", w.Code)
	}
	w = post(handler, `{"rid":"`+strings.Repeat("a", maxBodyBytes)+`"}`)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d", w.Code)
	}
}
