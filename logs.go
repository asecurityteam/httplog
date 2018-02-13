package stridelogs

// Base implements the core, developer schema that all events share.
type Base struct {
	Service   string   `logevent:"service"`
	Schema    string   `logevent:"schema,default=developer"`
	UGCDirty  []string `logevent:"ugc_dirty"`
	Version   string   `logevent:"version"`
	Host      string   `logevent:"host"`
	Env       string   `logevent:"env"`
	Time      string   `logevent:"time"`
	RequestID string   `logevent:"request_id"`
}

// Access implements the access log schema.
type Access struct {
	Base
	Schema                 string `logevent:"schema,default=access"`
	SourceIP               string `logevent:"src_ip"`
	ForwardedFor           string `logevent:"forwarded_for"`
	DestinationIP          string `logevent:"dest_ip"`
	Site                   string `logevent:"site"`
	HTTPRequestContentType string `logevent:"http_request_content_type"`
	HTTPMethod             string `logevent:"http_method"`
	HTTPReferrer           string `logevent:"http_referrer"`
	HTTPUserAgent          string `logevent:"http_user_agent"`
	URIPath                string `logevent:"uri_path"`
	URIQuery               string `logevent:"uri_query"`
	Scheme                 string `logevent:"scheme"`
	Port                   int    `logevent:"port"`
	Bytes                  int    `logevent:"bytes"`
	BytesOut               int    `logevent:"bytes_out"`
	BytesIn                int    `logevent:"bytes_in"`
	Duration               int    `logevent:"duration"`
	HTTPContentType        string `logevent:"http_content_type"`
	Status                 int    `logevent:"status"`
	Message                string `logevent:"message,default=access"`
}

// Event implements the schema for all service events. It can be embedded within
// a richer schema to create compliant service logs.
type Event struct {
	Base
	Schema        string `logevent:"schema,default=event"`
	User          string `logevent:"user"`
	ObjectID      string `logevent:"object_id"`
	ObjectType    string `logevent:"object_type"`
	Status        int    `logevent:"status"`
	Result        string `logevent:"result"`
	Action        string `logevent:"action"`
	TransactionID string `logevent:"transaction_id"`
	SessionID     string `logevent:"session_id"`
}
