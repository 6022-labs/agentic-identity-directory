package settings

import (
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed agent_contracts.json
var agentContractsJson []byte

type agentContractsFile struct {
	Chains []struct {
		ChainId                 uint64 `json:"chainId"`
		AgentCollectionsManager string `json:"agentCollectionsManager"`
		AgentEnsRegistry        string `json:"agentEnsRegistry"`
	} `json:"chains"`
}

type AgentContracts struct {
	AgentCollectionsManager common.Address
	AgentEnsRegistry        common.Address
}

// AgentContractsSettings holds the 6022 contracts of every configured chain, baked into the build.
type AgentContractsSettings map[uint64]*AgentContracts

// NewAgentContractsSettings fails when a configured chain has no baked contracts.
func NewAgentContractsSettings(logger *zap.Logger, chainsSettings *ChainsSettings) *AgentContractsSettings {
	var file agentContractsFile
	if err := json.Unmarshal(agentContractsJson, &file); err != nil {
		logger.Fatal("failed to parse built-in agent_contracts.json", zap.Error(err))
	}

	baked := map[uint64]*AgentContracts{}
	for _, chain := range file.Chains {
		if !common.IsHexAddress(chain.AgentCollectionsManager) || !common.IsHexAddress(chain.AgentEnsRegistry) {
			logger.Fatal("agent_contracts.json holds an invalid address", zap.Uint64("chainId", chain.ChainId))
		}
		baked[chain.ChainId] = &AgentContracts{
			AgentCollectionsManager: common.HexToAddress(chain.AgentCollectionsManager),
			AgentEnsRegistry:        common.HexToAddress(chain.AgentEnsRegistry),
		}
	}

	settings := AgentContractsSettings{}
	for _, chainId := range chainsSettings.ChainIds() {
		contracts, ok := baked[chainId]
		if !ok {
			logger.Fatal("6022 is not deployed on this chain", zap.Uint64("chainId", chainId))
		}
		settings[chainId] = contracts
	}
	return &settings
}

func (s *AgentContractsSettings) Get(chainId uint64) *AgentContracts {
	if s == nil {
		return nil
	}
	return (*s)[chainId]
}
