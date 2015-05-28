package keeper

import (
	"net/http"
)

// deprecated
func ServeMonitor(address string) error {
	return ListenAndServe(address)
}

func ListenAndServe(address string) error {
	http.HandleFunc("/_server/monitor/", HandleMonitor)
	http.HandleFunc("/_server/stacks/", HandleStack)
	return http.ListenAndServe(address, nil)
}
