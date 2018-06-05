package watch

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	compute "google.golang.org/api/compute/v1"
)

// From https://golang.org/src/time/format.go
const GCPSnapshotTimestampLayout string = "2006-01-02T15:04:05Z07:00"

type Watcher struct {
	GSC            snapshot.GCPSnapClientInterface
	WatchInterval  int
	RetentionHours int
	IntervalSecs   int
}

type WatcherInterface interface {
	Watch()
	deleteSnapshots(sl []compute.Snapshot)
	createSnapshot(d compute.Disk)
	pollZonalOperation(operation, zone string)
}

func (w *Watcher) Watch() {

	for t := time.Tick(time.Second * time.Duration(w.IntervalSecs)); ; <-t {
		retentionStart := time.Now().Add(-time.Duration(w.RetentionHours) * time.Hour)
		lastAcceptedCreation := time.Now().Add(time.Duration(-w.IntervalSecs) * time.Second)

		log.WithFields(log.Fields{
			"Watch Interval in secs":  w.WatchInterval,
			"Retention Period Start":  retentionStart,
			"last accepted snap time": lastAcceptedCreation,
		}).Info("Initiating new disk watch cycle")

		// Get disks
		disks, err := w.GSC.GetDiskList()
		if err != nil {
			log.Error(err)
			continue
		}

		for _, disk := range disks {
			log.Debug("Checking disk: ", disk.Name)

			// Get snapshots per disk created by the snapshotter
			snaps, err := w.GSC.ListClientCreatedSnapshots(disk.SelfLink)
			if err != nil {
				log.Fatal(err)
			}

			// Initialise values for delete list and snapshot flag
			snapsToDelete := []compute.Snapshot{}
			snapNeeded := true

			// Check timestamps of all snapshots to update the above vars
			for _, snap := range snaps {
				snapTime, err := time.Parse(GCPSnapshotTimestampLayout, snap.CreationTimestamp)
				if err != nil {
					log.Error("failed to parse timestamp:", err)
					continue
				}

				// If created before retention start time we need to delete
				if snapTime.Before(retentionStart) {
					snapsToDelete = append(snapsToDelete, snap)
				}

				// If a snap was taken after last accepted creation time we do not need a new one
				if snapTime.After(lastAcceptedCreation) {
					snapNeeded = false
				}
			}

			// Delete old snaps
			if err := w.deleteSnapshots(snapsToDelete); err != nil {
				log.Error(err)
			}

			// Take snapshot if needed
			if snapNeeded {
				if err := w.createSnapshot(disk); err != nil {
					log.Error(err)
				}
			}

		}

	}
}

func (w *Watcher) deleteSnapshots(sl []compute.Snapshot) error {
	for _, s := range sl {
		log.Info("Attempting to delete snapshot: ", s.Name)
		op, err := w.GSC.DeleteSnapshot(s.Name)
		if err != nil {
			return err
		}

		// Delete snapshot is a global operation!!!
		go w.pollGlobalOperation(op)

	}
	return nil
}

func (w *Watcher) createSnapshot(d compute.Disk) error {
	log.Debug("Attempt to take snapshot of disk: ", d.Name)
	op, err := w.GSC.CreateSnapshot(d.Name, d.Zone)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("New snapshot of disk: %v operation: %v", d.Name, op))

	// Create snapshot is a zonal operation!!!
	go w.pollZonalOperation(op, d.Zone)

	return nil
}

func (w *Watcher) pollZonalOperation(operation, zone string) {
	for {
		status, err := w.GSC.GetZonalOperationStatus(operation, zone)
		if err != nil {
			log.Error("Operation failed: ", operation, err)
			break
		}
		if status == "DONE" {
			log.Info("Operation succeeded: ", operation)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *Watcher) pollGlobalOperation(operation string) {
	for {
		status, err := w.GSC.GetGlobalOperationStatus(operation)
		if err != nil {
			log.Error("Operation failed: ", operation, err)
			break
		}
		if status == "DONE" {
			log.Info("Operation succeeded: ", operation)
			break
		}
		time.Sleep(1 * time.Second)
	}
}
