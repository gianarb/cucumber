package step

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gianarb/planner"
	"go.uber.org/zap"
)

type RunInstance struct {
	EC2svc   *ec2.EC2
	Tags     map[string]string
	VpcID    *string
	SubnetID *string
	logger   *zap.Logger
}

func (s *RunInstance) Name() string {
	return "run-instance"
}

func (s *RunInstance) Do(ctx context.Context) ([]planner.Procedure, error) {
	tags := []*ec2.Tag{}
	for k, v := range s.Tags {
		if k == "cluster-name" {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(v),
			})
		}
		tags = append(tags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	steps := []planner.Procedure{}
	instanceInput := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0378588b4ae11ec24"),
		InstanceType: aws.String("t2.micro"),
		//UserData:              &userData,
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
		SubnetId: s.SubnetID,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags:         tags,
			},
		},
	}
	_, err := s.EC2svc.RunInstances(instanceInput)
	if err != nil {
		return steps, err
	}
	return steps, nil
}

func (s *RunInstance) WithLogger(logger *zap.Logger) {
	s.logger = logger.With(zap.String("step", s.Name()))
}
