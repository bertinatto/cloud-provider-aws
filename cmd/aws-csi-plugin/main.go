package main

import (
	"flag"

	"github.com/bertinatto/cloud-provider-aws/pkg/csi/driver"
)

func main() {
	var (
		endpoint  = flag.String("endpoint", "unix://tmp/csi.sock", "CSI Endpoint")
		nodeID    = flag.String("node", "CSINode", "Node ID")
		keyID     = flag.String("key", "", "AWS Access Key ID")
		secretKey = flag.String("secret", "", "Secret Access Key")
		region    = flag.String("region", "", "Region")
	)
	flag.Parse()

	d := driver.NewDriver(*endpoint, *nodeID, *keyID, *secretKey, *region)
	d.Run()
}
