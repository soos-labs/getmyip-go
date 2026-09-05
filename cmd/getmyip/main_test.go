package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cliServer поднимает мок API и наводит на него CLI через GETMYIP_BASE_URL.
func cliServer(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ip":"45.82.64.40","country":"Netherlands","country_code":"NL","org":"WorldStream B.V.","asn":49981,"timezone":"Europe/Amsterdam"}`))
	})
	mux.HandleFunc("GET /v1/8.8.8.8", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ip":"8.8.8.8","country":"United States","country_code":"US","org":"Google LLC","asn":15169}`))
	})
	mux.HandleFunc("GET /v1/bad", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid ip"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("GETMYIP_BASE_URL", srv.URL)
}

func TestRunHuman(t *testing.T) {
	cliServer(t)
	var out, errb strings.Builder
	if code := run([]string{"8.8.8.8"}, &out, &errb); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errb.String())
	}
	for _, want := range []string{"IP:       8.8.8.8", "United States (US)", "Google LLC (AS15169)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunMe(t *testing.T) {
	cliServer(t)
	var out, errb strings.Builder
	if code := run(nil, &out, &errb); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "45.82.64.40") {
		t.Errorf("me output: %s", out.String())
	}
}

func TestRunJSONAndField(t *testing.T) {
	cliServer(t)
	var out, errb strings.Builder
	if code := run([]string{"-json", "8.8.8.8"}, &out, &errb); code != 0 {
		t.Fatalf("json: code=%d", code)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out.String()), &m); err != nil || m["ip"] != "8.8.8.8" {
		t.Errorf("-json output invalid: %v %s", err, out.String())
	}

	out.Reset()
	if code := run([]string{"-field", "org", "8.8.8.8"}, &out, &errb); code != 0 {
		t.Fatalf("field: code=%d", code)
	}
	if strings.TrimSpace(out.String()) != "Google LLC" {
		t.Errorf("-field org = %q", out.String())
	}
}

func TestRunErrors(t *testing.T) {
	cliServer(t)
	var out, errb strings.Builder
	if code := run([]string{"bad"}, &out, &errb); code != 1 {
		t.Errorf("api error: code=%d, want 1", code)
	}
	if !strings.Contains(errb.String(), "invalid ip") {
		t.Errorf("stderr: %s", errb.String())
	}

	errb.Reset()
	if code := run([]string{"-field", "nope", "8.8.8.8"}, &out, &errb); code != 2 {
		t.Errorf("unknown field: code=%d, want 2", code)
	}
}
