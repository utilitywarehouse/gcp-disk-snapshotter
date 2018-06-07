package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"

	"github.com/utilitywarehouse/gcp-disk-snapshotter/models"
)

const (
	SnapshotterLabel      string = "gcp_disk_snapshotter"
	SnapshotterLabelValue string = "true"
)

var googleClient *http.Client

type GCPSnapClient struct {
	Project        string
	Zones          []string
	SnapPrefix     string
	ComputeService compute.Service
}

type GCPSnapClientInterface interface {
	GetDisksFromLabel(label *models.Label) ([]compute.Disk, error)
	GetDisksFromDescription(label *models.Description) ([]compute.Disk, error)
	ListSnapshots(diskSelfLink string) ([]compute.Snapshot, error)
	ListClientCreatedSnapshots(diskSelfLink string) ([]compute.Snapshot, error)
	CreateSnapshot(diskName, zone string) (string, error)
	DeleteSnapshot(snapName string) (string, error)
	GetZonalOperationStatus(operation, zone string) (string, error)
	GetGlobalOperationStatus(operation string) (string, error)
}

// Basic Init function for the snapshotter
func CreateGCPSnapClient(project, snapPrefix string, zones []string) *GCPSnapClient {

	ctx := context.Background()
	googleClient, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log.Fatal("Failed to create google client:", err)
	}

	computeService, err := compute.New(googleClient)
	if err != nil {
		log.Fatal("Failed to create compute service:", err)
	}

	return &GCPSnapClient{
		Project:        project,
		Zones:          zones,
		SnapPrefix:     snapPrefix,
		ComputeService: *computeService,
	}
}

// In case of a gcp link it returns the target (final part after /)
func formatLinkString(in string) string {

	if strings.ContainsAny(in, "/") {
		elems := strings.Split(in, "/")
		return elems[len(elems)-1]
	}
	return in
}

// GetDiskList: Returns a list of disks that contain one of the given labels
func (gsc *GCPSnapClient) GetDisksFromLabel(label *models.Label) ([]compute.Disk, error) {

	disks := []compute.Disk{}

	for _, zone := range gsc.Zones {
		resp, err := gsc.ComputeService.Disks.List(gsc.Project, zone).Do()
		if err != nil {
			return disks, errors.Wrap(err, "error listing disks")
		}

		for _, disk := range resp.Items {
			if val, ok := disk.Labels[label.Key]; ok {
				if label.Value == val {
					disks = append(disks, *disk)
				}
			}
		}
	}

	return disks, nil
}

func (gsc *GCPSnapClient) GetDisksFromDescription(desc *models.Description) ([]compute.Disk, error) {

	disks := []compute.Disk{}

	for _, zone := range gsc.Zones {
		resp, err := gsc.ComputeService.Disks.List(gsc.Project, zone).Do()
		if err != nil {
			return disks, errors.Wrap(err, "error listing disks")
		}

		for _, disk := range resp.Items {
			var dObj map[string]string
			if err = json.Unmarshal([]byte(disk.Description), &dObj); err != nil {
				log.Debug("Skipping: error unmarshalling disk description to map: ", err)
				continue
			}
			if val, ok := dObj[desc.Key]; ok {
				if desc.Value == val {
					disks = append(disks, *disk)
				}
			}
		}
	}

	return disks, nil
}

// ListSnapshots: Lists Snapshots for a given disk
func (gsc *GCPSnapClient) ListSnapshots(diskSelfLink string) ([]compute.Snapshot, error) {

	snapshots := []compute.Snapshot{}

	resp, err := gsc.ComputeService.Snapshots.List(gsc.Project).Do()
	if err != nil {
		return snapshots, errors.Wrap(err, "error requesting snapshots list:")
	}
	for _, snap := range resp.Items {
		// If not created by the snapshotter just ignore
		if val, ok := snap.Labels[SnapshotterLabel]; ok {
			if val != SnapshotterLabelValue {
				continue
			}
		} else {
			continue
		}

		// If it's a snapshot of the input disk add to the list
		if snap.SourceDisk == diskSelfLink {
			snapshots = append(snapshots, *snap)
		}
	}
	return snapshots, nil
}

// ListClientCreatedSnapshots: Lists snapshots for a given disk that were create by the client,
// meaning that they have the SnapshotterLabel
func (gsc *GCPSnapClient) ListClientCreatedSnapshots(diskSelfLink string) ([]compute.Snapshot, error) {

	snaps, err := gsc.ListSnapshots(diskSelfLink)
	if err != nil {
		return []compute.Snapshot{}, err
	}

	res := []compute.Snapshot{}
	for _, snap := range snaps {
		if val, ok := snap.Labels[SnapshotterLabel]; ok {
			if SnapshotterLabelValue == val {
				res = append(res, snap)
				continue
			}
		}
	}

	return res, nil
}

// CreateSnapshot: Gets a disk name and a zone, issues a create snapshot command to api
// and returns a link to the create snapshot operation
func (gsc *GCPSnapClient) CreateSnapshot(diskName, zone string) (string, error) {

	// format zone if link
	zn := formatLinkString(zone)

	// lowercase letters, numeric characters, underscores and dashes, at most 63 characters long
	snapLabels := map[string]string{
		SnapshotterLabel: SnapshotterLabelValue,
	}

	// Name must match regex '(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)'
	snapshot := &compute.Snapshot{
		Description: fmt.Sprintf("Snapshot of %s", diskName),
		Name:        fmt.Sprintf("%s%s-%s", gsc.SnapPrefix, diskName, time.Now().Format("20060102150405")),
		Labels:      snapLabels,
	}

	resp, err := gsc.ComputeService.Disks.CreateSnapshot(gsc.Project, zn, diskName, snapshot).Do()
	if err != nil {
		return "", errors.Wrap(err, "error taking disk snapshot:")
	}

	return resp.SelfLink, nil
}

// DeleteSnapshot: Gets a snapshot name and issues a delete. Returns a link to the delete operation
func (gsc *GCPSnapClient) DeleteSnapshot(snapName string) (string, error) {

	resp, err := gsc.ComputeService.Snapshots.Delete(gsc.Project, snapName).Do()
	if err != nil {
		return "", errors.Wrap(err, "error deleting snapshot:")
	}

	return resp.SelfLink, nil
}

func parseOperationOut(operation *compute.Operation) (string, error) {

	// Get status (Possible values: "DONE", "PENDING", "RUNNING") and errors
	status := operation.Status
	if operation.Error != nil {
		var err_msgs []string
		for _, err := range operation.Error.Errors {
			err_msgs = append(err_msgs, err.Message)
		}
		return status, errors.New(strings.Join(err_msgs, ","))
	}
	return status, nil

}

func (gsc *GCPSnapClient) GetZonalOperationStatus(operation, zone string) (string, error) {

	// Format in case of link
	operation = formatLinkString(operation)
	zone = formatLinkString(zone)

	op, err := gsc.ComputeService.ZoneOperations.Get(gsc.Project, zone, operation).Do()
	if err != nil {
		return "", errors.Wrap(err, "error getting zonal operation:")
	}

	return parseOperationOut(op)
}

func (gsc *GCPSnapClient) GetGlobalOperationStatus(operation string) (string, error) {

	// Format in case of link
	operation = formatLinkString(operation)

	op, err := gsc.ComputeService.GlobalOperations.Get(gsc.Project, operation).Do()
	if err != nil {
		return "", errors.Wrap(err, "error getting global operation:")
	}

	return parseOperationOut(op)
}
