package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/vulcand/oxy/forward"
)

var resourceHeader = GetOpt("RESOURCE_HEADER", "")
var apiRegExp = regexp.MustCompile("^/v[\\d.]+/(\\w+)")

type HeaderRewriter struct {
}

func (r *HeaderRewriter) Rewrite(req *http.Request) {
	// Get all headers
	for name, _ := range req.Header {
		// clear off 'l5d-ctx-*' 'l5d-dtab' 'l5d-sample'
		lcName := strings.ToLower(name)

		if (lcName == "l5d-dtab") ||
			(lcName == "l5d-sample") ||
			(lcName == "dtab-local") ||
			(strings.HasPrefix(lcName, "l5d-ctx-")) {
			req.Header.Del(name)
		}
	}

	if resourceHeader != "" {
		path := req.URL.Path
		matches := apiRegExp.FindStringSubmatch(path)
		var resourceName string
		if len(matches) == 2 {
			resourceName = matches[1]
		} else {
			resourceName = "unknown"
		}

		req.Header.Add(resourceHeader, resource)

		log.Printf("Adding resource header %s for %s", resource, path)
	}
}

func main() {

	port := GetOpt("PORT", "443")
	linkerdHost := GetOpt("LINKERD_HOST", "")
	linkerdPort := GetOpt("LINKERD_PORT", "")
	sslCertFile := GetOpt("SSL_CERT_FILE", "/etc/ssl/tls.crt")
	sslKeyFile := GetOpt("SSL_KEY_FILE", "/etc/ssl/tls.key")

	uri := "http://" + linkerdHost + ":" + linkerdPort

	fwdUrl, err := url.ParseRequestURI(uri)

	// Forwards incoming requests to our linker strips out l5d-* headers, adds proper forwarding headers
	keepHost := forward.PassHostHeader(true)

	headerRewriter := forward.Rewriter(new(HeaderRewriter))

	fwd, err := forward.New(keepHost, headerRewriter)

	if err != nil {
		log.Fatalln(err)
	}

	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Forward to the host linkerd

		req.URL = fwdUrl
		fwd.ServeHTTP(w, req)
	})

	if err != nil {
		log.Fatalln(err)
	}

	// that's it! our reverse proxy is ready!
	// TLS config from https://cipherli.st/
	s := &http.Server{
		Addr:    ":" + port,
		Handler: redirect,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	}

	log.Fatal(s.ListenAndServeTLS(sslCertFile, sslKeyFile))
}

func GetOpt(name string, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}
