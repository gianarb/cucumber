package step

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"go.uber.org/zap"

	"github.com/gianarb/planner"
)

type ReconcileNodes struct {
	EC2svc        *ec2.EC2
	Tags          map[string]string
	VpcID         *string
	SubnetID      *string
	CurrentNumber int
	DesiredNumber int
	logger        *zap.Logger
}

func (s *ReconcileNodes) Name() string {
	return "reconcile-node"
}

func (s *ReconcileNodes) Do(ctx context.Context) ([]planner.Procedure, error) {
	s.logger.Info("need to reconcile number of running nodes", zap.Int("current", s.CurrentNumber), zap.Int("desired", s.DesiredNumber))
	steps := []planner.Procedure{}
	if s.CurrentNumber > s.DesiredNumber {
		for ii := s.DesiredNumber; ii < s.CurrentNumber; ii++ {
			// TODO: remove instances if they are too many
		}
	} else {
		ii := s.CurrentNumber
		if ii == 0 {
			ii = ii + 1
		}
		for i := ii; i < s.DesiredNumber; i++ {
			steps = append(steps, &RunInstance{
				EC2svc:   s.EC2svc,
				Tags:     s.Tags,
				VpcID:    s.VpcID,
				SubnetID: s.SubnetID,
			})
		}
	}
	return steps, nil
}

func (s *ReconcileNodes) WithLogger(logger *zap.Logger) {
	s.logger = logger.With(zap.String("step", s.Name()))
}
