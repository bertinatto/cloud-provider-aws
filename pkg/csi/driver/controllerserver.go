package driver

import (
	"context"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"
	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

var volTags = make(map[string]string)

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume name not provided")
	}

	// Volume already exists, so returns it
	if volID, ok := volTags[volName]; ok {
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				Id: volID,
			},
		}, nil
	}

	cloud, err := aws.GetAWSProvider()
	if err != nil {
		glog.V(3).Infof("Failed to get provider: %v", err)
		return nil, err
	}

	opts := &aws.VolumeOptions{CapacityGB: 4}
	v, err := cloud.CreateDisk(opts)
	if err != nil {
		glog.V(3).Infof("Failed to create volume: %v", err)
		return nil, nil
	}
	volID := string(v)

	volTags[volName] = volID
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

	if id, ok := volTags[volID]; ok {
		volID = id
	}

	cloud, err := aws.GetAWSProvider()
	if err != nil {
		glog.V(3).Infof("Failed to get provider: %v", err)
		return nil, err
	}

	_, err = cloud.DeleteDisk(aws.KubernetesVolumeID(volID))
	if err != nil {
		glog.V(3).Infof("Failed to delete volume: %v", err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, nil
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}
