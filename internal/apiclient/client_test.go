package apiclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testPayload struct {
	Name string `json:"name"`
}

// newTestClient returns a Client pointed at the given handler.
func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	return NewClient(srv.URL, "xrp_test-token", nil), srv
}

func TestDoEnvelopeSuccess(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer xrp_test-token" {
			t.Errorf("Authorization = %q, want Bearer xrp_test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"name":"ok"}}`))
	})
	defer srv.Close()

	got, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	if err != nil {
		t.Fatalf("doEnvelope returned error: %v", err)
	}
	if got.Name != "ok" {
		t.Fatalf("data = %+v, want {Name:ok}", got)
	}
}

func TestDoEnvelopeSuccessFalseWithoutErrorIsError(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":null,"data":{"name":"ghost"}}`))
	})
	defer srv.Close()

	got, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	if err == nil {
		t.Fatalf("doEnvelope returned nil error with data %+v: success:false without error must not silently return the zero value", got)
	}
	if got.Name != "" {
		t.Fatalf("data = %+v, want zero value on error", got)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.Code != ErrInternal {
		t.Fatalf("code = %q, want %q", apiErr.Code, ErrInternal)
	}
}

func TestDoEnvelopeSuccessFalseWithError(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"VALIDATION_ERROR","message":"bad input"}}`))
	})
	defer srv.Close()

	_, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.Code != ErrValidation || apiErr.Message != "bad input" {
		t.Fatalf("apiErr = %+v, want code VALIDATION_ERROR message 'bad input'", apiErr)
	}
}

func TestDoEnvelopeUnauthorizedMaps401(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","message":"token expired"}}`))
	})
	defer srv.Close()

	_, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("HTTPStatus = %d, want 401", apiErr.HTTPStatus)
	}
	if apiErr.Code != ErrUnauthorized || apiErr.Message != "token expired" {
		t.Fatalf("apiErr = %+v, want code UNAUTHORIZED message 'token expired'", apiErr)
	}
}

func TestDoEnvelopeNon200PlainBodyAsMessage(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`upstream exploded`))
	})
	defer srv.Close()

	_, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("HTTPStatus = %d, want 500", apiErr.HTTPStatus)
	}
	if apiErr.Message != "upstream exploded" {
		t.Fatalf("message = %q, want raw body as message", apiErr.Message)
	}
}

func TestDoEnvelopeNon200EnvelopeErrorWinsOverBody(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","message":"no such user"}}`))
	})
	defer srv.Close()

	_, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.HTTPStatus != http.StatusNotFound || apiErr.Code != ErrNotFound {
		t.Fatalf("apiErr = %+v, want 404 NOT_FOUND", apiErr)
	}
}

func TestDoEnvelopeUnparseableEnvelopeIsError(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{this is not json`))
	})
	defer srv.Close()

	_, err := doEnvelope[testPayload](context.Background(), c, http.MethodGet, "/api/v1/ping", nil)
	if err == nil {
		t.Fatal("doEnvelope returned nil error for unparseable envelope")
	}
}

func TestHealthNon200MapsStatus(t *testing.T) {
	c, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`not ready`))
	})
	defer srv.Close()

	_, err := c.Health(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want *APIError", err, err)
	}
	if apiErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Fatalf("HTTPStatus = %d, want 503", apiErr.HTTPStatus)
	}
}
