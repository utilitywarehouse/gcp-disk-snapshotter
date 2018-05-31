package main

import (
	"flag"
	"os"
	"strings"

	"github.com/utilitywarehouse/gcp-disk-snapshotter/models"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/watch"
)

var (
	// flags
	flagProject         = flag.String("project", "", "(Required) GCP Project to use")
	flagZones           = flag.String("zones", "", "(Required) Comma separated list of zones where projects disks may live")
	flagLabels          = flag.String("labels", "", "(Required) Comma separated list of disk labels in format <name>:<value>")
	flagSnapPrefix      = flag.String("snap_prefix", "", "Prefix for created snapshots")
	flagRetentionHours  = flag.Int("retention_hours", 720, "Retention Duration in hours. Defaults to 720h = 1 month")
	flagIntervalSeconds = flag.Int("interval_secs", 43200, "Interval between snapshots in seconds. Defaults to 43200s = 12h")
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

func main() {

	// Flag Parsing
	flag.Parse()

	if *flagProject == "" {
		usage()
	}
	project := *flagProject

	if *flagZones == "" {
		usage()
	}
	zones := strings.Split(*flagZones, ",")

	if *flagLabels == "" {
		usage()
	}

	labels := &models.LabelList{}

	for _, label := range strings.Split(*flagLabels, ",") {

		l := strings.Split(label, ":")
		if len(l) < 2 {
			usage()
		}

		labels.AddLabel(l[0], l[1])
	}

	snapPrefix := *flagSnapPrefix
	retentionHours := *flagRetentionHours
	intervalSecs := *flagIntervalSeconds

	// Create a snapshotter
	gsc := snapshot.CreateGCPSnapClient(project, snapPrefix, zones, *labels)

	// Start watching
	watch.Watch(gsc, retentionHours, intervalSecs)

}
