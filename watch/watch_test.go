package watch

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/gcp-disk-snapshotter/snapshot"
	compute "google.golang.org/api/compute/v1"
)

func TestCreateSnapshot(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Watcher with mocked GCPSnapClient interface
	mgsc := snapshot.NewMockGCPSnapClientInterface(mockCtrl)
	watcher := &Watcher{
		GSC: mgsc,
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
	)
	err := watcher.createSnapshot(d)
	if err != nil {
		t.Fatal(err)
	}
	waitForOp(op_res)

	// Error Run
	testErr := errors.New("test error")
	//gomock.InOrder(
	expectCreateSnapshotAndReturnError(mgsc, d.Name, d.Zone, testErr)
	//)
	err = watcher.createSnapshot(d)
	if err == nil {
		t.Fatal("No error returned!")
	}
	assert.EqualError(t, err, "test error")

}

func TestDeleteSnapshots(t *testing.T) {

	mockCtrl := gomock.NewController(t)

	// Watcher with mocked GCPSnapClient interface
	mgsc := snapshot.NewMockGCPSnapClientInterface(mockCtrl)
	watcher := &Watcher{
		GSC: mgsc,
	}

	// Channel used to wait for operation polling
	op_res := make(chan bool)

	// Empty List
	sl := []compute.Snapshot{}
	err := watcher.deleteSnapshots(sl)
	if err != nil {
		t.Fatal(err)
	}

	// Add a test snapshot to the list
	s := compute.Snapshot{
		Name: "test1",
	}
	sl = append(sl, s)

	gomock.InOrder(
		expectDeleteSnapshotAndReturnSuccessfully(mgsc, sl[0].Name),
		expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res),
	)
	err = watcher.deleteSnapshots(sl)
	if err != nil {
		t.Fatal(err)
	}
	waitForOp(op_res)

	// test multiple
	s = compute.Snapshot{
		Name: "test2",
	}
	sl = append(sl, s)

	// We cannot expect calls to keep order on multiple snapshots deletion
	expectDeleteSnapshotAndReturnSuccessfully(mgsc, sl[0].Name)
	expectDeleteSnapshotAndReturnSuccessfully(mgsc, sl[1].Name)
	expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res)
	expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res)

	err = watcher.deleteSnapshots(sl)
	if err != nil {
		t.Fatal(err)
	}
	waitForOp(op_res)
	waitForOp(op_res)

	// Error Run
	testErr := errors.New("test error")
	// We cannot expect calls to keep order on multiple snapshots deletion
	expectDeleteSnapshotAndReturnSuccessfully(mgsc, sl[0].Name)
	expectDeleteSnapshotAndReturnError(mgsc, sl[1].Name, testErr)
	expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res)
	expectGetGlobalOperationStatusAndWriteToChannel(mgsc, "op", op_res)

	err = watcher.deleteSnapshots(sl)
	if err == nil {
		t.Fatal("No error returned!")
	}
	assert.EqualError(t, err, "test error")

	// Wait for 1 operation since the other call fails
	waitForOp(op_res)
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
