package postgres

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyberbeast/httpcache"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	pg "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func delayedResponse(delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Write([]byte("ok"))
	}
}

func startContainer(t *testing.T, username, pw string) (*pg.PostgresContainer, func()) {
	container, err := pg.Run(t.Context(), "postgres:16-alpine", pg.WithUsername(username), pg.WithPassword(pw), pg.BasicWaitStrategies())
	if err != nil {
		t.Fatalf("starting postgres testcontainer: %v", err)
	}

	return container, func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("terminating postgres testcontainer: %v", err)
		}
	}
}

func TestTransport(t *testing.T) {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	container, fn := startContainer(t, "user", "password")
	defer fn()

	connstr, err := container.ConnectionString(t.Context(), "sslmode=disable")
	if err != nil {
		t.Fatalf("generating postgres connection string: %v", err)
	}

	delay := 3 * time.Second
	srv := httptest.NewServer(delayedResponse(delay))
	defer srv.Close()

	db, err := pgx.Connect(t.Context(), connstr)
	if err != nil {
		t.Fatalf("connecting to postgres database: %v", err)
	}

	tr, err := httpcache.NewTransport(t.Context(), Connection{db}, nil)
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
