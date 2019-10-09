<a id="markdown-httplog---structured-access-log-types-for-http-services" name="httplog---structured-access-log-types-for-http-services"></a>
# httplog - structured access log types for HTTP services. #
[![GoDoc](https://godoc.org/github.com/asecurityteam/httplog?status.svg)](https://godoc.org/github.com/asecurityteam/httplog)
[![Build Status](https://travis-ci.org/asecurityteam/httplog.png?branch=master)](https://travis-ci.org/asecurityteam/httplog)
[![codecov.io](https://codecov.io/github/asecurityteam/httplog/coverage.svg?branch=master)](https://codecov.io/github/asecurityteam/httplog?branch=master)

*Status: Production*

<!-- TOC -->

- [httplog - structured access log types for HTTP services.](#httplog---structured-access-log-types-for-http-services)
    - [Overview](#overview)
    - [Usage](#usage)
    - [Contributing](#contributing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview ##

Atlassian development teams often operate multiple dozens of individual services
spanning up to hundreds of instances. To help with operations, we've aligned
with our company's global logging standard. This standard defines the structure
of most logs in addition to defining the kinds of logs that should be emitted
by services.

This packages exports several structs that represent the structured log events
defined by Atlassian's specification:

- `httplog.Base`

  This represents the minimum, common base of all log events. All events must
  contain the fields defined in this struct.

- `httplog.Access`

  This defines the fields needed in all system access logs. These are modelled
  after HTTP request logs available from other systems like Apache or NGINX.

- `httplog.Event`

  This struct represents the base for all developer defined system events. This
  struct should be embedded in all logs emitted by a service.

<a id="markdown-usage" name="usage"></a>
## Usage ##

This package is primarily focused on providing a logging harness around HTTP
services and exposes a wrapper for `http.Handler` that provides this feature.

```golang
var middleware = httplog.NewMiddleware()
http.ListenAndServer(":8080", middleware(http.DefaultServeMux))
```

The default behaviour of the middleware is to provide Atlassian logging spec
compliant access logs for your HTTP service. Quite a few default settings are
selected which can be overridden with the available function arguments.

```golang
var middleware = httplog.NewMiddleware(
  MiddlewareOptionTag("key", "value"), // Add arbitrary annotations to all logs.
  MiddlewareOptionService("myService"), // Set the service name field to a custom value.
  MiddlewareOptionHost("customHost"), // Set the host field to something other than the system hostname.
  MiddlewareOptionVersion("1.2.3"), // Set the service version that is active.
  MiddlewareOptionEnv("staging"), // Set the environment the service is running in.
  // Set the function used to populate the request_id field in all log events.
  MiddlewareOptionRequestID(func(r *http.Request) string { return httptrace.TraceIDFromContext(r.Context()) }),
  // Set the function used to populate the transaction_id field in all developer events.
  MiddlewareOptionTransactionID(func(r *http.Request) string { return httptrace.SpanIDFromContext(r.Context()) }),
  MiddlewareOptionLevel("DEBUG"), // Set the minimum log level to be emitted.
  MiddlewareOptionPatchSTDLib, // Reconfigure `log.Println`, etc., to use JSON format.
  MiddlewareOptionConsole, // Disable JSON in favour of a human readable line format.
)
```

<a id="markdown-contributing" name="contributing"></a>
## Contributing ##

<a id="markdown-license" name="license"></a>
### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
