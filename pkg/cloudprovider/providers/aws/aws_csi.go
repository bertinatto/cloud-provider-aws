package aws

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang/glog"
)

//var (
//once        sync.Once
//configFile  string
//AWSInstance *Cloud
//)

//func InitAWSProvider(cfg string) {
//once.Do(func() { configFile = cfg })
//}

var (
	AWSInstance *Cloud = nil
	configFile  string = ""
)

func InitAWSProvider(cfg string) {
	configFile = cfg
}

func GetAWSProvider() (*Cloud, error) {
	if AWSInstance != nil {
		return AWSInstance, nil
	}

	var r io.Reader
	if configFile != "" {
		f, err := os.Open(configFile)
		if err != nil {
			return nil, fmt.Errorf("unable to open AWS config file: %v", err)
		}
		defer f.Close()
		r = f
	}

	cfg, err := readAWSCloudConfig(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read AWS config file: %v", err)
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}

	var provider credentials.Provider
	if cfg.Global.RoleARN == "" {
		provider = &ec2rolecreds.EC2RoleProvider{
			Client: ec2metadata.New(sess),
		}
	} else {
		glog.Infof("Using AWS assumed role %v", cfg.Global.RoleARN)
		provider = &stscreds.AssumeRoleProvider{
			Client:  sts.New(sess),
			RoleARN: cfg.Global.RoleARN,
		}
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			provider,
			&credentials.SharedCredentialsProvider{},
		})

	aws := newAWSSDKProvider(creds)
	return newAWSCloud(*cfg, aws)

}
