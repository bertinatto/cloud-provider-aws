all: aws-driver

aws-csi-driver:
	mkdir -p bin
	go build -o bin/aws-csi-driver ./cmd/aws-csi-plugin

test:
	go test github.com/bertinatto/cloud-provider-aws/pkg/csi/driver
