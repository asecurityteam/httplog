package stridelog

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"bitbucket.org/atlassian/logevent"
	"github.com/rs/xlog"
)

type ctxKey string

var (
	ctxKeyTransactionID = ctxKey("_stridelog_transaction_id")
	ctxKeyBase          = ctxKey("_stridelog_base")
)

type recordingReader struct {
	io.ReadCloser
	bytesRead *int32
}

func (r *recordingReader) BytesRead() int {
	return int(atomic.LoadInt32(r.bytesRead))
}

func (r *recordingReader) Read(p []byte) (int, error) {
	var n, e = r.ReadCloser.Read(p)
	atomic.AddInt32(r.bytesRead, int32(n))
	return n, e
}

// Middleware wraps an HTTP handler with Stride standard access logs and
// provides, via context, tools for constructing higher level log events that
// contain the Stride standard attributes.
type Middleware struct {
	service        string
	version        string
	host           string
	env            string
	tags           map[string]interface{}
	level          xlog.Level
	patchSTDLib    bool
	outputStream   xlog.Output
	conf           xlog.Config
	requestID      func(*http.Request) string
	transactionID  func(context.Context) string
	xlogMiddleware func(http.Handler) http.Handler
	next           http.Handler
}

func (m *Middleware) serveHTTP(w http.ResponseWriter, r *http.Request) {
	var srcIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	var dstIP, dstPortStr, _ = net.SplitHostPort(r.Context().Value(http.LocalAddrContextKey).(net.Addr).String())
	var dstPort, _ = strconv.Atoi(dstPortStr)
	var base = Base{
		Service:   m.service,
		Version:   m.version,
		Host:      m.host,
		Env:       m.env,
		RequestID: m.requestID(r),
	}
	var access = Access{
		Base:          base,
		SourceIP:      srcIP,
		ForwardedFor:  r.Header.Get("X-Forwarded-For"),
		DestinationIP: dstIP,
		Site:          r.Host,
		HTTPRequestContentType: r.Header.Get("Content-Type"),
		HTTPMethod:             r.Method,
		HTTPReferrer:           r.Referer(),
		HTTPUserAgent:          r.UserAgent(),
		URIPath:                r.URL.Path,
		URIQuery:               r.URL.Query().Encode(),
		Scheme:                 r.URL.Scheme,
		Port:                   dstPort,
	}

	r = r.WithContext(
		context.WithValue(
			context.WithValue(r.Context(), ctxKeyTransactionID, m.transactionID),
			ctxKeyBase,
			base,
		),
	)
	var wrapper = wrapWriter(w, r.ProtoMajor)
	var bodyWrapper = &recordingReader{r.Body, new(int32)}
	r.Body = bodyWrapper
	var start = time.Now()
	m.next.ServeHTTP(wrapper, r)
	access.Duration = int(time.Since(start).Nanoseconds() / 1e6)
	access.BytesOut = wrapper.BytesWritten()
	access.BytesIn = bodyWrapper.BytesRead()
	access.Bytes = access.BytesIn + access.BytesOut
	access.HTTPContentType = wrapper.Header().Get("Content-Type")
	access.Status = wrapper.Status()
	logevent.FromContext(r.Context()).Info(access)
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.xlogMiddleware(http.HandlerFunc(m.serveHTTP)).ServeHTTP(w, r)
}

// MiddlewareOption is used to configure the HTTP server middleware.
type MiddlewareOption func(*Middleware) *Middleware

// MiddlewareOptionTag applies a static key/value pair to all logs.
func MiddlewareOptionTag(tagName string, tagValue interface{}) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.tags[tagName] = tagValue
		return m
	}
}

// MiddlewareOptionService sets the name of the running service as it will
// appear in the logs. The default value is the hostname of the system.
func MiddlewareOptionService(name string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.service = name
		return m
	}
}

// MiddlewareOptionHost sets the name of the system host as it will
// appear in the logs. The default value is the self-reported hostname of the
// system.
func MiddlewareOptionHost(name string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.host = name
		return m
	}
}

// MiddlewareOptionVersion sets the version of the service as it will
// appear in the logs. The default value is "latest".
func MiddlewareOptionVersion(name string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.version = name
		return m
	}
}

