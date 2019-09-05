package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/gianarb/cucumber"
	"github.com/gianarb/cucumber/plan"
	"github.com/gianarb/planner"
	"go.uber.org/zap"
)

const modeReconcile = "reconcile"
const modeOneShot = "one-shot"

func main() {
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger = logger.With(zap.String("app", "cucumber"))

	hostedZoneID := os.Getenv("AWS_HOSTED_ZONE")
	if hostedZoneID == "" {
		logger.Fatal("AWS_HOSTED_ZONE is required. Please set it.")
	}
	requestPath := os.Getenv("CUCUMBER_REQUEST")
	if requestPath == "" {
		logger.Fatal("CUCUMBER_REQUEST is required. Please set it.")
	}

	mode := os.Getenv("CUCUMBER_MODE")
	if mode != modeReconcile {
		mode = modeOneShot
	}

	logger.Info("cucumber starts", zap.String("mode", mode))

	req, err := cucumber.ParseRequestFromFile(requestPath)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("impossible to parse the request file. Please check it.")
	}

	p := plan.CreatePlan{
		ClusterName:  req.Name,
		NodesNumber:  req.NodesNumber,
		DNSRecord:    req.DNSName,
		HostedZoneID: hostedZoneID,
		Tags: map[string]string{
			"app":          "cucumber",
			"cluster-name": req.Name,
		},
	}

	scheduler := planner.NewScheduler()
	scheduler.WithLogger(logger)

	if err := scheduler.Execute(ctx, &p); err != nil {
		logger.With(zap.Error(err)).Fatal("cucumber ended with an error")
	}

	wg := sync.WaitGroup{}

	if mode == modeReconcile {
		wg.Add(1)
		go func() {
			logger := logger.With(zap.String("from", "reconciliation"))
			scheduler.WithLogger(logger)
			for {
				logger.Info("reconciliation loop started")
				if err := scheduler.Execute(ctx, &p); err != nil {
					logger.With(zap.Error(err)).Warn("cucumber reconciliation failed.")
				}
				time.Sleep(10 * time.Second)
				logger.Info("reconciliation loop ended")
			}
		}()
	}
	wg.Wait()

	logger.Info("cucumber ended without any error")
}
