package settings

import (
	_ "embed"
	"encoding/json"
	"slices"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// SyncSettingsKey is shared with the blockchain adapter's chains settings, which ignore the sync fields.
const SyncSettingsKey = "chains"

const (
	defaultConfirmations       = 20
	defaultPollIntervalSeconds = 15
	defaultLogWindowBlocks     = 10_000
)

//go:embed sync_start_blocks.json
var syncStartBlocksJson []byte

type syncStartBlocksFile struct {
	StartBlocks []struct {
		ChainId    uint64 `json:"chainId"`
		StartBlock uint64 `json:"startBlock"`
	} `json:"startBlocks"`
}

// ChainSyncSettings is how one chain is indexed; Confirmations is how far behind the head it stays.
type ChainSyncSettings struct {
	StartBlock      uint64
	Confirmations   uint64
	PollInterval    time.Duration
	LogWindowBlocks uint64
}

type chainSyncSettingsRaw struct {
	Testnet             *bool   `koanf:"testnet"`
	Confirmations       *uint64 `koanf:"confirmations"`
	PollIntervalSeconds uint64  `koanf:"poll_interval_seconds"`
	LogWindowBlocks     uint64  `koanf:"log_window_blocks"`
}

type SyncSettings map[uint64]*ChainSyncSettings

// NewSyncSettings fails without a chain, on a chain 6022 is not deployed on, or on a testnet and mainnet mix.
func NewSyncSettings(logger *zap.Logger, k *koanf.Koanf) *SyncSettings {
	raw := map[uint64]*chainSyncSettingsRaw{}
	if err := k.UnmarshalWithConf(SyncSettingsKey, &raw, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal chains settings", zap.Error(err))
	}
	if len(raw) == 0 {
		logger.Fatal("configure at least one chain under chains (or CHAINS__<id>__RPC_HTTP_URL env)")
	}

	var file syncStartBlocksFile
	if err := json.Unmarshal(syncStartBlocksJson, &file); err != nil {
		logger.Fatal("failed to parse built-in sync_start_blocks.json", zap.Error(err))
	}
	startBlocks := map[uint64]uint64{}
	for _, entry := range file.StartBlocks {
		startBlocks[entry.ChainId] = entry.StartBlock
	}

	settings := SyncSettings{}
	testnets := map[bool]bool{}
	for chainId, chain := range raw {
		if chain == nil {
			logger.Fatal("chain entry is nil", zap.Uint64("chainId", chainId))
		}
		startBlock, ok := startBlocks[chainId]
		if !ok {
			logger.Fatal("6022 is not deployed on this chain", zap.Uint64("chainId", chainId))
		}
		// Required: defaulting to mainnet would let a forgotten testnet flag mix networks.
		if chain.Testnet == nil {
			logger.Fatal("chains.<id>.testnet is required", zap.Uint64("chainId", chainId))
		}
		testnets[*chain.Testnet] = true
		if chain.PollIntervalSeconds == 0 {
			chain.PollIntervalSeconds = defaultPollIntervalSeconds
		}
		if chain.LogWindowBlocks == 0 {
			chain.LogWindowBlocks = defaultLogWindowBlocks
		}
		settings[chainId] = &ChainSyncSettings{
			StartBlock:      startBlock,
			Confirmations:   lo.FromPtrOr(chain.Confirmations, defaultConfirmations),
			PollInterval:    time.Duration(chain.PollIntervalSeconds) * time.Second,
			LogWindowBlocks: chain.LogWindowBlocks,
		}
	}
	if len(testnets) > 1 {
		logger.Fatal("chains mix testnet and mainnet: a testnet agent would be discovered as a mainnet one")
	}

	return &settings
}

func (s *SyncSettings) ChainIds() []uint64 {
	chainIds := lo.Keys(*s)
	slices.Sort(chainIds)
	return chainIds
}

func (s *SyncSettings) Get(chainId uint64) *ChainSyncSettings {
	return (*s)[chainId]
}
