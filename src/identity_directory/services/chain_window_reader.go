package services

import (
	"context"
	"errors"
	"fmt"

	common_errors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// ChainWindowReader turns a block range into fresh state: logs only say what changed, the contracts say to what.
type ChainWindowReader struct {
	logger                           *zap.Logger
	agentCollectionsManagerRequester interfaces.AgentCollectionsManagerRequesterInterface
	agentCollectionRequester         interfaces.AgentCollectionRequesterInterface
	agentEnsRegistryRequester        interfaces.AgentEnsRegistryRequesterInterface
	agentCollectionRepository        repositories.AgentCollectionRepositoryInterface
	ensNameBuilder                   interfaces.EnsNameBuilderInterface
}

func NewChainWindowReader(
	logger *zap.Logger,
	agentCollectionsManagerRequester interfaces.AgentCollectionsManagerRequesterInterface,
	agentCollectionRequester interfaces.AgentCollectionRequesterInterface,
	agentEnsRegistryRequester interfaces.AgentEnsRegistryRequesterInterface,
	agentCollectionRepository repositories.AgentCollectionRepositoryInterface,
	ensNameBuilder interfaces.EnsNameBuilderInterface,
) *ChainWindowReader {
	return &ChainWindowReader{
		logger:                           logger,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentCollectionRequester:         agentCollectionRequester,
		agentEnsRegistryRequester:        agentEnsRegistryRequester,
		agentCollectionRepository:        agentCollectionRepository,
		ensNameBuilder:                   ensNameBuilder,
	}
}

func (r *ChainWindowReader) Read(
	ctx context.Context,
	chainId, fromBlock, toBlock, headBlock uint64,
) (*models.ChainWindow, error) {
	knownAddresses, err := r.agentCollectionRepository.ListAddresses(ctx, chainId)
	if err != nil {
		return nil, fmt.Errorf("list known collections: %w", err)
	}
	createdAddresses, err := r.agentCollectionsManagerRequester.ListCreatedCollections(ctx, chainId, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("list created collections: %w", err)
	}
	// A collection created in this window has logs in it too.
	collectionAddresses := lo.Uniq(append(knownAddresses, createdAddresses...))

	changes, err := r.agentCollectionRequester.ListChanges(ctx, chainId, collectionAddresses, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("list collection changes: %w", err)
	}
	ensRecords, err := r.agentEnsRegistryRequester.ListTextChanges(ctx, chainId, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("list ens text changes: %w", err)
	}

	collections, err := r.readCollections(
		ctx,
		chainId,
		lo.Uniq(append(createdAddresses, changes.CollectionAddresses...)),
		headBlock,
	)
	if err != nil {
		return nil, err
	}

	agentRefs := changes.AgentRefs
	// addCollection can register a contract that already minted: its agents emit nothing new.
	for _, createdAddress := range createdAddresses {
		existingRefs, err := r.agentCollectionRequester.ListAgentRefs(ctx, chainId, createdAddress, headBlock)
		if errors.Is(err, common_errors.ErrContractUnreadable) {
			r.logger.Warn("skipping an unreadable collection", zap.Uint64("chainId", chainId),
				zap.String("collection", createdAddress), zap.Error(err))
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("list agents of collection %s: %w", createdAddress, err)
		}
		agentRefs = append(agentRefs, existingRefs...)
	}

	agents, err := r.readAgents(ctx, chainId, lo.Uniq(agentRefs), collections, headBlock)
	if err != nil {
		return nil, err
	}

	knownCollections := lo.SliceToMap(collectionAddresses, func(address string) (string, bool) { return address, true })
	return &models.ChainWindow{
		ChainId:     chainId,
		FromBlock:   fromBlock,
		ToBlock:     toBlock,
		Collections: collections,
		Agents:      agents,
		EnsRecords: lo.Filter(ensRecords, func(record *models.AgentEnsRecord, _ int) bool {
			record.Sanitize()
			return knownCollections[record.CollectionAddress]
		}),
	}, nil
}

func (r *ChainWindowReader) readCollections(
	ctx context.Context,
	chainId uint64,
	addresses []string,
	headBlock uint64,
) ([]*models.AgentCollection, error) {
	collections := make([]*models.AgentCollection, 0, len(addresses))
	for _, address := range addresses {
		collection, err := r.agentCollectionRequester.GetCollection(ctx, chainId, address, headBlock)
		if errors.Is(err, common_errors.ErrContractUnreadable) {
			r.logger.Warn("skipping an unreadable collection", zap.Uint64("chainId", chainId),
				zap.String("collection", address), zap.Error(err))
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read collection %s: %w", address, err)
		}
		collection.Sanitize()
		collections = append(collections, collection)
	}
	return collections, nil
}

func (r *ChainWindowReader) readAgents(
	ctx context.Context,
	chainId uint64,
	agentRefs []models.AgentRef,
	collections []*models.AgentCollection,
	headBlock uint64,
) ([]*models.Agent, error) {
	collectionNames := lo.SliceToMap(collections, func(collection *models.AgentCollection) (string, string) {
		return collection.Address, collection.Name
	})

	agents := make([]*models.Agent, 0, len(agentRefs))
	for _, agentRef := range agentRefs {
		collectionName, err := r.collectionName(ctx, chainId, agentRef.CollectionAddress, collectionNames)
		if err != nil {
			return nil, err
		}
		if collectionName == "" {
			r.logger.Warn("skipping an agent of an unindexed collection", zap.Uint64("chainId", chainId),
				zap.String("collection", agentRef.CollectionAddress), zap.String("tokenId", agentRef.TokenId))
			continue
		}

		agent, err := r.agentCollectionRequester.GetAgent(ctx, chainId, agentRef, headBlock)
		if errors.Is(err, common_errors.ErrContractUnreadable) {
			r.logger.Warn("skipping an unreadable agent", zap.Uint64("chainId", chainId),
				zap.String("collection", agentRef.CollectionAddress), zap.String("tokenId", agentRef.TokenId),
				zap.Error(err))
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read agent %s/%s: %w", agentRef.CollectionAddress, agentRef.TokenId, err)
		}
		agent.Sanitize()
		agent.EnsName = r.ensNameBuilder.Build(chainId, collectionName, agent.Name)
		agents = append(agents, agent)
	}
	return agents, nil
}

// collectionName is empty when the collection is in neither this window nor the index.
func (r *ChainWindowReader) collectionName(
	ctx context.Context,
	chainId uint64,
	address string,
	collectionNames map[string]string,
) (string, error) {
	if name, ok := collectionNames[address]; ok {
		return name, nil
	}
	collection, err := r.agentCollectionRepository.Get(ctx, chainId, address)
	if err != nil {
		return "", fmt.Errorf("get collection %s: %w", address, err)
	}
	if collection == nil {
		collectionNames[address] = ""
		return "", nil
	}
	collectionNames[address] = collection.Name
	return collection.Name, nil
}
