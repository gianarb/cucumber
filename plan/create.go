package plan

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/gianarb/cucumber"
	"github.com/gianarb/cucumber/step"
	"go.uber.org/zap"
)

type CreatePlan struct {
	ClusterName  string
	NodesNumber  int
	DNSRecord    string
	HostedZoneID string
	Tags         map[string]string
	logger       *zap.Logger
}

func (p *CreatePlan) Name() string {
	return "create"
}

func (p *CreatePlan) Create(ctx context.Context) ([]cucumber.Procedure, error) {
	var err error
	steps := []cucumber.Procedure{}

	sess := session.Must(session.NewSession())
	ec2Svc := ec2.New(sess, &aws.Config{Region: aws.String("us-east-1")})
	route52Svc := route53.New(sess, &aws.Config{Region: aws.String("us-east-1")})
	describeVpcsOutput, err := ec2Svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:app"),
				Values: []*string{aws.String("cucumber")},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(describeVpcsOutput.Vpcs) == 0 {
		return nil, fmt.Errorf("No vpc found with the tag app=cucumber")
	}

	vpcID := describeVpcsOutput.Vpcs[0].VpcId
	p.logger.Info("Found VPC", zap.String("vpc_id", *vpcID))
	subnetOutput, err := ec2Svc.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(subnetOutput.Subnets) == 0 {
		return nil, fmt.Errorf("No subnets found in vpc %s.", *vpcID)
	}

	subnetID := subnetOutput.Subnets[0].SubnetId

	p.logger.Info("Found subnet", zap.String("subnet_id", *subnetID))

	resp, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("pending"), aws.String("running")},
			},
			{
				Name:   aws.String("tag:cluster-name"),
				Values: []*string{aws.String(p.ClusterName)},
			},
			{
				Name:   aws.String("tag:app"),
				Values: []*string{aws.String("cucumber")},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	currentInstances := countInstancesByResp(resp)
	if len(currentInstances) != p.NodesNumber {
		steps = append(steps, &step.ReconcileNodes{
			EC2svc:        ec2Svc,
			Tags:          p.Tags,
			VpcID:         vpcID,
			SubnetID:      subnetID,
			CurrentNumber: len(currentInstances),
			DesiredNumber: p.NodesNumber,
		})
	}

	p.logger.Info("Checking if DNS already exists", zap.String("dns", p.DNSRecord))

	targetIPs := []*string{}
	for _, instance := range currentInstances {
		targetIPs = append(targetIPs, instance.PublicIpAddress)
	}

	if ips, err := net.LookupIP(p.DNSRecord); err != nil && len(targetIPs) > 0 {
		steps = append(steps, &step.CreateDNSRecord{
			Route53Svc:   route52Svc,
			DNSRecord:    p.DNSRecord,
			HostedZoneID: p.HostedZoneID,
			TargetIPs:    targetIPs,
		})
	} else if len(ips) != len(targetIPs) {
		//TODO: This condition is not enough. We need to reconcile IPs also if
		//the targetIPS and the IP set in the record are different.
		steps = append(steps, &step.UpdateDNSRecord{
			Route53Svc:   route52Svc,
			DNSRecord:    p.DNSRecord,
			HostedZoneID: p.HostedZoneID,
			TargetIPs:    targetIPs,
		})
	}

	return steps, nil
}

func (p *CreatePlan) WithLogger(logger *zap.Logger) {
	p.logger = logger.With(zap.String("plan_name", p.Name()))
}

// countInstancesByResp calculate the number of instances returned by the ec2.DescribeInstacesOutput
func countInstancesByResp(resp *ec2.DescribeInstancesOutput) []*ec2.Instance {
	instances := []*ec2.Instance{}
	for _, reservation := range resp.Reservations {
		for _, iii := range reservation.Instances {
			instances = append(instances, iii)
		}
	}
	return instances
}
