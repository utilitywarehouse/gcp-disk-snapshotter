package watch

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/metrics"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	compute "google.golang.org/api/compute/v1"
)

func TestCreateSnapshot(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Watcher with mocked GCPSnapClient interface
	mgsc := snapshot.NewMockGCPSnapClientInterface(mockCtrl)
	metrics := metrics.NewMockPrometheusInterface(mockCtrl)

	watcher := &Watcher{
		GSC:     mgsc,
		Metrics: metrics,
	}

	// test compute disk
	d := compute.Disk{
		Name: "test",
		Zone: "test",
	}
	// Channel used to wait for operation polling
	op_res := make(chan bool)

	// Successful Run
	gomock.InOrder(
		expectCreateSnapshotAndReturnSuccessfully(mgsc, d.Name, d.Zone),
		expectGetZonalOperationStatusAndWriteToChannel(mgsc, "op", d.Zone, op_res),
		expectUpdateOperationStatus(metrics, "zonal", true),
	)
	err := watcher.createSnapshot(d)
	if err != nil {
		t.Fatal(err)
	}
	waitForOp(op_res)

	// Error Run
	testErr := errors.New("test error")

	expectCreateSnapshotAndReturnError(mgsc, d.Name, d.Zone, testErr)

	err = watcher.createSnapshot(d)
	if err == nil {
		t.Fatal("No error returned!")
	}
	assert.EqualError(t, err, "test error")

}

func TestDeleteSnapshot(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	metrics := metrics.NewMockPrometheusInterface(mockCtrl)

	// Watcher with mocked GCPSnapClient interface
	mgsc := snapshot.NewMockGCPSnapClientInterface(mockCtrl)
	watcher := &Watcher{
		GSC:     mgsc,
		Metrics: metrics,
	}

	// Channel used to wait for operation polling
	op_res := make(chan bool)

	// Create a test snapshot to delete
	s := compute.Snapshot{
		Name: "test1",
	}

	gomock.InOrder(
		expectDeleteSnapshotAndReturnSuccessfully(mgsc, s.Name),
		expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res),
		expectUpdateOperationStatus(metrics, "global", true),
	)
	err := watcher.deleteSnapshot(s)
	if err != nil {
		t.Fatal(err)
	}
	waitForOp(op_res)

	// Error Run
	testErr := errors.New("test error")

	expectDeleteSnapshotAndReturnError(mgsc, s.Name, testErr)

	err = watcher.deleteSnapshot(s)
	if err == nil {
		t.Fatal("No error returned!")
	}
	assert.EqualError(t, err, "test error")

}

func waitForOp(op_res chan bool) {
	select {
	case <-op_res:
		break
	}
}

func expectCreateSnapshotAndReturnSuccessfully(gsc *snapshot.MockGCPSnapClientInterface, name, zone string) *gomock.Call {
	return gsc.EXPECT().CreateSnapshot(name, zone).Times(1).Return("op", nil)
}

func expectCreateSnapshotAndReturnError(gsc *snapshot.MockGCPSnapClientInterface, name, zone string, err error) *gomock.Call {
	return gsc.EXPECT().CreateSnapshot(name, zone).Times(1).Return("op", err)
}

func expectDeleteSnapshotAndReturnSuccessfully(gsc *snapshot.MockGCPSnapClientInterface, name string) *gomock.Call {
	return gsc.EXPECT().DeleteSnapshot(name).Times(1).Return("op", nil)
}

func expectDeleteSnapshotAndReturnError(gsc *snapshot.MockGCPSnapClientInterface, name string, err error) *gomock.Call {
	return gsc.EXPECT().DeleteSnapshot(name).Times(1).Return("op", err)
}

func expectGetZonalOperationStatusAndWriteToChannel(gsc *snapshot.MockGCPSnapClientInterface, operation, zone string, op_ch chan bool) *gomock.Call {
	return gsc.EXPECT().GetZonalOperationStatus(operation, zone).Times(1).Do(
		func(operation, zone string) {
			op_ch <- true
		},
	).Return("DONE", nil)
}

func expectGetGlobalOperationStatusAndWriteToChannel(gsc *snapshot.MockGCPSnapClientInterface, operation string, op_ch chan bool) *gomock.Call {
	return gsc.EXPECT().GetGlobalOperationStatus(operation).Times(1).Do(
		func(operation string) {
			op_ch <- true
		},
	).Return("DONE", nil)
}

func expectUpdateOperationStatus(m *metrics.MockPrometheusInterface, operation_type string, success bool) *gomock.Call {
	return m.EXPECT().UpdateOperationStatus(operation_type, success).Times(1).Return()
}