// MiddlewareOptionEnv sets the name of the service environment as it will
// appear in the logs. The default value is "production".
func MiddlewareOptionEnv(name string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.env = name
		return m
	}
}

// MiddlewareOptionRequestID sets the function that is called on each incoming
// request to set the request_id field. The default value for this option
// is a function that returns a hex encoded zero value like "0000000000000000".
func MiddlewareOptionRequestID(requestID func(*http.Request) string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.requestID = requestID
		return m
	}
}

// MiddlewareOptionTransactionID sets the function that is called on creation of
// each new event within a request and is used to populate the value of
// transaction_id. The default value is a function that returns a hex encoded
// zero value like "0000000000000000".
func MiddlewareOptionTransactionID(transactionID func(context.Context) string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.transactionID = transactionID
		return m
	}
}

// levelFromString converts a string log level name into an xlog.Level type
// for use with xlog.
func levelFromString(level string) xlog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return xlog.LevelDebug
	case "INFO":
		return xlog.LevelInfo
	case "WARN":
		return xlog.LevelWarn
	case "ERROR":
		return xlog.LevelError
	case "FATAL":
		return xlog.LevelFatal
	default:
		return xlog.LevelDebug
	}
}

// MiddlewareOptionLevel sets the lowest logging level that will appear in the
// stream. The default value is INFO. Possible values are DEBUG, INFO, WARN,
// ERROR, and FATAL.
func MiddlewareOptionLevel(level string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.level = levelFromString(level)
		return m
	}
}

// MiddlewareOptionPatchSTDLib reconfigures the global, stdlib logger to use
// a structured output. Request based annotations will not be available in
// converted stdlib logs and events are only emitted at INFO level. By default,
// the stdlib log package is left as is.
func MiddlewareOptionPatchSTDLib(m *Middleware) *Middleware {
	m.patchSTDLib = true
	return m
}

// MiddlewareOptionConsole converts the default JSON output into a colourised,
// tab/space delimited output for better human readability.
func MiddlewareOptionConsole(m *Middleware) *Middleware {
	m.outputStream = xlog.NewConsoleOutputW(os.Stdout, xlog.NewLogfmtOutput(os.Stdout))
	return m
}

// NewMiddleware generates an HTTP handler wrapper that performs access logging
// and injects a partial Event object into the context for later use.
func NewMiddleware(options ...MiddlewareOption) func(http.Handler) http.Handler {
	var hostname, _ = os.Hostname()
	return func(next http.Handler) http.Handler {
		var m = &Middleware{
			service:       hostname,
			version:       "latest",
			host:          hostname,
			env:           "production",
			tags:          make(map[string]interface{}),
			level:         xlog.LevelInfo,
			patchSTDLib:   false,
			outputStream:  xlog.NewJSONOutput(os.Stdout),
			requestID:     func(*http.Request) string { return fmt.Sprintf("%X", int64(0)) },
			transactionID: func(context.Context) string { return fmt.Sprintf("%X", int64(0)) },
			next:          next,
		}
		for _, option := range options {
			m = option(m)
		}
		m.conf = xlog.Config{
			Level:  m.level,
			Fields: xlog.F(m.tags),
			Output: xlog.OutputFunc(func(fields map[string]interface{}) error {
				return m.outputStream.Write(fields)
			}),
			DisablePooling: true,
		}
		if m.patchSTDLib {
			log.SetFlags(0)
			log.SetOutput(xlog.New(m.conf))
		}
		m.xlogMiddleware = xlog.NewHandler(m.conf)
		return m
	}
}

// NewEvent generates a partially populated Event object with data from the
// context.
func NewEvent(ctx context.Context) Event {
	return Event{
		Base:          ctx.Value(ctxKeyBase).(Base),
		TransactionID: ctx.Value(ctxKeyTransactionID).(func(context.Context) string)(ctx),
	}
}

// OutOfBand returns a context with all of the configuration provided to the
// middleware. This is provided with the primary intent of allowing for log
// emissions during runtime setup (such as main.go) and background routines that
// are not attached to a request or request context.
func OutOfBand(ctx context.Context, middleware func(http.Handler) http.Handler) context.Context {
	if m, ok := middleware(nil).(*Middleware); ok {
		ctx = xlog.NewContext(ctx, xlog.New(m.conf))
	}
	return ctx
}
