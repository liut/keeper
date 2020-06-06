package keeper

import (
	"net/http"
)

var (
	_ Muxer = (*http.ServeMux)(nil)
)

// Muxer like http.ServerMux, go-chi/chi.Router
type Muxer interface {
	Handle(pattern string, handler http.Handler)
}

// StrapMux ...
func StrapMux(mux Muxer) {
	mux.Handle("/_server/monitor/", http.HandlerFunc(HandleMonitor))
	mux.Handle("/_server/stacks/", http.HandlerFunc(HandleStack))
}

// ListenAndServe 没什么用的接口
func ListenAndServe(address string, handler http.Handler) error {
	mux := http.NewServeMux()
	StrapMux(mux)
	mux.Handle("/", handler)
	return http.ListenAndServe(address, mux)
}
