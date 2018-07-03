package driver

import (
	"context"
	"testing"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	project           = "test-project"
	zone              = "test-zone"
	node              = "test-node"
	driver            = "test-driver"
	defaultVolumeSize = 10
)

// Create Volume Tests
func TestCreateVolumeArguments(t *testing.T) {
	// Define "normal" parameters
	stdVolCap := []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	stdCapRange := &csi.CapacityRange{RequiredBytes: int64(20 * 1024 * 1024 * 1024)}
	stdParams := map[string]string{
		"zone": zone,
		"type": "test-type",
	}

	testCases := []struct {
		name       string
		req        *csi.CreateVolumeRequest
		expVol     *csi.Volume
		expErrCode codes.Code
	}{
		{
			name: "success normal",
			req: &csi.CreateVolumeRequest{
				Name:               "vol-test-123",
				CapacityRange:      stdCapRange,
				VolumeCapabilities: stdVolCap,
				Parameters:         stdParams,
			},
			expVol: &csi.Volume{
				CapacityBytes: GBToBytes(20),
				Id:            project + "/" + zone + "/" + "test-vol",
				Attributes:    nil,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Logf("Test case: %s", tc.name)
		awsDriver := NewDriver(&aws.FakeCloudProvider{}, "", "")

		resp, err := awsDriver.CreateVolume(context.TODO(), tc.req)
		//check response
		if err != nil {
			serverError, ok := status.FromError(err)
			if !ok {
				t.Fatalf("Could not get error status code from err: %v", serverError)
			}
			if serverError.Code() != tc.expErrCode {
				t.Fatalf("Expected error code: %v, got: %v", tc.expErrCode, serverError.Code())
			}
			continue
		}
		if tc.expErrCode != codes.OK {
			t.Fatalf("Expected error: %v, got no error", tc.expErrCode)
		}

		// Make sure responses match
		vol := resp.GetVolume()
		if vol == nil {
			// If one is nil but not both
			t.Fatalf("Expected volume %v, got nil volume", tc.expVol)
		}

		//if vol.GetCapacityBytes() != tc.expVol.GetCapacityBytes() {
		//t.Fatalf("Expected volume capacity bytes: %v, got: %v", vol.GetCapacityBytes(), tc.expVol.GetCapacityBytes())
		//}

		//if vol.GetId() != tc.expVol.GetId() {
		//t.Fatalf("Expected volume id: %v, got: %v", vol.GetId(), tc.expVol.GetId())
		//}

		//for akey, aval := range tc.expVol.GetAttributes() {
		//if gotVal, ok := vol.GetAttributes()[akey]; !ok || gotVal != aval {
		//t.Fatalf("Expected volume attribute for key %v: %v, got: %v", akey, aval, gotVal)
		//}
		//}
		//if tc.expVol.GetAttributes() == nil && vol.GetAttributes() != nil {
		//t.Fatalf("Expected volume attributes to be nil, got: %#v", vol.GetAttributes())
		//}

	}
}

func GBToBytes(num int) int64 {
	return int64(num * 1024 * 1024 * 1024)
}
