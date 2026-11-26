package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	common_errors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/abi"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// agentEventNames are the events after which an agent must be read again; the logs cannot rebuild it.
var agentEventNames = []string{
	"Minted",
	"Transfer",
	"AgentAttributeUpdated",
	"AgentAddressAdded",
	"AgentAddressRemoved",
	"AgentImageUpdated",
	"AgentImageRemoved",
}

const (
	descriptionUpdatedEventName = "DescriptionUpdated"
	// maxAgentsPerCollection bounds what a hostile collection can make us enumerate.
	maxAgentsPerCollection = 100_000
)

type AgentCollectionRequester struct {
	AbstractRpcRequester
	filterer      *abi.AgentCollectionV1Filterer
	topics        []common.Hash
	eventNames    map[common.Hash]string
	agentRefFuncs map[string]func(types.Log) (*big.Int, error)
}

func NewAgentCollectionRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	chainsSettings *settings.ChainsSettings,
) (*AgentCollectionRequester, error) {
	contractAbi, err := abi.AgentCollectionV1MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("parse agent collection abi: %w", err)
	}
	filterer, err := abi.NewAgentCollectionV1Filterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}

	requester := &AgentCollectionRequester{
		AbstractRpcRequester: AbstractRpcRequester{
			logger:           logger,
			ethClientFactory: ethClientFactory,
			chainsSettings:   chainsSettings,
		},
		filterer:   filterer,
		eventNames: map[common.Hash]string{},
	}
	for _, eventName := range append(agentEventNames, descriptionUpdatedEventName) {
		topic := contractAbi.Events[eventName].ID
		requester.topics = append(requester.topics, topic)
		requester.eventNames[topic] = eventName
	}
	requester.agentRefFuncs = map[string]func(types.Log) (*big.Int, error){
		"Minted":                requester.mintedTokenId,
		"Transfer":              requester.transferTokenId,
		"AgentAttributeUpdated": requester.attributeUpdatedTokenId,
		"AgentAddressAdded":     requester.addressAddedTokenId,
		"AgentAddressRemoved":   requester.addressRemovedTokenId,
		"AgentImageUpdated":     requester.imageUpdatedTokenId,
		"AgentImageRemoved":     requester.imageRemovedTokenId,
	}
	return requester, nil
}

func (r *AgentCollectionRequester) ListChanges(
	ctx context.Context,
	chainId uint64,
	collectionAddresses []string,
	fromBlock, toBlock uint64,
) (*models.CollectionChanges, error) {
	addresses := lo.Map(collectionAddresses, func(address string, _ int) common.Address {
		return common.HexToAddress(address)
	})
	logs, err := r.filterLogs(ctx, chainId, addresses, r.topics, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}

	changes := &models.CollectionChanges{}
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		collectionAddress := strings.ToLower(log.Address.Hex())
		eventName := r.eventNames[log.Topics[0]]
		if eventName == descriptionUpdatedEventName {
			changes.CollectionAddresses = append(changes.CollectionAddresses, collectionAddress)
			continue
		}
		tokenIdOf, ok := r.agentRefFuncs[eventName]
		if !ok {
			continue
		}
		tokenId, err := tokenIdOf(log)
		if err != nil {
			// A collection with another event layout would otherwise stall the chain on this window forever.
			r.logger.Warn("skipping an unparsable collection log", zap.Uint64("chainId", chainId),
				zap.String("collection", collectionAddress), zap.String("event", eventName), zap.Error(err))
			continue
		}
		changes.AgentRefs = append(changes.AgentRefs, models.AgentRef{
			CollectionAddress: collectionAddress,
			TokenId:           tokenId.String(),
		})
	}
	changes.CollectionAddresses = lo.Uniq(changes.CollectionAddresses)
	changes.AgentRefs = lo.Uniq(changes.AgentRefs)
	return changes, nil
}

