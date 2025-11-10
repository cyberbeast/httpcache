package sqlite

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/cyberbeast/httpcache"
	_ "modernc.org/sqlite"
)

func TestTransport(t *testing.T) {
	delay := 3 * time.Second
	srv := httptest.NewServer(delayedResponse(delay))
	defer srv.Close()

	db, err := sql.Open("sqlite", path.Join("file://", t.TempDir(), "temp.db"))
	if err != nil {
		t.Fatalf("creating sqlite db: %v", err)
	}

	tr, err := httpcache.NewTransport(t.Context(), DB{db}, nil)
	if err != nil {
		t.Fatalf("couldn't initialize transport for test: %v", err)
	}

	client := &http.Client{Transport: tr}

	cold, err := measureDuration(get(client, srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	cached, err := measureDuration(get(client, srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	if cold <= (delay) {
		t.Fatalf("expected cold latency to be lower than or equal to %d, got %d\n", delay, cold)
	}

	speedupIsAboveMinRatio(t, cold, cached, 0.75)
}

func TestTransportInvalidateAllResponses(t *testing.T) {
	delay := 2 * time.Second
	srv := httptest.NewServer(delayedResponse(delay))
	defer srv.Close()

	db, err := sql.Open("sqlite", path.Join("file://", t.TempDir(), "temp.db"))
	if err != nil {
		t.Fatalf("creating sqlite db: %v", err)
	}

	tr, err := httpcache.NewTransport(t.Context(), DB{db}, nil)
	if err != nil {
		t.Fatalf("couldn't initialize transport for test: %v", err)
	}

	client := &http.Client{Transport: tr}

	_, err = measureDuration(get(client, srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = measureDuration(get(client, srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	if err := tr.InvalidateAllResponses(t.Context()); err != nil {
		t.Fatalf("couldn't invalidate all responses in the cache: %v", err)
	}

	cold, err := measureDuration(get(client, srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	if cold <= (delay) {
		t.Fatalf("expected cold latency to be lower than or equal to %d, got %d\n", delay, cold)
	}
}

func delayedResponse(delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Write([]byte("ok"))
	}
}

func speedupIsAboveMinRatio(t *testing.T, cold, cached time.Duration, minRatio float64) {
	t.Helper()

	ratio := float64(cold-cached) / float64(cached)
	if ratio <= minRatio {
		t.Fatalf("expected cached latency to be at least %.2fx faster than cold latency, got %.2fx\n", minRatio, ratio)
	}
}

func get(client *http.Client, url string) func() error {
	return func() error {
		resp, err := client.Get(url)
		if err != nil {
			return err
		}

		if _, err := io.ReadAll(resp.Body); err != nil {
			return err
		}

		if err := resp.Body.Close(); err != nil {
			return err
		}

		return nil
	}
}

func measureDuration(fn func() error) (time.Duration, error) {
	start := time.Now()
	latency := func() time.Duration { return time.Since(start) }

	if err := fn(); err != nil {
		return latency(), err
	}

	return latency(), nil
}
