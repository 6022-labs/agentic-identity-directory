package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/abi"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type AgentEnsRegistryRequester struct {
	AbstractRpcRequester
	agentContractsSettings *settings.AgentContractsSettings
	filterer               *abi.AgentEnsRegistryFilterer
	agentTextUpdatedTopic  common.Hash
	parentNodeUpdatedTopic common.Hash
	managerUpdatedTopic    common.Hash
}

func NewAgentEnsRegistryRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	chainsSettings *settings.ChainsSettings,
	agentContractsSettings *settings.AgentContractsSettings,
) (*AgentEnsRegistryRequester, error) {
	contractAbi, err := abi.AgentEnsRegistryMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("parse agent ens registry abi: %w", err)
	}
	filterer, err := abi.NewAgentEnsRegistryFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}
	return &AgentEnsRegistryRequester{
		AbstractRpcRequester: AbstractRpcRequester{
			logger:           logger,
			ethClientFactory: ethClientFactory,
			chainsSettings:   chainsSettings,
		},
		agentContractsSettings: agentContractsSettings,
		filterer:               filterer,
		agentTextUpdatedTopic:  contractAbi.Events["AgentTextUpdated"].ID,
		parentNodeUpdatedTopic: contractAbi.Events["ParentNodeUpdated"].ID,
		managerUpdatedTopic:    contractAbi.Events["ManagerUpdated"].ID,
	}, nil
}

func (r *AgentEnsRegistryRequester) ListTextChanges(
	ctx context.Context,
	chainId, fromBlock, toBlock uint64,
) ([]*models.AgentEnsRecord, error) {
	contracts := r.agentContractsSettings.Get(chainId)
	if contracts == nil {
		return nil, fmt.Errorf("no 6022 contracts on chain %d", chainId)
	}

	logs, err := r.filterLogs(
		ctx,
		chainId,
		[]common.Address{contracts.AgentEnsRegistry},
		[]common.Hash{r.agentTextUpdatedTopic, r.parentNodeUpdatedTopic, r.managerUpdatedTopic},
		fromBlock,
		toBlock,
	)
	if err != nil {
		return nil, err
	}

	records := make([]*models.AgentEnsRecord, 0, len(logs))
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		if log.Topics[0] != r.agentTextUpdatedTopic {
			// Records are keyed by agent, not by node: a new parent or manager leaves them stale until a resync.
			r.logger.Warn("AgentEnsRegistry admin change: resync this chain",
				zap.Uint64("chainId", chainId), zap.Uint64("block", log.BlockNumber))
			continue
		}
		event, err := r.filterer.ParseAgentTextUpdated(log)
		if err != nil {
			return nil, fmt.Errorf("parse AgentTextUpdated log: %w", err)
		}
		records = append(records, &models.AgentEnsRecord{
			ChainId:           chainId,
			CollectionAddress: strings.ToLower(event.Collection.Hex()),
			TokenId:           event.TokenId.String(),
			Key:               event.Key,
			Value:             event.Value,
		})
	}
	return records, nil
}