func (r *AgentCollectionRequester) GetCollection(
	ctx context.Context,
	chainId uint64,
	collectionAddress string,
	atBlock uint64,
) (*models.AgentCollection, error) {
	caller, opts, err := r.caller(ctx, chainId, collectionAddress, atBlock)
	if err != nil {
		return nil, err
	}

	collection := &models.AgentCollection{ChainId: chainId, Address: collectionAddress}
	err = r.callContract(ctx, func() error {
		var err error
		if collection.Name, err = caller.Name(opts); err != nil {
			return err
		}
		if collection.Symbol, err = caller.Symbol(opts); err != nil {
			return err
		}
		collection.Description, err = caller.Description(opts)
		return err
	})
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (r *AgentCollectionRequester) ListAgentRefs(
	ctx context.Context,
	chainId uint64,
	collectionAddress string,
	atBlock uint64,
) ([]models.AgentRef, error) {
	caller, opts, err := r.caller(ctx, chainId, collectionAddress, atBlock)
	if err != nil {
		return nil, err
	}

	var nextTokenId *big.Int
	err = r.callContract(ctx, func() error {
		var err error
		nextTokenId, err = caller.NextTokenId(opts)
		return err
	})
	if err != nil {
		return nil, err
	}

	if !nextTokenId.IsInt64() || nextTokenId.Int64() > maxAgentsPerCollection+1 {
		return nil, fmt.Errorf("%w: nextTokenId %s is out of range", common_errors.ErrContractUnreadable, nextTokenId)
	}

	// Token ids run from 1 and the contract cannot burn.
	agentRefs := []models.AgentRef{}
	for tokenId := big.NewInt(1); tokenId.Cmp(nextTokenId) < 0; tokenId.Add(tokenId, big.NewInt(1)) {
		agentRefs = append(agentRefs, models.AgentRef{CollectionAddress: collectionAddress, TokenId: tokenId.String()})
	}
	return agentRefs, nil
}

func (r *AgentCollectionRequester) GetAgent(
	ctx context.Context,
	chainId uint64,
	agentRef models.AgentRef,
	atBlock uint64,
) (*models.Agent, error) {
	caller, opts, err := r.caller(ctx, chainId, agentRef.CollectionAddress, atBlock)
	if err != nil {
		return nil, err
	}
	tokenId, ok := new(big.Int).SetString(agentRef.TokenId, 10)
	if !ok {
		return nil, fmt.Errorf("token id %q is not a decimal number", agentRef.TokenId)
	}

	var information abi.AgentInformation
	var owner common.Address
	err = r.callContract(ctx, func() error {
		var err error
		if information, err = caller.InformationOf(opts, tokenId); err != nil {
			return err
		}
		owner, err = caller.OwnerOf(opts, tokenId)
		return err
	})
	if err != nil {
		return nil, err
	}

	agent := &models.Agent{
		ChainId:           chainId,
		CollectionAddress: agentRef.CollectionAddress,
		TokenId:           agentRef.TokenId,
		Name:              information.Name,
		Creator:           strings.ToLower(information.Creator.Hex()),
		Owner:             strings.ToLower(owner.Hex()),
		Images:            r.keyValues(information.Images),
		Attributes:        r.keyValues(information.Attributes),
	}
	if information.CloneOf.IsSet {
		cloneOf := information.CloneOf.Value.String()
		agent.CloneOf = &cloneOf
	}
	for _, address := range information.Addresses {
		agent.Addresses = append(agent.Addresses, models.AgentAddress{
			ChainId:           chainId,
			CollectionAddress: agentRef.CollectionAddress,
			TokenId:           agentRef.TokenId,
			Type:              address.AddressType,
			Value:             address.Value,
		})
	}
	return agent, nil
}

func (r *AgentCollectionRequester) caller(
	ctx context.Context,
	chainId uint64,
	collectionAddress string,
	atBlock uint64,
) (*abi.AgentCollectionV1Caller, *bind.CallOpts, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, nil, err
	}
	caller, err := abi.NewAgentCollectionV1Caller(common.HexToAddress(collectionAddress), client)
	if err != nil {
		return nil, nil, err
	}
	return caller, &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(atBlock)}, nil
}

func (r *AgentCollectionRequester) keyValues(keyValues []abi.KeyValue) map[string]string {
	values := make(map[string]string, len(keyValues))
	for _, keyValue := range keyValues {
		values[keyValue.Key] = keyValue.Value
	}
	return values
}

func (r *AgentCollectionRequester) mintedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseMinted(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) transferTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseTransfer(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) attributeUpdatedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseAgentAttributeUpdated(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) addressAddedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseAgentAddressAdded(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) addressRemovedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseAgentAddressRemoved(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) imageUpdatedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseAgentImageUpdated(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}

func (r *AgentCollectionRequester) imageRemovedTokenId(log types.Log) (*big.Int, error) {
	event, err := r.filterer.ParseAgentImageRemoved(log)
	if err != nil {
		return nil, err
	}
	return event.TokenId, nil
}
