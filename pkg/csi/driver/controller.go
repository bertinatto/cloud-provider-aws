package driver

import (
	"context"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"
	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume name not provided")
	}

	// TODO: magic number?
	// Default volume size is 4 GiB
	volSizeBytes := int64(4 * 1024 * 1024 * 1024)
	if req.GetCapacityRange() != nil {
		volSizeBytes = int64(req.GetCapacityRange().GetRequiredBytes())
	}
	volSizeGB := int(volumeutil.RoundUpSize(volSizeBytes, 1024*1024*1024))

	const volNameTagKey = "VolumeName"
	volumes, err := d.cloud.GetVolumesByTagName(volNameTagKey, volName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var volID string
	if len(volumes) == 1 {
		volID = volumes[0]
	} else if len(volumes) > 1 {
		return nil, status.Error(codes.Internal, "multiple volumes with same name")
	} else {
		v, err := d.cloud.CreateDisk(&aws.VolumeOptions{
			CapacityGB: volSizeGB,
			Tags:       map[string]string{volNameTagKey: volName},
		})
		if err != nil {
			glog.V(3).Infof("Failed to create volume: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		awsID, err := v.MapToAWSVolumeID()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		volID = string(awsID)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id: volID,
			//Attributes: map[string]string{
			//"availability": "",
			//},
		},
	}, nil
}

func (d *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	_, err := d.cloud.DeleteDisk(aws.KubernetesVolumeID(volID))
	if err != nil {
		glog.V(3).Infof("Failed to delete volume: %v", err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	nodeID := types.NodeName(req.GetNodeId())
	devicePath, err := d.cloud.AttachDisk(aws.KubernetesVolumeID(volID), nodeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pInfo := map[string]string{
		"DevicePath": devicePath,
	}
	return &csi.ControllerPublishVolumeResponse{PublishInfo: pInfo}, nil
}

func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var caps []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	return resp, nil
}

func (d *Driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
