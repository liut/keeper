package keeper // import "lcgc/liut/keeper"

import (
	"log"
	"net/http"
	"runtime"
	"runtime/pprof"
)

func HandleStack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	debug := 1
	pf := r.FormValue("pf")
	switch pf {
	case "goroutine", "heap", "threadcreate", "block":
		err := pprof.Lookup(pf).WriteTo(w, debug)
		if err != nil {
			log.Print("profile.WriteTo error", err)
		}
	default:
		w.Write(stacks(r.FormValue("all") == "yes"))
	}
}

func stacks(all bool) []byte {
	// 堆栈可能很大，只输出最近的几次，这些数字应该够了，如果不够，那说明出了大问题，而且肯定不在这儿
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}
