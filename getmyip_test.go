package getmyip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Result{IP: "45.82.64.40", IPVersion: 4, Country: "Netherlands", CountryCode: "NL", ASN: 49981, Org: "WorldStream B.V."})
	})
	mux.HandleFunc("GET /v1/8.8.8.8", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Result{IP: "8.8.8.8", IPVersion: 4, Country: "United States", CountryCode: "US", ASN: 15169, Org: "Google LLC"})
	})
	mux.HandleFunc("GET /v1/bad", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid ip"}`))
	})
	mux.HandleFunc("POST /v1/batch", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"ip":"8.8.8.8","country":"United States"},{"ip":"1.1.1.1","country":"Australia"}]}`))
	})
	return httptest.NewServer(mux)
}

func TestClient(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	c := New(WithBaseURL(srv.URL))
	ctx := context.Background()

	me, err := c.Me(ctx)
	if err != nil || me.CountryCode != "NL" || me.ASN != 49981 {
		t.Fatalf("Me: %+v err=%v", me, err)
	}

	r, err := c.Lookup(ctx, "8.8.8.8")
	if err != nil || r.Org != "Google LLC" {
		t.Fatalf("Lookup: %+v err=%v", r, err)
	}

	batch, err := c.Batch(ctx, []string{"8.8.8.8", "1.1.1.1"})
	if err != nil || len(batch) != 2 || batch[1].Country != "Australia" {
		t.Fatalf("Batch: %+v err=%v", batch, err)
	}

	if _, err := c.Lookup(ctx, "bad"); err == nil {
		t.Fatalf("Lookup(bad): expected error")
	}
}

func TestHeaders(t *testing.T) {
	var ua, key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua, key = r.Header.Get("User-Agent"), r.Header.Get("X-API-Key")
		_ = json.NewEncoder(w).Encode(Result{IP: "1.2.3.4"})
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL), WithAPIKey("k1"))
	if _, err := c.Me(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ua != "getmyip-go/"+Version {
		t.Errorf("User-Agent = %q, want getmyip-go/%s", ua, Version)
	}
	if key != "k1" {
		t.Errorf("X-API-Key = %q, want k1", key)
	}
}
