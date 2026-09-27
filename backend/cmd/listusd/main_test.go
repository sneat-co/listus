package main

import (
	"errors"
	"net/http"
	"os"
	"testing"
)

func TestMain_Success(t *testing.T) {
	origListen := listenAndServe
	origFatal := logFatalf
	t.Cleanup(func() {
		listenAndServe = origListen
		logFatalf = origFatal
	})

	_ = os.Unsetenv("LISTUS_ADDR")
	calledAddr := ""
	listenAndServe = func(addr string, _ http.Handler) error {
		calledAddr = addr
		return nil
	}

	main()

	if calledAddr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", calledAddr)
	}
}

func TestMain_Failure(t *testing.T) {
	origListen := listenAndServe
	origFatal := logFatalf
	t.Cleanup(func() {
		listenAndServe = origListen
		logFatalf = origFatal
	})

	_ = os.Setenv("LISTUS_ADDR", ":9999")
	calledAddr := ""
	listenAndServe = func(addr string, _ http.Handler) error {
		calledAddr = addr
		return errors.New("listen error")
	}

	fatalCalled := false
	logFatalf = func(format string, v ...any) {
		fatalCalled = true
	}

	main()

	if calledAddr != ":9999" {
		t.Fatalf("expected addr :9999, got %q", calledAddr)
	}
	if !fatalCalled {
		t.Fatal("expected logFatalf to be called")
	}
}
