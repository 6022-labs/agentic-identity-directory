package settings

import (
	"slices"
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const ChainsSettingsKey = "chains"

const (
	defaultRequestTimeoutSeconds   = 10
	defaultMaxAddressesPerLogQuery = 10
)

// ChainConfig is how one chain's RPC is reached; MaxAddressesPerLogQuery caps the addresses of one eth_getLogs.
type ChainConfig struct {
	RpcHttpUrl              string
	RequestTimeout          time.Duration
	MaxAddressesPerLogQuery int
}

type chainConfigRaw struct {
	RpcHttpUrl              string `koanf:"rpc_http_url"`
	RequestTimeoutSeconds   uint64 `koanf:"request_timeout_seconds"`
	MaxAddressesPerLogQuery int    `koanf:"max_addresses_per_log_query"`
}

type ChainsSettings map[uint64]*ChainConfig

func NewChainsSettings(logger *zap.Logger, k *koanf.Koanf) *ChainsSettings {
	raw := map[uint64]*chainConfigRaw{}
	if err := k.UnmarshalWithConf(ChainsSettingsKey, &raw, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal chains settings", zap.Error(err))
	}

	settings := ChainsSettings{}
	for chainId, chain := range raw {
		if chain == nil {
			logger.Fatal("chain entry is nil", zap.Uint64("chainId", chainId))
		}
		chain.RpcHttpUrl = strings.TrimSpace(chain.RpcHttpUrl)
		if chain.RpcHttpUrl == "" {
			logger.Fatal("chain rpc_http_url is required", zap.Uint64("chainId", chainId))
		}
		if chain.RequestTimeoutSeconds == 0 {
			chain.RequestTimeoutSeconds = defaultRequestTimeoutSeconds
		}
		if chain.MaxAddressesPerLogQuery <= 0 {
			chain.MaxAddressesPerLogQuery = defaultMaxAddressesPerLogQuery
		}
		settings[chainId] = &ChainConfig{
			RpcHttpUrl:              chain.RpcHttpUrl,
			RequestTimeout:          time.Duration(chain.RequestTimeoutSeconds) * time.Second,
			MaxAddressesPerLogQuery: chain.MaxAddressesPerLogQuery,
		}
	}
	return &settings
}

func (s *ChainsSettings) ChainIds() []uint64 {
	if s == nil {
		return nil
	}
	chainIds := lo.Keys(*s)
	slices.Sort(chainIds)
	return chainIds
}

func (s *ChainsSettings) Get(chainId uint64) *ChainConfig {
	if s == nil {
		return nil
	}
	return (*s)[chainId]
}
