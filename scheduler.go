package cucumber

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Scheduler struct {
	stepCounter int
	logger      *zap.Logger
}

func (s *Scheduler) WithLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *Scheduler) Execute(ctx context.Context, p Plan) error {
	uuidGenerator := uuid.New()
	logger := s.logger.With(zap.String("execution_id", uuidGenerator.String()))
	start := time.Now()
	p.WithLogger(logger)
	logger.Info("Started execution plan " + p.Name())
	s.stepCounter = 0
	for {
		steps, err := p.Create(ctx)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		if len(steps) == 0 {
			break
		}
		err = s.react(ctx, steps, logger)
		if err != nil {
			logger.Error(err.Error(), zap.String("execution_time", time.Since(start).String()), zap.Int("step_executed", s.stepCounter))
			return err
		}
	}
	logger.Info("Plan executed without errors.", zap.String("execution_time", time.Since(start).String()), zap.Int("step_executed", s.stepCounter))
	return nil
}

func (s *Scheduler) react(ctx context.Context, steps []Procedure, logger *zap.Logger) error {
	for _, step := range steps {
		s.stepCounter = s.stepCounter + 1
		step.WithLogger(logger)
		innerSteps, err := step.Do(ctx)
		if err != nil {
			logger.Error("Step failed.", zap.String("step", step.Name()), zap.Error(err))
			return err
		}
		if len(innerSteps) > 0 {
			if err := s.react(ctx, innerSteps, logger); err != nil {
				return err
			}
		}
	}
	return nil
}
