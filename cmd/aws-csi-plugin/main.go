package main

import (
	"flag"
	"log"

	"github.com/bertinatto/cloud-provider-aws/pkg/cloudprovider/providers/aws"
	"github.com/bertinatto/cloud-provider-aws/pkg/csi/driver"
)

func main() {
	var (
		endpoint = flag.String("endpoint", "unix://tmp/csi.sock", "CSI Endpoint")
		nodeID   = flag.String("node", "CSINode", "Node ID")
	)
	flag.Parse()

	cloudProvider, err := aws.GetAWSProvider()
	if err != nil {
		log.Fatalln(err)
	}

	drv := driver.NewDriver(cloudProvider, *endpoint, *nodeID)
	if err := drv.Run(); err != nil {
		log.Fatalln(err)
	}
}
