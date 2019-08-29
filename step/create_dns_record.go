package step

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/gianarb/cucumber"
	"go.uber.org/zap"
)

type CreateDNSRecord struct {
	Route53Svc   *route53.Route53
	DNSRecord    string
	HostedZoneID string
	TargetIPs    []*string
	logger       *zap.Logger
}

func (s *CreateDNSRecord) Name() string {
	return "create_dns_record"
}

func (s *CreateDNSRecord) Do(ctx context.Context) ([]cucumber.Procedure, error) {
	var err error
	steps := []cucumber.Procedure{}

	rr := []*route53.ResourceRecord{}
	for _, ip := range s.TargetIPs {
		rr = append(rr, &route53.ResourceRecord{Value: ip})
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionCreate),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(s.DNSRecord),
						Type:            aws.String(route53.RRTypeA),
						SetIdentifier:   aws.String("cucumber app"),
						TTL:             aws.Int64(100),
						ResourceRecords: rr,
						Weight:          aws.Int64(100),
					},
				},
			},
			Comment: aws.String("from cucumber app"),
		},
		HostedZoneId: aws.String(s.HostedZoneID),
	}

	_, err = s.Route53Svc.ChangeResourceRecordSets(params)
	if err != nil {
		s.logger.Warn("DNS Record creation failed", zap.Error(err))
		return steps, nil
	}

	// Hack to allow DNS propagation
	time.Sleep(5 * time.Second)

	return steps, err
}

func (s *CreateDNSRecord) WithLogger(logger *zap.Logger) {
	s.logger = logger.With(zap.String("step", s.Name()))
}
