package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/dghubble/sessions"
	"github.com/koding/websocketproxy"
)

var (
	bind = flag.String("http", ":80", "listen address")
	host = flag.String("host", "beyond.colofoo.net", "hostname of self, eg. when generating OAuth redirect URLs")

	healthPath  = flag.String("health-path", "/healthz/ping", "URL of the health endpoint")
	healthReply = flag.String("health-reply", "ok", "response body of the health endpoint")

	cookieAge  = flag.Int("cookie-age", 3600*6, "MaxAge setting in seconds")
	cookieDom  = flag.String("cookie-domain", ".colofoo.net", "session cookie domain")
	cookieKey1 = flag.String("cookie-key1", "t8yG1gmeEyeb7pQpw544UeCTyDfPkE6u", "key1 of cookie crypto pair")
	cookieKey2 = flag.String("cookie-key2", "Q599vrruZRhLFC144thCRZpyHM7qGDjt", "key2 of cookie crypto pair")
	cookieName = flag.String("cookie-name", "transcend", "session cookie name")

	fouroFourMessage = flag.String("404-message", "Please contact your network administrators to whitelist this system.", "message to use for unlisted hosts when learning is disabled or fails")
	fouroOneCode     = flag.Int("401-code", 418, "status to respond when a user needs authentication")

	skipVerify = flag.Bool("insecure-skip-verify", false, "allow TLS backends without valid certificates")
	wsCompress = flag.Bool("websocket-compression", false, "allow websocket transport compression (gorilla/experimental)")

	store *sessions.CookieStore
)

func init() {
	store = sessions.NewCookieStore([]byte(*cookieKey1), []byte(*cookieKey2))
	store.Config.Domain = *cookieDom
	store.Config.MaxAge = *cookieAge

	// allow insecure backends
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *skipVerify},
	}
	if websocketproxy.DefaultDialer.TLSClientConfig == nil {
		websocketproxy.DefaultDialer.TLSClientConfig = &tls.Config{}
	}
	websocketproxy.DefaultDialer.TLSClientConfig.InsecureSkipVerify = *skipVerify
	websocketproxy.DefaultDialer.EnableCompression = *wsCompress
	websocketproxy.DefaultUpgrader.EnableCompression = *wsCompress
	websocketproxy.DefaultUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(*bind, http.HandlerFunc(handler)))
}

func setup() error {
	err := oidcSetup()
	if err == nil {
		err = refreshFence()
	}
	if err == nil {
		err = refreshSites()
	}
	if err == nil {
		err = refreshWhitelist()
	}
	if err == nil {
		err = reproxy()
	}
	return err
}
