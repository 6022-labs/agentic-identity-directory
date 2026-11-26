// Package fixtures deploys the 6022 contract stack onto a throwaway chain; see README.md.
package fixtures

import (
	"context"
	"embed"
	"fmt"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed artifacts/*.json
var artifactFiles embed.FS

// Mirror the contract's tuple layouts; mints and texts go through raw ABI packing.
type agentAddress struct {
	AddressType string `abi:"addressType"`
	Value       string `abi:"value"`
}

type keyValue struct {
	Key   string `abi:"key"`
	Value string `abi:"value"`
}

type nullableUint256 struct {
	IsSet bool     `abi:"isSet"`
	Value *big.Int `abi:"value"`
}

// AgentStack is a freshly deployed 6022 stack; the anvil account administers and moderates everything.
type AgentStack struct {
	Manager          common.Address
	Creator          common.Address
	AgentEnsRegistry common.Address

	deployer *contracts.Deployer
}

func DeployAgentStack(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
) (*AgentStack, error) {
	deployer, err := contracts.NewDeployer(ctx, client, anvilAccountSettings)
	if err != nil {
		return nil, err
	}

	manager, boundManager, err := deploy(deployer, "AgentCollectionsManager")
	if err != nil {
		return nil, err
	}
	descriptor, _, err := deploy(deployer, "AgentCollectionsDescriptor", "ipfs://")
	if err != nil {
		return nil, err
	}
	implementation, _, err := deploy(deployer, "AgentCollectionV1")
	if err != nil {
		return nil, err
	}
	creator, _, err := deploy(deployer, "AgentCollectionCreatorV1", manager, descriptor, implementation)
	if err != nil {
		return nil, err
	}
	agentEnsRegistry, _, err := deploy(deployer, "AgentEnsRegistry", manager, [32]byte{1})
	if err != nil {
		return nil, err
	}

	var creatorRole [32]byte
	if err := deployer.Call(boundManager, &creatorRole, "CREATOR_ROLE"); err != nil {
		return nil, err
	}
	if err := deployer.Transact(boundManager, "grantRole", creatorRole, creator); err != nil {
		return nil, err
	}

	return &AgentStack{
		Manager:          manager,
		Creator:          creator,
		AgentEnsRegistry: agentEnsRegistry,
		deployer:         deployer,
	}, nil
}

// CreateCollection creates a collection through the creator and returns its address.
func (s *AgentStack) CreateCollection(name string) (common.Address, error) {
	boundCreator, err := bindContract(s.deployer, "AgentCollectionCreatorV1", s.Creator)
	if err != nil {
		return common.Address{}, err
	}
	if err := s.deployer.Transact(
		boundCreator,
		"createCollection",
		name,
		strings.ToUpper(name),
		"agents that "+name,
		s.deployer.Address,
		s.deployer.Address,
	); err != nil {
		return common.Address{}, err
	}

	boundManager, err := bindContract(s.deployer, "AgentCollectionsManager", s.Manager)
	if err != nil {
		return common.Address{}, err
	}
	// Collections are 1-indexed, so the one just created is the last id.
	var nextCollectionId *big.Int
	if err := s.deployer.Call(boundManager, &nextCollectionId, "nextCollectionId"); err != nil {
		return common.Address{}, err
	}
	var collection common.Address
	err = s.deployer.Call(boundManager, &collection, "collections", new(big.Int).Sub(nextCollectionId, big.NewInt(1)))
	return collection, err
}

// MintAgent mints to the anvil account; evmAddress must be unique per collection.
func (s *AgentStack) MintAgent(collection common.Address, agentName string, evmAddress common.Address) error {
	return s.transact("AgentCollectionV1", collection, "mint",
		s.deployer.Address,
		agentName,
		[]agentAddress{{AddressType: "evm", Value: strings.ToLower(evmAddress.Hex())}},
		[]keyValue{{Key: "default", Value: "bafybeibsulc7sgpck3w7ur7eqa6fwaazcsswn3dg3rfl65zka3knajx43i"}},
		[]keyValue{{Key: "role", Value: "expert"}, {Key: "mission", Value: "integration testing"}},
		nullableUint256{IsSet: false, Value: big.NewInt(0)},
	)
}

func (s *AgentStack) SetAttribute(collection common.Address, tokenId int64, key, value string) error {
	return s.transact("AgentCollectionV1", collection, "addOrUpdateAgentAttribute", big.NewInt(tokenId), key, value)
}

func (s *AgentStack) UpdateDescription(collection common.Address, description string) error {
	return s.transact("AgentCollectionV1", collection, "updateDescription", description)
}

func (s *AgentStack) Transfer(collection common.Address, tokenId int64, to common.Address) error {
	return s.transact("AgentCollectionV1", collection, "transferFrom", s.deployer.Address, to, big.NewInt(tokenId))
}

// PublishUrl provisions the agent's ENS node with its url.
func (s *AgentStack) PublishUrl(collection common.Address, tokenId int64, url string) error {
	return s.transact("AgentEnsRegistry", s.AgentEnsRegistry, "initAgentTexts",
		collection, big.NewInt(tokenId), []keyValue{{Key: "url", Value: url}})
}

func (s *AgentStack) SetText(collection common.Address, tokenId int64, key, value string) error {
	return s.transact(
		"AgentEnsRegistry",
		s.AgentEnsRegistry,
		"setAgentText",
		collection,
		big.NewInt(tokenId),
		key,
		value,
	)
}

func (s *AgentStack) transact(contractName string, address common.Address, method string, args ...any) error {
	bound, err := bindContract(s.deployer, contractName, address)
	if err != nil {
		return err
	}
	return s.deployer.Transact(bound, method, args...)
}

func deploy(deployer *contracts.Deployer, name string, args ...any) (common.Address, *bind.BoundContract, error) {
	contract, err := load(name)
	if err != nil {
		return common.Address{}, nil, err
	}
	return deployer.Deploy(name, contract, args...)
}

func bindContract(deployer *contracts.Deployer, name string, address common.Address) (*bind.BoundContract, error) {
	contract, err := load(name)
	if err != nil {
		return nil, err
	}
	return deployer.Bind(contract, address), nil
}

func load(name string) (*contracts.CompiledContract, error) {
	raw, err := artifactFiles.ReadFile("artifacts/" + name + ".json")
	if err != nil {
		return nil, fmt.Errorf("no committed artifact for %s (see README.md): %w", name, err)
	}
	return contracts.Parse(raw)
}
