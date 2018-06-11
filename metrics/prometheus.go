package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/go-operational/op"
)

type Prometheus struct {
	createSnapshotSuccess *prometheus.CounterVec
	deleteSnapshotSuccess *prometheus.CounterVec
	operationSuccess      *prometheus.CounterVec
}

// PrometheusInterface allows for mocking out the functionality of Prometheus when testing the full process of an apply run.
type PrometheusInterface interface {
	Init()
	UpdateCreateSnapshotStatus(disk string, success bool)
	UpdateDeleteSnapshotStatus(disk string, success bool)
	UpdateOperationStatus(operation_type string, success bool)
}

func (p *Prometheus) Init() {
	p.createSnapshotSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gcp_disk_snapshotter_create_api_call_count",
		Help: "Success metric for snapshots created per disk",
	},
		[]string{
			// Path of the file that was applied
			"disk",
			// Result: true if creation api call was successful, false otherwise
			"success",
		},
	)
	p.deleteSnapshotSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gcp_disk_snapshotter_delete_api_call_count",
		Help: "Success metric for snapshots deletion per disk",
	},
		[]string{
			// Path of the file that was applied
			"disk",
			// Result: true if deletion api call was successful, false otherwise
			"success",
		},
	)
	p.operationSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gcp_disk_snapshotter_operation_count",
		Help: "Success metric for operations initiated by disk snapshotter",
	},
		[]string{
			// Global or Zonal
			"operation_type",
			// Result: true if the operation was successful, false otherwise
			"success",
		},
	)
	prometheus.MustRegister(p.createSnapshotSuccess)
	prometheus.MustRegister(p.deleteSnapshotSuccess)
	prometheus.MustRegister(p.operationSuccess)

	go startServer()
}

func startServer() {
	log.Info("starting HTTP endpoints ...")

	mux := http.NewServeMux()
	mux.Handle("/__/", op.NewHandler(
		op.NewStatus("gcp-disk-snapshotter", "gcp-disk-snapshotter handles snapshot creation/deletion on gcp for a given set of disks").
			AddOwner("infrastructure", "#infra").
			AddLink("github", "https://github.com/utilitywarehouse/gcp-disk-snapshotter").
			SetRevision("master").
			AddChecker("running", func(cr *op.CheckResponse) { cr.Healthy("service is running") }).
			ReadyAlways(),
	))
	mux.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(":5000", mux); err != nil {
		log.Fatal("could not start HTTP router: ", err)
	}
}

// UpdateCreateSnapshotStatus increments the given disk's Counter for either successful create attempts or failed apply attempts.
func (p *Prometheus) UpdateCreateSnapshotStatus(disk string, success bool) {
	p.createSnapshotSuccess.With(prometheus.Labels{
		"disk": disk, "success": strconv.FormatBool(success),
	}).Inc()
}

// UpdateDeleteSnapshotStatus increments the given disk's Counter for either successful delete attempts or failed apply attempts.
func (p *Prometheus) UpdateDeleteSnapshotStatus(disk string, success bool) {
	p.deleteSnapshotSuccess.With(prometheus.Labels{
		"disk": disk, "success": strconv.FormatBool(success),
	}).Inc()
}

// UpdateDeleteSnapshotStatus increments the given disk's Counter for either successful delete attempts or failed apply attempts.
func (p *Prometheus) UpdateOperationStatus(operation_type string, success bool) {
	p.operationSuccess.With(prometheus.Labels{
		"operation_type": operation_type, "success": strconv.FormatBool(success),
	}).Inc()
}
