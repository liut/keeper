package keeper

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/liut/keeper/utils/numbers"
)

// vars
var (
	BootstrapPrefix = "https://cdn.bootcdn.net/ajax/libs/twitter-bootstrap/4.5.3/"
)

var (
	startTime = time.Now()
)

// SysStatus ...
type SysStatus struct {
	Name         string `json:"hostname"`
	Uptime       string `json:"serverUptime"`
	NumGoroutine int    `json:"currentGoroutine"`

	// General statistics.
	MemAllocated string `json:"currentMemoryUsage"`   // bytes allocated and still in use
	MemTotal     string `json:"totalMemoryAllocated"` // bytes allocated (even if freed)
	MemSys       string `json:"memoryObtained"`       // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 `json:"pointerLookupTimes"`   // number of pointer lookups
	MemMallocs   uint64 `json:"memoryAllocateTimes"`  // number of mallocs
	MemFrees     uint64 `json:"memoryFreeTimes"`      // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string `json:"currentHeapUsage"`   // bytes allocated and still in use
	HeapSys      string `json:"heapMemoryObtained"` // bytes obtained from system
	HeapIdle     string `json:"heapMemoryIdle"`     // bytes in idle spans
	HeapInuse    string `json:"heapMemoryInUse"`    // bytes in non-idle span
	HeapReleased string `json:"heapMemoryReleased"` // bytes released to the OS
	HeapObjects  uint64 `json:"heapObjects"`        // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  string `json:"bootstrapStackUsage"` // bootstrap stacks
	StackSys    string `json:"stackMemoryObtained"`
	MSpanInuse  string `json:"mspanStructuresUsage"` // mspan structures
	MSpanSys    string `json:"mspanStructuresObtained"`
	MCacheInuse string `json:"mcacheStructuresUsage"` // mcache structures
	MCacheSys   string `json:"mcacheStructuresObtained"`
	BuckHashSys string `json:"profilingBucketHashTableObtained"` // profiling bucket hash table
	GCSys       string `json:"gcMetadataObtained"`               // GC metadata
	OtherSys    string `json:"otherSystemAllocationObtained"`    // other system allocations

	// Garbage collector statistics.
	NextGC       string `json:"nextGcRecycle"` // next run in HeapAlloc time (bytes)
	LastGC       string `json:"lastGcTime"`    // last run in absolute time (ns)
	PauseTotalNs string `json:"totalGcPause"`
	PauseNs      string `json:"lastGcPause"` // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32 `json:"gcTimes"`
}

// CurrentSystemStatus ...
func CurrentSystemStatus() *SysStatus {
	name, _ := os.Hostname()
	sysStatus := &SysStatus{Name: name}
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
	return sysStatus
}

// StatsToJSON ...
func StatsToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(CurrentSystemStatus())
}

// StatsToHTML ...
func StatsToHTML(w io.Writer) error {
	return tmpl.Execute(w, map[string]interface{}{"SysStatus": CurrentSystemStatus(), "BootstrapPrefix": BootstrapPrefix})
}

// HandleMonitor ...
func HandleMonitor(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Header.Get("Accept"), "application/json") ||
		r.FormValue("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := StatsToJSON(w)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if strings.HasPrefix(r.Header.Get("Accept"), "text/html") ||
		r.FormValue("format") == "html" {
		err := StatsToHTML(w)
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
<title>status of {{.SysStatus.Name}}</title>
<link rel="stylesheet" href="{{.BootstrapPrefix}}css/bootstrap.min.css">
<style>
.panel-header {padding-left: 4em;}
dt, dd {line-height: 1.6;}
.dl-horizontal dt {width: 246px;}
.dl-horizontal dd {margin-left: 270px;}
</style>
</head>
<body><div class="container">
<div class="panel panel-radius">
<div class="panel-header"><h3>System status of {{.SysStatus.Name}}</h3></div>
<div class="panel-body">
<dl class="dl-horizontal">
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
