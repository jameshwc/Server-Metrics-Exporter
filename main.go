package main

import (
	"log"
	"net/http"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CPU struct {
	Total, User, System, Idle uint64
}

type Mem struct {
	Used, Available, Total uint64
}

type Disk struct {
	Used, Total uint64
}

func (m *Mem) Usage() float64 {
	return 1 - float64(m.Available)/float64(m.Total)
}

func (d *Disk) Usage() float64 {
	return float64(d.Used) / float64(d.Total)
}

type Server struct {
	CPU
	Mem
	Disk
}

func CalCpuUsage() CPU {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Fatal("stat read fail")
	}
	var st CPU
	for _, s := range stat.CPUStats {
		st.Total += s.IOWait + s.IRQ + s.Idle + s.Nice + s.SoftIRQ + s.Steal + s.System + s.User
		st.User += s.User
		st.System += s.System
		st.Idle += s.Idle
	}
	return st
}

func CalMemUsage() Mem {
	stat, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.Fatal("stat read fail")
	}
	return Mem{Used: stat.MemTotal - stat.MemAvailable, Available: stat.MemAvailable, Total: stat.MemTotal}
}

func CalDiskUsage() Disk {
	stat, err := linuxproc.ReadDisk("/")
	if err != nil {
		log.Fatal("stat read fail")
	}
	return Disk{Used: stat.Used, Total: stat.All}
}

func NewServer() *Server {
	return &Server{CalCpuUsage(), CalMemUsage(), CalDiskUsage()}
}

var (
	namespace = ""
	uptime    = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "uptime",
			Help:      "HTTP service uptime.",
		}, nil,
	)

	userCPU = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_cpu_usage",
			Help:      "User CPU Usage.",
		}, nil,
	)

	systemCPU = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "system_cpu_usage",
			Help:      "System CPU Usage.",
		}, nil,
	)

	memUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "system_mem_usage",
			Help:      "System Mem Usage.",
		}, nil,
	)

	diskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "system_disk_usage",
			Help:      "System Disk Usage.",
		}, nil,
	)
)

func recordServerMetrics() {

	prev := NewServer()

	for range time.Tick(time.Second) {
		cur := NewServer()

		cpuTotal := float64(cur.CPU.Total - prev.CPU.Total)
		uptime.WithLabelValues().Inc()
		userCPU.WithLabelValues().Set(float64(cur.CPU.User-prev.CPU.User) / cpuTotal * 100)
		systemCPU.WithLabelValues().Set(float64(cur.CPU.System-prev.CPU.System) / cpuTotal * 100)
		memUsage.WithLabelValues().Set(cur.Mem.Usage() * 100)
		diskUsage.WithLabelValues().Set(cur.Disk.Usage() * 100)

		prev = cur
	}
}

func init() {
	prometheus.MustRegister(uptime, userCPU, systemCPU, memUsage, diskUsage)
	go recordServerMetrics()
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
