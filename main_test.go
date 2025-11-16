package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

// TestHandler_Success simulates a successful call to the handler:
// - ?id=earth
// - upstream API returns mock JSON
// - handler should forward status/body and include CORS headers.
func TestHandler_Success(t *testing.T) {
	t.Helper()

	// 1. Start a mock upstream server that pretends to be the French API.
	apiKey := "TEST_KEY"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert that the handler sent the correct Authorization header.
		gotAuth := r.Header.Get("Authorization")
		wantAuth := "Bearer " + apiKey
		if gotAuth != wantAuth {
			t.Fatalf("expected Authorization %q, got %q", wantAuth, gotAuth)
		}

		// Assert that the handler requested the right path, e.g. /earth.
		if r.URL.Path != "/earth" {
			t.Fatalf("expected path /earth, got %q", r.URL.Path)
		}

		// Return a fake JSON response like the real API would.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"earth","isPlanet":true}`))
	}))
	defer ts.Close()

	// 2. Configure environment variables for the handler.
	t.Setenv("API_KEY", apiKey)
	t.Setenv("UPSTREAM_BASE_URL", ts.URL)

	// 3. Build a fake API Gateway request with ?id=earth.
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"id": "earth",
		},
	}

	// 4. Call the handler just like Lambda would.
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// 5. Assert that the handler forwarded the upstream status code.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 6. Assert that the body is the JSON we returned from the mock server.
	const wantBody = `{"id":"earth","isPlanet":true}`
	if resp.Body != wantBody {
		t.Fatalf("unexpected body.\nwant: %s\ngot:  %s", wantBody, resp.Body)
	}

	// 7. Assert that CORS headers are present.
	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin '*', got %q", resp.Headers["Access-Control-Allow-Origin"])
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}

// TestHandler_MissingID verifies that we get a 400 when ?id is missing.
func TestHandler_MissingID(t *testing.T) {
	t.Setenv("API_KEY", "TEST_KEY")
	// UPSTREAM_BASE_URL not needed; handler should fail before calling upstream.

	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{}, // no "id"
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Fatalf("expected CORS header, got %q", resp.Headers["Access-Control-Allow-Origin"])
	}
}

// TestHandler_MissingAPIKey verifies that a missing API_KEY env var
// is treated as a server configuration error (500).
func TestHandler_MissingAPIKey(t *testing.T) {
	// Ensure API_KEY is not set.
	os.Unsetenv("API_KEY")

	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"id": "earth",
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}

	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Fatalf("expected CORS header, got %q", resp.Headers["Access-Control-Allow-Origin"])
	}
}

// TestHandler_Upstream500 verifies that if the upstream API returns a 500,
// the handler forwards that status code and body back to the caller,
// still including CORS headers.

func TestHandler_Upstream500(t *testing.T) {
	t.Helper()

	apiKey := "TEST_KEY"

	// 1. Mock upstream server that always responds with 500.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We still expect the Authorization header to be sent correctly.
		gotAuth := r.Header.Get("Authorization")
		wantAuth := "Bearer " + apiKey
		if gotAuth != wantAuth {
			t.Fatalf("expected Authorization %q, got %q", wantAuth, gotAuth)
		}

		// Path should still reflect the requested body id.
		if r.URL.Path != "/earth" {
			t.Fatalf("expected path /earth, got %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"upstream failure"}`))
	}))
	defer ts.Close()

	// 2. Configure handler env vars to use the mock server.
	t.Setenv("API_KEY", apiKey)
	t.Setenv("UPSTREAM_BASE_URL", ts.URL)

	// 3. Fake API Gateway request (?id=earth).
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"id": "earth",
		},
	}

	// 4. Invoke handler.
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// 5. Assert that the handler forwards the 500 status code.
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// 6. Assert that the body is exactly what the upstream returned.
	const wantBody = `{"error":"upstream failure"}`
	if resp.Body != wantBody {
		t.Fatalf("unexpected body.\nwant: %s\ngot:  %s", wantBody, resp.Body)
	}

	// 7. Assert CORS headers are still present on error.
	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin '*', got %q", resp.Headers["Access-Control-Allow-Origin"])
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}
