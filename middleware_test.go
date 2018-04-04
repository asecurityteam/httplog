package stridelog

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bitbucket.org/atlassian/logevent"
	"github.com/golang/mock/gomock"
	"github.com/rs/xlog"
)

type fixtureHandler struct{}

func (fixtureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func middlewareOptionOutput(output xlog.Output) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.outputStream = output
		return m
	}
}

func TestMiddlewareOptions(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var output = NewMockOutput(ctrl)
	var result = NewMiddleware(
		middlewareOptionOutput(output),
		MiddlewareOptionTag("test", "test"),
		MiddlewareOptionHost("host"),
		MiddlewareOptionService("service"),
		MiddlewareOptionVersion("version"),
		MiddlewareOptionEnv("env"),
		MiddlewareOptionRequestID(func(*http.Request) string { return "reqid" }),
	)
	var m = result(fixtureHandler{}).(*Middleware)
	var req = httptest.NewRequest(http.MethodGet, "/", ioutil.NopCloser(bytes.NewBufferString(``)))
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}))

	output.EXPECT().Write(gomock.Any()).Do(func(tags map[string]interface{}) {
		if value, ok := tags["test"]; !ok || value != "test" {
			t.Fatal("MiddlewareOptionTag did not annotate logs")
		}
		if value, ok := tags["service"]; !ok || value != "service" {
			t.Fatalf("MiddlewareOptionService did not update log annotations, %v", tags)
		}
		if value, ok := tags["host"]; !ok || value != "host" {
			t.Fatalf("MiddlewareOptionHost did not update log annotations, %v", tags)
		}
		if value, ok := tags["version"]; !ok || value != "version" {
			t.Fatalf("MiddlewareOptionVersion did not update log annotations, %v", tags)
		}
		if value, ok := tags["env"]; !ok || value != "env" {
			t.Fatalf("MiddlewareOptionEnv did not update log annotations, %v", tags)
		}
		if value, ok := tags["request_id"]; !ok || value != "reqid" {
			t.Fatalf("MiddlewareOptionRequestID did not update log annotations, %v", tags)
		}
	})
	m.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddlewareOptionLevel(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var output = NewMockOutput(ctrl)
	var result = NewMiddleware(middlewareOptionOutput(output), MiddlewareOptionLevel("ERROR"))
	var m = result(fixtureHandler{}).(*Middleware)
	var req = httptest.NewRequest(http.MethodGet, "/", ioutil.NopCloser(bytes.NewBufferString(``)))
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}))

	output.EXPECT().Write(gomock.Any()).Times(0)
	m.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddlewareOptionPatchSTDLib(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	defer log.SetFlags(0)
	defer log.SetOutput(os.Stderr)

	var output = NewMockOutput(ctrl)
	var result = NewMiddleware(middlewareOptionOutput(output), MiddlewareOptionPatchSTDLib)
	var _ = result(fixtureHandler{}).(*Middleware)
	output.EXPECT().Write(gomock.Any()).Do(func(tags map[string]interface{}) {
		if value, ok := tags["message"]; !ok || value != "test" {
			t.Fatalf("MiddlewareOptionPatchSTDLib did not patch the std lib, %v", tags)
		}
	})
	log.Print("test")
}

type fixtureHandlerTransactionID struct{}

func (fixtureHandlerTransactionID) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logevent.FromContext(r.Context()).Info(NewEvent(r.Context()))
}

func TestMiddlewareOptionTransactionID(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var output = NewMockOutput(ctrl)
	var result = NewMiddleware(middlewareOptionOutput(output), MiddlewareOptionTransactionID(func(context.Context) string { return "test" }))
	var m = result(fixtureHandlerTransactionID{}).(*Middleware)
	var req = httptest.NewRequest(http.MethodGet, "/", ioutil.NopCloser(bytes.NewBufferString(``)))
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}))

	output.EXPECT().Write(gomock.Any()).Do(func(tags map[string]interface{}) {
		if value, ok := tags["transaction_id"]; !ok || value != "test" {
			t.Fatalf("MiddlewareOptionEnv did not update log annotations, %v", tags)
		}
	})
	output.EXPECT().Write(gomock.Any())
	m.ServeHTTP(httptest.NewRecorder(), req)
}
