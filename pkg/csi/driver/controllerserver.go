package driver

import (
	"context"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"
	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
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

	cloud, err := aws.GetAWSProvider()
	if err != nil {
		glog.V(3).Infof("Failed to get provider: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	const volNameTagKey = "VolumeName"
	volumes, err := cloud.GetVolumesByTagName(volNameTagKey, volName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var volID string
	if len(volumes) == 1 {
		volID = volumes[0]
	} else if len(volumes) > 1 {
		return nil, status.Error(codes.Internal, "multiple volumes reported by Cinder with same name")
	} else {
		v, err := cloud.CreateDisk(&aws.VolumeOptions{
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

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	cloud, err := aws.GetAWSProvider()
	if err != nil {
		glog.V(3).Infof("Failed to get provider: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = cloud.DeleteDisk(aws.KubernetesVolumeID(volID))
	if err != nil {
		glog.V(3).Infof("Failed to delete volume: %v", err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	cloud, err := aws.GetAWSProvider()
	if err != nil {
		glog.V(3).Infof("Failed to get provider: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Failed to get provider")
	}

	nodeID := types.NodeName(req.GetNodeId())
	devicePath, err := cloud.AttachDisk(aws.KubernetesVolumeID(volID), nodeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pInfo := map[string]string{
		"DevicePath": devicePath,
	}
	return &csi.ControllerPublishVolumeResponse{PublishInfo: pInfo}, nil
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}
