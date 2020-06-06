package keeper

import (
	"net/http"
)

// Router like http.ServerMux, go-chi/chi.Router
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// HandleByRouter ...
func HandleByRouter(r Router) {
	r.HandleFunc("/_server/monitor/", HandleMonitor)
	r.HandleFunc("/_server/stacks/", HandleStack)
}

// ListenAndServe ...
func ListenAndServe(address string, handler http.Handler) error {
	http.HandleFunc("/_server/monitor/", HandleMonitor)
	http.HandleFunc("/_server/stacks/", HandleStack)
	return http.ListenAndServe(address, handler)
}
