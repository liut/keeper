package keeper

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
	"tuluu.com/liut/keeper/utils/numbers"
)

var (
	startTime       = time.Now()
	BootstrapPrefix = "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/"
)

var sysStatus struct {
	Uptime       string `json:"server_uptime"`
	NumGoroutine int    `json:"current_goroutine"`

	// General statistics.
	MemAllocated string `json:"current_memory_usage"`   // bytes allocated and still in use
	MemTotal     string `json:"total_memory_allocated"` // bytes allocated (even if freed)
	MemSys       string `json:"memory_obtained"`        // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 `json:"pointer_lookup_times"`   // number of pointer lookups
	MemMallocs   uint64 `json:"memory_allocate_times"`  // number of mallocs
	MemFrees     uint64 `json:"memory_free_times"`      // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string `json:"current_heap_usage"`   // bytes allocated and still in use
	HeapSys      string `json:"heap_memory_obtained"` // bytes obtained from system
	HeapIdle     string `json:"heap_memory_idle"`     // bytes in idle spans
	HeapInuse    string `json:"heap_memory_in_use"`   // bytes in non-idle span
	HeapReleased string `json:"heap_memory_released"` // bytes released to the OS
	HeapObjects  uint64 `json:"heap_objects"`         // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  string `json:"bootstrap_stack_usage"` // bootstrap stacks
	StackSys    string `json:"stack_memory_obtained"`
	MSpanInuse  string `json:"mspan_structures_usage"` // mspan structures
	MSpanSys    string `json:"mspan_structures_obtained"`
	MCacheInuse string `json:"mcache_structures_usage"` // mcache structures
	MCacheSys   string `json:"mcache_structures_obtained"`
	BuckHashSys string `json:"profiling_bucket_hash_table_obtained"` // profiling bucket hash table
	GCSys       string `json:"gc_metadata_obtained"`                 // GC metadata
	OtherSys    string `json:"other_system_allocation_obtained"`     // other system allocations

	// Garbage collector statistics.
	NextGC       string `json:"next_gc_recycle"` // next run in HeapAlloc time (bytes)
	LastGC       string `json:"last_gc_time"`    // last run in absolute time (ns)
	PauseTotalNs string `json:"total_gc_pause"`
	PauseNs      string `json:"last_gc_pause"` // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32 `json:"gc_times"`
}

func updateSystemStatus() {
	sysStatus.Uptime = numbers.TimeSincePro(startTime)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	sysStatus.NumGoroutine = runtime.NumGoroutine()

	sysStatus.MemAllocated = numbers.PrettySize(int64(m.Alloc))
	sysStatus.MemTotal = numbers.PrettySize(int64(m.TotalAlloc))
	sysStatus.MemSys = numbers.PrettySize(int64(m.Sys))
	sysStatus.Lookups = m.Lookups
	sysStatus.MemMallocs = m.Mallocs
	sysStatus.MemFrees = m.Frees

	sysStatus.HeapAlloc = numbers.PrettySize(int64(m.HeapAlloc))
	sysStatus.HeapSys = numbers.PrettySize(int64(m.HeapSys))
	sysStatus.HeapIdle = numbers.PrettySize(int64(m.HeapIdle))
	sysStatus.HeapInuse = numbers.PrettySize(int64(m.HeapInuse))
	sysStatus.HeapReleased = numbers.PrettySize(int64(m.HeapReleased))
	sysStatus.HeapObjects = m.HeapObjects

	sysStatus.StackInuse = numbers.PrettySize(int64(m.StackInuse))
	sysStatus.StackSys = numbers.PrettySize(int64(m.StackSys))
	sysStatus.MSpanInuse = numbers.PrettySize(int64(m.MSpanInuse))
	sysStatus.MSpanSys = numbers.PrettySize(int64(m.MSpanSys))
	sysStatus.MCacheInuse = numbers.PrettySize(int64(m.MCacheInuse))
	sysStatus.MCacheSys = numbers.PrettySize(int64(m.MCacheSys))
	sysStatus.BuckHashSys = numbers.PrettySize(int64(m.BuckHashSys))
	sysStatus.GCSys = numbers.PrettySize(int64(m.GCSys))
	sysStatus.OtherSys = numbers.PrettySize(int64(m.OtherSys))

	sysStatus.NextGC = numbers.PrettySize(int64(m.NextGC))
	sysStatus.LastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	sysStatus.PauseTotalNs = fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000)
	sysStatus.PauseNs = fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	sysStatus.NumGC = m.NumGC
}

