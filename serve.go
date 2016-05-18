package keeper // import "lcgc/liut/keeper"

import (
	"net/http"
)

func ListenAndServe(address string, handler http.Handler) error {
	http.HandleFunc("/_server/monitor/", HandleMonitor)
	http.HandleFunc("/_server/stacks/", HandleStack)
	return http.ListenAndServe(address, handler)
}
