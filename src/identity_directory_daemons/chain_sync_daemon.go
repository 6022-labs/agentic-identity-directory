package identity_directory_daemons

import (
	"context"
	"sync"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"go.uber.org/zap"
)

// ChainSyncDaemon runs one sync loop per chain; a failed pass is retried at the next tick, other chains go on.
type ChainSyncDaemon struct {
	logger       *zap.Logger
	syncSettings *settings.SyncSettings
	syncChain    *use_cases.SyncChain
	cancel       context.CancelFunc
	waitGroup    sync.WaitGroup
}

func NewChainSyncDaemon(
	logger *zap.Logger,
	syncSettings *settings.SyncSettings,
	syncChain *use_cases.SyncChain,
) *ChainSyncDaemon {
	return &ChainSyncDaemon{logger: logger, syncSettings: syncSettings, syncChain: syncChain}
}

func (d *ChainSyncDaemon) Start(ctx context.Context) {
	ctx, d.cancel = context.WithCancel(ctx)
	for _, chainId := range d.syncSettings.ChainIds() {
		pollInterval := d.syncSettings.Get(chainId).PollInterval
		d.waitGroup.Go(func() { d.run(ctx, chainId, pollInterval) })
	}
}

func (d *ChainSyncDaemon) Stop() {
	if d.cancel != nil {
		d.cancel()
	}
	d.waitGroup.Wait()
}

func (d *ChainSyncDaemon) run(ctx context.Context, chainId uint64, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		err := d.syncChain.Execute(ctx, &requests.SyncChainRequest{ChainId: chainId})
		if err != nil && ctx.Err() == nil {
			d.logger.Error("chain sync pass failed", zap.Uint64("chainId", chainId), zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
