package watch

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	compute "google.golang.org/api/compute/v1"
)

// From https://golang.org/src/time/format.go
const GCPSnapshotTimestampLayout string = "2006-01-02T15:04:05Z07:00"

func Watch(gsc *snapshot.GCPSnapClient, retentionHours, intervalSecs int) {

	for t := time.Tick(time.Second * 60); ; <-t {

		retentionStart := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
		lastAcceptedCreation := time.Now().Add(time.Duration(-intervalSecs) * time.Second)
		log.WithFields(log.Fields{
			"Retention Start":         retentionStart,
			"last accepted snap time": lastAcceptedCreation,
		}).Info("Initiating watch cycle")

		// Get disks
		disks, err := gsc.GetDiskList()
		if err != nil {
			log.Error(err)
			continue
		}

		for _, disk := range disks {
			log.Info("Checking disk: ", disk.Name)

			// Get snapshots per disk created by the snapshotter
			snaps, err := gsc.ListClientCreatedSnapshots(disk.SelfLink)
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
			if err := deleteSnapshots(gsc, snapsToDelete); err != nil {
				log.Error(err)
			}

			// Take snapshot if needed
			if snapNeeded {
				if err := createSnapshot(gsc, disk); err != nil {
					log.Error(err)
				}
			}

		}
	}
}

func deleteSnapshots(gsc *snapshot.GCPSnapClient, sl []compute.Snapshot) error {
	for _, s := range sl {
		log.Info("Attempting to delete snapshot: ", s.Name)
		op, err := gsc.DeleteSnapshot(s.Name)
		if err != nil {
			return err
		}

		// Delete snapshot is a global operation!!!
		go pollGlobalOperation(gsc, op)

	}
	return nil
}

func createSnapshot(gsc *snapshot.GCPSnapClient, d compute.Disk) error {
	log.Info("Taking snapshot of disk: ", d.Name)
	op, err := gsc.CreateSnapshot(d.Name, d.Zone)
	if err != nil {
		return err
	}

	// Create snapshot is a zonal operation!!!
	go pollZonalOperation(gsc, op, d.Zone)

	return nil
}

func pollZonalOperation(gsc *snapshot.GCPSnapClient, operation, zone string) {
	for {
		status, err := gsc.GetZonalOperationStatus(operation, zone)
		if err != nil {
			log.Info("Operation failed: ", operation)
			log.Error(err)
			break
		}
		if status == "DONE" {
			log.Info("Operation succeeded: ", operation)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func pollGlobalOperation(gsc *snapshot.GCPSnapClient, operation string) {
	for {
		status, err := gsc.GetGlobalOperationStatus(operation)
		if err != nil {
			log.Info("Operation failed: ", operation)
			log.Error(err)
			break
		}
		if status == "DONE" {
			log.Info("Operation succeeded: ", operation)
			break
		}
		time.Sleep(1 * time.Second)
	}
}
