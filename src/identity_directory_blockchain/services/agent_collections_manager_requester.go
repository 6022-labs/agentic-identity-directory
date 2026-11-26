package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/abi"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type AgentCollectionsManagerRequester struct {
	AbstractRpcRequester
	agentContractsSettings *settings.AgentContractsSettings
	filterer               *abi.AgentCollectionsManagerFilterer
	collectionCreatedTopic common.Hash
}

func NewAgentCollectionsManagerRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	chainsSettings *settings.ChainsSettings,
	agentContractsSettings *settings.AgentContractsSettings,
) (*AgentCollectionsManagerRequester, error) {
	contractAbi, err := abi.AgentCollectionsManagerMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("parse agent collections manager abi: %w", err)
	}
	filterer, err := abi.NewAgentCollectionsManagerFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerRequester{
		AbstractRpcRequester: AbstractRpcRequester{
			logger:           logger,
			ethClientFactory: ethClientFactory,
			chainsSettings:   chainsSettings,
		},
		agentContractsSettings: agentContractsSettings,
		filterer:               filterer,
		collectionCreatedTopic: contractAbi.Events["CollectionCreated"].ID,
	}, nil
}

func (r *AgentCollectionsManagerRequester) ListCreatedCollections(
	ctx context.Context,
	chainId, fromBlock, toBlock uint64,
) ([]string, error) {
	contracts := r.agentContractsSettings.Get(chainId)
	if contracts == nil {
		return nil, fmt.Errorf("no 6022 contracts on chain %d", chainId)
	}

	logs, err := r.filterLogs(
		ctx,
		chainId,
		[]common.Address{contracts.AgentCollectionsManager},
		[]common.Hash{r.collectionCreatedTopic},
		fromBlock,
		toBlock,
	)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0, len(logs))
	for _, log := range logs {
		event, err := r.filterer.ParseCollectionCreated(log)
		if err != nil {
			return nil, fmt.Errorf("parse CollectionCreated log: %w", err)
		}
		addresses = append(addresses, strings.ToLower(event.CollectionAddress.Hex()))
	}
	return addresses, nil
}
