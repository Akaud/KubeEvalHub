package runner

import (
	"context"
	"log"
	"time"

	"agent/internal/config"
	"agent/internal/kube"
	"agent/internal/model"
	"agent/internal/transport"
)

type Runner struct {
	cfg     config.Config
	kube    *kube.Clients
	backend *transport.Client
}

func New(cfg config.Config, kubeClients *kube.Clients, backend *transport.Client) *Runner {
	return &Runner{
		cfg:     cfg,
		kube:    kubeClients,
		backend: backend,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.cfg.ScrapeInterval)
	defer ticker.Stop()

	for {
		if err := r.runOnce(ctx); err != nil {
			log.Printf("runOnce error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Runner) runOnce(ctx context.Context) error {
	cluster, err := kube.DiscoverCluster(ctx, r.kube)
	if err != nil {
		return err
	}

	samples, err := kube.CollectSamples(ctx, r.kube)
	if err != nil {
		return err
	}

	if len(samples) == 0 {
		return nil
	}

	req := model.PushMetricsRequest{
		Cluster: cluster,
		Samples: samples,
	}

	if err := r.backend.PushMetrics(ctx, req); err != nil {
		return err
	}

	_ = r.backend.Heartbeat(ctx)

	return nil
}
