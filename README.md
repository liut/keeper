keeper
======

>  Web monitor for golang application

````go
package main

import (
	"github.com/liut/keeper"
	"net/http"
)

func main() {
	http.HandleFunc("/_server/monitor/", keeper.HandleMonitor)
	http.HandleFunc("/_server/stacks/", keeper.HandleStack)

	// other handles

	http.ListenAndServe(":8080", nil)
}

````
