package cucumber

import (
	"context"

	"go.uber.org/zap"
)

type Scheduler struct {
	logger *zap.Logger
}

func (s *Scheduler) WithLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *Scheduler) Execute(ctx context.Context, p Plan) error {
	p.WithLogger(s.logger)
	s.logger.Info("Started execution plan " + p.Name())
	for {
		steps, err := p.Create(ctx)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
		if len(steps) == 0 {
			break
		}
		err = s.react(ctx, steps)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}
	s.logger.Info("Plan executed without errors.")
	return nil
}

func (s *Scheduler) react(ctx context.Context, steps []Procedure) error {
	for _, step := range steps {
		step.WithLogger(s.logger)
		innerSteps, err := step.Do(ctx)
		if err != nil {
			return err
		}
		if len(innerSteps) > 0 {
			if err := s.react(ctx, innerSteps); err != nil {
				return err
			}
		}
	}
	return nil
}
