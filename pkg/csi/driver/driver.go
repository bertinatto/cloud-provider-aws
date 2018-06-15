package driver

import (
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"
	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

const (
	driverName    = "csi-aws"
	driverVersion = "0.1"
)

type Driver struct {
	endpoint string
	nodeID   string

	keyID     string
	secretKey string
	region    string

	csiDriver *csicommon.CSIDriver
}

func NewDriver(endpoint, nodeID, key, secret, region string) *Driver {
	glog.Infof("Driver: %v version: %v", driverName, driverVersion)

	csiDriver := csicommon.NewCSIDriver(driverName, driverVersion, nodeID)

	caps := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	}
	csiDriver.AddControllerServiceCapabilities(caps)

	vcs := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	csiDriver.AddVolumeCapabilityAccessModes(vcs)

	return &Driver{
		endpoint:  endpoint,
		keyID:     key,
		secretKey: secret,
		region:    region,
		csiDriver: csiDriver,
	}
}

func NewControllerServer(d *Driver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
	}
}

func NewNodeServer(d *Driver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
	}
}

func (d *Driver) Run() {
	//aws.InitAWSProvider("/home/fjb/.aws/credentials")
	aws.InitAWSProvider("")
	csicommon.RunControllerandNodePublishServer(d.endpoint, d.csiDriver, NewControllerServer(d), NewNodeServer(d))
}

func (d *Driver) Stop() {
	return
}
