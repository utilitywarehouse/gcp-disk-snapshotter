package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/models"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/watch"
)

var (
	// flags
	flagProject       = flag.String("project", "", "(Required) GCP Project to use")
	flagZones         = flag.String("zones", "", "(Required) Comma separated list of zones where projects disks may live")
	flagConfFile      = flag.String("conf_file", "", "(Required) Path of the configuration file tha contains the targets based on label or description")
	flagSnapPrefix    = flag.String("snap_prefix", "", "Prefix for created snapshots")
	flagWatchInterval = flag.Int("watch_interval", 60, "Interval between watch cycles in seconds. Defaults to 60s")
	flagLogLevel      = flag.String("log_level", "info", "Log Level, defaults to INFO")
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

func initLogging(logLevel string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal("error parsing log level: %v", err)
	}
	log.SetLevel(level)
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

	if *flagConfFile == "" {
		usage()
	}

	snapPrefix := *flagSnapPrefix
	watchInterval := *flagWatchInterval
	logLevel := *flagLogLevel

	// Init logging
	initLogging(logLevel)

	// Load config
	snapshotConfigs := loadSnapshotConfig(*flagConfFile)

	log.Debug("Reading Configuration")
	for _, d := range snapshotConfigs.Descriptions {
		log.Debug("description: ", d.Description.Key, " ", d.Description.Value)
	}
	for _, l := range snapshotConfigs.Labels {
		log.Debug("label: ", l.Label.Key, " ", l.Label.Value)
	}
	// Create a snapshotter
	gsc := snapshot.CreateGCPSnapClient(project, snapPrefix, zones)

	// Start watching
	watcher := &watch.Watcher{
		GSC:           gsc,
		WatchInterval: watchInterval,
	}
	watcher.Watch(snapshotConfigs)

}

func loadSnapshotConfig(snapshotConfigFile string) *models.SnapshotConfigs {
	confFile, err := os.Open(snapshotConfigFile)
	if err != nil {
		log.Fatal("Error while opening volume snapshot config file: ", err)
	}
	fileContent, err := ioutil.ReadAll(confFile)
	if err != nil {
		log.Fatal("Error while reading volume snapshot config file: ", err)
	}
	snapshotConfigs := &models.SnapshotConfigs{}
	if err = json.Unmarshal(fileContent, snapshotConfigs); err != nil {
		log.Fatal("Error unmarshalling snapshots config file: ", err)
	}
	return snapshotConfigs
}
