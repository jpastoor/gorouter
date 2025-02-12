package handlers

import (
	"fmt"
	"net/http"
	"strings"

	router_http "code.cloudfoundry.org/gorouter/common/http"
	"code.cloudfoundry.org/gorouter/config"
	"code.cloudfoundry.org/gorouter/route"
)

const (
	cacheMaxAgeSeconds = 2
	VcapCookieId       = "__VCAP_ID__"
)

func AddRouterErrorHeader(rw http.ResponseWriter, val string) {
	rw.Header().Set(router_http.CfRouterError, val)
}

func addInvalidResponseCacheControlHeader(rw http.ResponseWriter) {
	rw.Header().Set(
		"Cache-Control",
		fmt.Sprintf("public,max-age=%d", cacheMaxAgeSeconds),
	)
}

func addNoCacheControlHeader(rw http.ResponseWriter) {
	rw.Header().Set(
		"Cache-Control",
		"no-cache, no-store",
	)
}

func hostWithoutPort(reqHost string) string {
	host := reqHost

	// Remove :<port>
	pos := strings.Index(host, ":")
	if pos >= 0 {
		host = host[0:pos]
	}

	return host
}

func IsWebSocketUpgrade(request *http.Request) bool {
	// websocket should be case insensitive per RFC6455 4.2.1
	return strings.ToLower(upgradeHeader(request)) == "websocket"
}

func upgradeHeader(request *http.Request) string {
	// handle multiple Connection field-values, either in a comma-separated string or multiple field-headers
	for _, v := range request.Header[http.CanonicalHeaderKey("Connection")] {
		// upgrade should be case insensitive per RFC6455 4.2.1
		if strings.Contains(strings.ToLower(v), "upgrade") {
			return request.Header.Get("Upgrade")
		}
	}

	return ""
}

func EndpointIteratorForRequest(request *http.Request, loadBalanceMethod string, stickySessionCookieNames config.StringSet) (route.EndpointIterator, error) {
	reqInfo, err := ContextRequestInfo(request)
	if err != nil {
		return nil, fmt.Errorf("could not find reqInfo in context")
	}
	return reqInfo.RoutePool.Endpoints(loadBalanceMethod, getStickySession(request, stickySessionCookieNames)), nil
}

func getStickySession(request *http.Request, stickySessionCookieNames config.StringSet) string {
	// Try choosing a backend using sticky session
	for stickyCookieName, _ := range stickySessionCookieNames {
		if _, err := request.Cookie(stickyCookieName); err == nil {
			if sticky, err := request.Cookie(VcapCookieId); err == nil {
				return sticky.Value
			}
		}
	}
	return ""
}
