package handler

import (
	"net/http"

	"github.com/jasonmichels/journey-registry/journey"
)

// JourneyHandler A journey handler to handle journey assets
type JourneyHandler func(http.ResponseWriter, *http.Request, *journey.DependencyAssets)

func (fn JourneyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, assets *journey.DependencyAssets) {
	fn(w, r, assets)
}
