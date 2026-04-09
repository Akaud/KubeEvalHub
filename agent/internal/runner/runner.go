package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync/atomic"
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

	cluster model.ClusterPayload

	lastInventoryHash string
	lastInventorySent time.Time

	metricsRunning   int32
	inventoryRunning int32
	heartbeatRunning int32
}

func New(cfg config.Config, kubeClients *kube.Clients, backend *transport.Client) *Runner {
	return &Runner{
		cfg:     cfg,
		kube:    kubeClients,
		backend: backend,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	cluster, err := kube.DiscoverCluster(ctx, r.kube)
	if err != nil {
		return err
	}
	r.cluster = cluster

	metricsInterval := r.cfg.ScrapeInterval
	if metricsInterval <= 0 {
		metricsInterval = time.Minute
	}

	inventoryInterval := metricsInterval * 10
	if inventoryInterval < 5*time.Minute {
		inventoryInterval = 5 * time.Minute
	}

	heartbeatInterval := metricsInterval
	if heartbeatInterval <= 0 {
		heartbeatInterval = time.Minute
	}

	metricsTicker := time.NewTicker(metricsInterval)
	inventoryTicker := time.NewTicker(inventoryInterval)
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer metricsTicker.Stop()
	defer inventoryTicker.Stop()
	defer heartbeatTicker.Stop()

	// initial sync (blocking is fine here)
	if err := r.runInventory(ctx, true); err != nil {
		log.Printf("initial inventory sync error: %v", err)
	}
	if err := r.runMetrics(ctx); err != nil {
		log.Printf("initial metrics push error: %v", err)
	}
	if err := r.runHeartbeat(ctx); err != nil {
		log.Printf("initial heartbeat error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-inventoryTicker.C:
			if !atomic.CompareAndSwapInt32(&r.inventoryRunning, 0, 1) {
				continue
			}

			go func() {
				defer atomic.StoreInt32(&r.inventoryRunning, 0)

				ctxTimeout, cancel := context.WithTimeout(context.Background(), r.cfg.RequestTimeout*2)
				defer cancel()

				if err := r.runInventory(ctxTimeout, false); err != nil {
					log.Printf("inventory sync error: %v", err)
				}
			}()

		case <-metricsTicker.C:
			if !atomic.CompareAndSwapInt32(&r.metricsRunning, 0, 1) {
				continue
			}

			go func() {
				defer atomic.StoreInt32(&r.metricsRunning, 0)

				ctxTimeout, cancel := context.WithTimeout(context.Background(), r.cfg.RequestTimeout)
				defer cancel()

				if err := r.runMetrics(ctxTimeout); err != nil {
					log.Printf("metrics push error: %v", err)
				}
			}()

		case <-heartbeatTicker.C:
			if !atomic.CompareAndSwapInt32(&r.heartbeatRunning, 0, 1) {
				continue
			}

			go func() {
				defer atomic.StoreInt32(&r.heartbeatRunning, 0)

				ctxTimeout, cancel := context.WithTimeout(context.Background(), r.cfg.RequestTimeout)
				defer cancel()

				if err := r.runHeartbeat(ctxTimeout); err != nil {
					log.Printf("heartbeat error: %v", err)
				}
			}()
		}
	}
}

func (r *Runner) runInventory(ctx context.Context, force bool) error {
	inventory, err := kube.CollectInventory(ctx, r.kube)
	if err != nil {
		return err
	}

	revisionHash, err := computeInventoryRevisionHash(inventory)
	if err != nil {
		return err
	}

	if !force && revisionHash == r.lastInventoryHash {
		return nil
	}

	collectedAt := time.Now().UTC()

	req := model.PushInventoryRequest{
		Cluster:      r.cluster,
		CollectedAt:  collectedAt,
		RevisionHash: revisionHash,
		Inventory:    inventory,
	}

	if err := r.backend.PushInventory(ctx, req); err != nil {
		return err
	}

	r.lastInventoryHash = revisionHash
	r.lastInventorySent = collectedAt
	return nil
}

func (r *Runner) runMetrics(ctx context.Context) error {
	samples, err := kube.CollectSamples(ctx, r.kube)
	if err != nil {
		return err
	}
	if len(samples) == 0 {
		return nil
	}

	req := model.PushMetricsRequest{
		Cluster:          r.cluster,
		BatchCollectedAt: time.Now().UTC(),
		Source:           "metrics-server",
		Samples:          samples,
	}

	return r.backend.PushMetrics(ctx, req)
}

func (r *Runner) runHeartbeat(ctx context.Context) error {
	return r.backend.Heartbeat(ctx)
}

func computeInventoryRevisionHash(inventory model.InventoryPayload) (string, error) {
	normalized, err := normalizeInventoryForHash(inventory)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeInventoryForHash(inventory model.InventoryPayload) (model.InventoryPayload, error) {
	b, err := json.Marshal(inventory)
	if err != nil {
		return model.InventoryPayload{}, err
	}

	var out model.InventoryPayload
	if err := json.Unmarshal(b, &out); err != nil {
		return model.InventoryPayload{}, err
	}

	sortInventoryPayload(&out)
	return out, nil
}
