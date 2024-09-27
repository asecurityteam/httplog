package httplog

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asecurityteam/logevent/v2"
	"github.com/golang/mock/gomock"
)

type fixtureHandler struct{}

func (fixtureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestMiddlewareOptions(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var logger = NewMockLogger(ctrl)
	var result = NewMiddleware(
		MiddlewareOptionTag("test", "test"),
		MiddlewareOptionHost("host"),
		MiddlewareOptionService("service"),
		MiddlewareOptionVersion("version"),
		MiddlewareOptionEnv("env"),
		MiddlewareOptionRedactParameter("test"),
		MiddlewareOptionRequestID(func(*http.Request) string { return "reqid" }),
	)
	var m = result(fixtureHandler{}).(*Middleware)
	var req = httptest.NewRequest(http.MethodGet, "/?test=something&test2=something", io.NopCloser(bytes.NewBufferString(``)))
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}))
	req = req.WithContext(logevent.NewContext(req.Context(), logger))

	logger.EXPECT().SetField("test", "test")
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt Access
		var ok bool
		if evt, ok = event.(Access); !ok {
			t.Error("middleware did not perform an access log")
		}
		if evt.Service != "service" {
			t.Fatalf("MiddlewareOptionService did not update log annotations, %v", evt)
		}
		if evt.Host != "host" {
			t.Fatalf("MiddlewareOptionHost did not update log annotations, %v", evt)
		}
		if evt.Version != "version" {
			t.Fatalf("MiddlewareOptionVersion did not update log annotations, %v", evt)
		}
		if evt.Env != "env" {
			t.Fatalf("MiddlewareOptionEnv did not update log annotations, %v", evt)
		}
		if evt.RequestID != "reqid" {
			t.Fatalf("MiddlewareOptionRequestID did not update log annotations, %v", evt)
		}
		if evt.URIQuery != "test=REDACTED&test2=something" {
			t.Fatalf("MiddlewareOptionRedactParameter did not redact parameter value, %v", evt)
		}
	})
	m.ServeHTTP(httptest.NewRecorder(), req)
}

type fixtureHandlerTransactionID struct{}

func (fixtureHandlerTransactionID) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logevent.FromContext(r.Context()).Info(NewEvent(r.Context()))
}

func TestMiddlewareOptionTransactionID(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var logger = NewMockLogger(ctrl)
	var result = NewMiddleware(MiddlewareOptionTransactionID(func(context.Context) string { return "test" }))
	var m = result(fixtureHandlerTransactionID{}).(*Middleware)
	var req = httptest.NewRequest(http.MethodGet, "/", io.NopCloser(bytes.NewBufferString(``)))
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.IPAddr{Zone: "", IP: net.ParseIP("127.0.0.1")}))
	req = req.WithContext(logevent.NewContext(req.Context(), logger))

	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt Event
		var ok bool
		if evt, ok = event.(Event); !ok {
			t.Error("handler did not log an Event")
		}
		if evt.TransactionID != "test" {
			t.Fatalf("MiddlewareOptionTransactionID did not update log annotations, %v", evt)
		}
	})
	logger.EXPECT().Info(gomock.Any())
	m.ServeHTTP(httptest.NewRecorder(), req)
}
