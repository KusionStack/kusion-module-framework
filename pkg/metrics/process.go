package metrics

/*
@ Package Description:
This package current collect process cpu and memory usage rate, memory rss usage, requests already processed by the process and can easily to expand other stats.
At the same time, by integrating Prometheus, process related metrics were exposed to the outside.

@ How to use this package ？
You can follow below step to record process metrics:
1. initialize process metrics server. for example:
    @Parameter: listenAddress -- (string) server listen address.
	errChan := metrics.InitProcessMetricsServer(":9090")
	select {
	case err := <-errChan:
		log.Fatalf("Failed to init process metrics server: %v", err)
	case <-os.Interrupt:
		log.Println("Received interrupt signal, shutting down.")
	}

2.You can invoke RecordProcessResourceUsageMetrics to record process resource usage metrics, for example:
2.1 Record current process metrics.
   @Parameter: pid -- (int32) process pid.
   err := metrics.RecordProcessResourceUsageMetrics(43123)
   if err != nil {
      log.Warn(err)
   }
2.2 Record multiple process metrics.
   for _, pid := range pids {
		err := metrics.RecordProcessResourceUsageMetrics(pid)
		if err != nil {
           log.Warn(err)
        }
   }

3. You can invoke RecordProcessRequestMetrics to record requests metrics processed by specific process, for example:
    @Parameter: pid -- (int32) process pid.
    @Parameter: traceID -- (string) request trace id.
    @Parameter: duration -- (float64) duration second for request is processed.
    start := time.Now()
    process(requests)
    duration := time.Since(start).Seconds()
	err := metrics.RecordProcessRequestMetrics(43123, "1111-2222-3333", duration)
    if err != nil {
       log.Warn(err)
    }

@ How kusion integrate with this package to record module process metrics ?
Kusion should manage the module subprocess information and regularly collect subprocess resource usage metrics.
At the same time, Kusion should record the request metrics processed by specific module process.

@ How to fetch the process metrics ?
You can test it locally by using curl command. for example:
curl -v http://127.0.0.1:9090/process/metrics

# HELP process_cpu_percent How many percent of the CPU time process uses.
# TYPE process_cpu_percent gauge
process_cpu_percent{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585"} 0.5315134891429808
# HELP process_memory_percent How many percent of the memory process uses.
# TYPE process_memory_percent gauge
process_memory_percent{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585"} 0.037479400634765625
# HELP process_memory_rss How many rss of the memory process uses.
# TYPE process_memory_rss gauge
process_memory_rss{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585"} 6.438912e+06
# HELP process_per_request_duration_seconds Process per request duration in seconds.
# TYPE process_per_request_duration_seconds gauge
process_per_request_duration_seconds{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",trace_id="11111"} 5.000896791
# HELP process_request_duration_seconds Process request duration in seconds.
# TYPE process_request_duration_seconds histogram
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.01"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.025"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.05"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.1"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.25"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="0.5"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="1"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="2.5"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="5"} 0
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="10"} 1
process_request_duration_seconds_bucket{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585",le="+Inf"} 1
process_request_duration_seconds_sum{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585"} 5.000896791
process_request_duration_seconds_count{hostname="xxxMacBook-Pro.local",pid="43638",pid_name="metrics",ppid="43585"} 1
*/

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	ProcessCPUPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_cpu_percent",
			Help: "How many percent of the cpu time process uses.",
		},
		[]string{"hostname", "pid", "pid_name", "ppid"},
	)

	ProcessMemoryPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_percent",
			Help: "How many percent of the memory process uses.",
		},
		[]string{"hostname", "pid", "pid_name", "ppid"},
	)

	ProcessMemoryRSS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_rss",
			Help: "How many rss of the memory process uses.",
		},
		[]string{"hostname", "pid", "pid_name", "ppid"},
	)

	ProcessRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "process_request_duration_seconds",
			Help:    "Process request duration in seconds.",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"hostname", "pid", "pid_name", "ppid"},
	)

	ProcessPerRequestDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_per_request_duration_seconds",
			Help: "Process per request duration in seconds.",
		},
		[]string{"hostname", "pid", "pid_name", "ppid", "trace_id"},
	)
)