func MonitorJsonTo(w io.Writer) error {
	updateSystemStatus()
	return json.NewEncoder(w).Encode(sysStatus)
}

func MonitorHtmlTo(w io.Writer) (err error) {
	updateSystemStatus()
	return tmpl.Execute(w, map[string]interface{}{"SysStatus": sysStatus, "BootstrapPrefix": BootstrapPrefix})
}

func HandleMonitor(w http.ResponseWriter, r *http.Request) {
	updateSystemStatus()
	if strings.HasSuffix(r.URL.Path, ".json") || r.FormValue("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := MonitorJsonTo(w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Print(err)
		}
		return
	}

	if strings.HasSuffix(r.URL.Path, ".html") || r.FormValue("format") == "html" {
		err := MonitorHtmlTo(w)
		if err != nil {
			log.Print(err)
		}
		return
	}

	http.NotFound(w, r)
}

var tmpl = template.Must(template.New("index").Parse(`<html>
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>monitor</title>
<link rel="stylesheet" href="{{.BootstrapPrefix}}css/bootstrap.min.css">
<style>
.dl-horizontal dt {width: 240px;}
.dl-horizontal dd {margin-left: 260px;}
</style>
</head>
<body><div class="container">
<div class="panel panel-radius">
<div class="panel-header"><h3>system monitor</h3></div>
<div class="panel-body">
<dl class="dl-horizontal admin-dl-horizontal">
   <dt>server_uptime</dt> <dd>{{.SysStatus.Uptime}}</dd>
   <dt>current_goroutine</dt> <dd>{{.SysStatus.NumGoroutine}}</dd>

   <hr/>
   <dt>current_memory_usage</dt> <dd>{{.SysStatus.MemAllocated}}</dd>
   <dt>total_memory_allocated</dt> <dd>{{.SysStatus.MemTotal}}</dd>
   <dt>memory_obtained</dt> <dd>{{.SysStatus.MemSys}}</dd>
   <dt>pointer_lookup_times</dt> <dd>{{.SysStatus.Lookups}}</dd>
   <dt>memory_allocate_times</dt> <dd>{{.SysStatus.MemMallocs}}</dd>
   <dt>memory_free_times</dt> <dd>{{.SysStatus.MemFrees}}</dd>

   <hr/>
   <dt>current_heap_usage</dt> <dd>{{.SysStatus.HeapAlloc}}</dd>
   <dt>heap_memory_obtained</dt> <dd>{{.SysStatus.HeapSys}}</dd>
   <dt>heap_memory_idle</dt> <dd>{{.SysStatus.HeapIdle}}</dd>
   <dt>heap_memory_in_use</dt> <dd>{{.SysStatus.HeapInuse}}</dd>
   <dt>heap_memory_released</dt> <dd>{{.SysStatus.HeapReleased}}</dd>
   <dt>heap_objects</dt> <dd>{{.SysStatus.HeapObjects}}</dd>

   <hr/>
   <dt>bootstrap_stack_usage</dt> <dd>{{.SysStatus.StackInuse}}</dd>
   <dt>stack_memory_obtained</dt> <dd>{{.SysStatus.StackSys}}</dd>
   <dt>mspan_structures_usage</dt> <dd>{{.SysStatus.MSpanInuse}}</dd>
   <dt>mspan_structures_obtained</dt> <dd>{{.SysStatus.HeapSys}}</dd>
   <dt>mcache_structures_usage</dt> <dd>{{.SysStatus.MCacheInuse}}</dd>
   <dt>mcache_structures_obtained</dt> <dd>{{.SysStatus.MCacheSys}}</dd>
   <dt>profiling_bucket_hash_table_obtained</dt> <dd>{{.SysStatus.BuckHashSys}}</dd>
   <dt>gc_metadata_obtained</dt> <dd>{{.SysStatus.GCSys}}</dd>
   <dt>other_system_allocation_obtained</dt> <dd>{{.SysStatus.OtherSys}}</dd>

   <hr>
   <dt>next_gc_recycle</dt> <dd>{{.SysStatus.NextGC}}</dd>
   <dt>last_gc_time</dt> <dd>{{.SysStatus.LastGC}}</dd>
   <dt>total_gc_pause</dt> <dd>{{.SysStatus.PauseTotalNs}}</dd>
   <dt>last_gc_pause</dt> <dd>{{.SysStatus.PauseNs}}</dd>
   <dt>gc_times</dt> <dd>{{.SysStatus.NumGC}}</dd>
</dl>
</div>
</div>
</div></body>
</html>
`))
