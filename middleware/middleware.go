package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	cf "github.com/jasonmichels/go-journey-server-utils/config"
	jh "github.com/jasonmichels/go-journey-server-utils/handler"
	"github.com/jasonmichels/journey-registry/journey"
	newrelic "github.com/newrelic/go-agent"
	"google.golang.org/grpc"
)

// NewRelicMiddleware New relic middleware that only adds new relic if you provided credentials
func NewRelicMiddleware(name string, key string, pattern string, next http.Handler) (string, http.Handler) {

	config := newrelic.NewConfig(name, key)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		log.Println("New relic issue and not being used: ", err)
	}

	if app != nil {
		log.Println("New relic is active")
		return newrelic.WrapHandle(app, "/", next)
	}

	return pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware Logging middleware to log each incoming request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)
		log.Printf("{Method: \"%v\", URL: \"%v\", RequestURI: \"%v\", Benchmark: \"%s\"}\n", r.Method, r.URL.Path, r.RequestURI, elapsed)
	})
}

// LocalAssetMiddleware Server file as local asset, if it is a local asset
func LocalAssetMiddleware(pathPrefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// first check to see if this is a request for an asset in the public folder
		if ok, path := isLocalAsset(r.URL, pathPrefix); ok {
			log.Printf("Serving local asset: %v", path)
			http.ServeFile(w, r, path)
			return
		}

		next.ServeHTTP(w, r)

	})
}

// isLocalAsset Check if is trying to serve local asset and if so, return the file path
func isLocalAsset(url *url.URL, pathPrefix string) (bool, string) {
	isLocal := false
	var file string

	if pathPrefix == "/" {
		pathPrefix = ""
	}

	switch filepath.Ext(url.Path) {
	case ".html", ".htm", "":
		isLocal = false
	default:
		isLocal = true
		path := strings.TrimPrefix(url.Path, pathPrefix)
		file = cf.PUBLIC + path
	}

	return isLocal, file
}

// JourneyAssetMiddleware Handle loading journey assets into handler
func JourneyAssetMiddleware(j *journey.Journey, registryURL string, next jh.JourneyHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// load assets and pass to the handler
		conn, err := grpc.Dial(registryURL, grpc.WithInsecure())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		c := journey.NewExplorerClient(conn)
		assets, err := c.GetDependencies(ctx, j)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r, assets)
	})
}