func InitProcessMetricsServer(listenAddress string) <-chan error {
	if listenAddress == "" {
		listenAddress = os.Getenv("PROCESS_METRICS_LISTEN_ADDRESS")
		if listenAddress == "" {
			listenAddress = ":9090"
		}
	}
	customRegistry := prometheus.NewRegistry()
	customRegistry.MustRegister(ProcessCPUPercent)
	customRegistry.MustRegister(ProcessMemoryPercent)
	customRegistry.MustRegister(ProcessMemoryRSS)
	customRegistry.MustRegister(ProcessRequestDuration)
	customRegistry.MustRegister(ProcessPerRequestDuration)

	var errCH chan error
	go func() {
		http.Handle("/process/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))
		errCH <- http.ListenAndServe(listenAddress, nil)
	}()
	return errCH
}

func RecordProcessResourceUsageMetrics(pid int32) error {
	pi, err := CollectProcessInfo(pid)
	if err != nil {
		return err
	}
	hostName, _ := os.Hostname()
	labels := prometheus.Labels{
		"hostname": hostName,
		"pid":      fmt.Sprintf("%v", pi.Pid),
		"pid_name": pi.ProcessName,
		"ppid":     fmt.Sprintf("%v", pi.ParentPid),
	}
	ProcessCPUPercent.With(labels).Set(pi.CPUInfo.CPUPercent)
	ProcessMemoryPercent.With(labels).Set(float64(pi.MemoryInfo.MemoryPercent))
	ProcessMemoryRSS.With(labels).Set(float64(pi.MemoryInfo.RSS))
	return nil
}

func RecordProcessRequestMetrics(pid int32, traceID string, duration float64) error {
	pi, err := CollectProcessInfo(pid)
	if err != nil {
		return err
	}
	hostName, _ := os.Hostname()
	labels := prometheus.Labels{
		"hostname": hostName,
		"pid":      fmt.Sprintf("%v", pi.Pid),
		"pid_name": pi.ProcessName,
		"ppid":     fmt.Sprintf("%v", pi.ParentPid),
	}
	ProcessRequestDuration.With(labels).Observe(duration)
	labels["trace_id"] = traceID
	ProcessPerRequestDuration.With(labels).Set(duration)
	return nil
}

type ProcessInfo struct {
	Pid         int32
	ParentPid   int32
	ProcessName string
	CPUInfo     *CPUInfo
	MemoryInfo  *MemoryInfo
}

type CPUInfo struct {
	CPUPercent float64
}

type MemoryInfo struct {
	MemoryPercent float32
	*process.MemoryInfoStat
}

func CollectProcessInfo(pid int32) (*ProcessInfo, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to create process[%v] instance: %v\n", pid, err)
	}
	name, err := p.Name()
	if err != nil {
		return nil, fmt.Errorf("failed to get process[%v] name: %v", pid, err)
	}
	parentPid, err := p.Ppid()
	if err != nil {
		return nil, fmt.Errorf("failed to get process[%v] parent process pid: %v", pid, err)
	}
	cpuPercent, err := p.CPUPercent()
	if err != nil {
		return nil, fmt.Errorf("failed to get process[%v] cpu percent: %v", pid, err)
	}
	memoryPercent, err := p.MemoryPercent()
	if err != nil {
		fmt.Errorf("failed to get process[%v] memory percent: %v", pid, err)
	}
	memInfo, err := p.MemoryInfo()
	if err != nil {
		fmt.Errorf("failed to get process[%v] memory info: %v", pid, err)
	}

	pi := &ProcessInfo{
		Pid:         pid,
		ParentPid:   parentPid,
		ProcessName: name,
		CPUInfo:     &CPUInfo{CPUPercent: cpuPercent},
		MemoryInfo: &MemoryInfo{
			MemoryPercent:  memoryPercent,
			MemoryInfoStat: memInfo,
		},
	}
	return pi, nil
}
