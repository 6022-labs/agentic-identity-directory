package services_test

import (
	"context"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/abi"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenListingCollectionChangesTestingSuite struct {
	stub *rpcStub
	logs *observer.ObservedLogs
	sut  *services.AgentCollectionRequester
}

func WhenListingCollectionChangesBeforeEach(
	t *testing.T,
	stub *rpcStub,
) *WhenListingCollectionChangesTestingSuite {
	core, observed := observer.New(zapcore.WarnLevel)
	sut, err := services.NewAgentCollectionRequester(zap.New(core), stub.ethClientFactory(t), stub.chainsSettings())
	require.NoError(t, err)
	return &WhenListingCollectionChangesTestingSuite{stub: stub, logs: observed, sut: sut}
}

// collectionLogs builds the eth_getLogs entries one collection emits.
type collectionLogs struct {
	collection common.Address
}

func (c collectionLogs) entry(logIndex int, data []byte, topics ...common.Hash) map[string]any {
	return map[string]any{
		"address":          c.collection.Hex(),
		"topics":           topics,
		"data":             hexutil.Encode(data),
		"blockNumber":      "0x5",
		"transactionHash":  common.Hash{1}.Hex(),
		"transactionIndex": "0x0",
		"blockHash":        common.Hash{2}.Hex(),
		"logIndex":         hexutil.EncodeUint64(uint64(logIndex)),
		"removed":          false,
	}
}

func TestWhenListingCollectionChanges(t *testing.T) {
	t.Parallel()

	contractAbi, err := abi.AgentCollectionV1MetaData.GetAbi()
	require.NoError(t, err)
	collection := common_tests.RandomEthAddress()
	owner := common.BytesToHash(common_tests.RandomEthAddress().Bytes())

	t.Run("Given a node that caps the logs one query returns", func(t *testing.T) {
		t.Parallel()

		capped := func(call int32, _ string) rpcAnswer {
			if call == 1 {
				return rpcAnswer{
					err: map[string]any{"code": -32005, "message": "query returned more than 10000 results"},
				}
			}
			return rpcAnswer{result: []map[string]any{}}
		}

		t.Run("Should split the range at once instead of retrying it", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionChangesBeforeEach(t, newRpcStub(t, capped))

			_, err := suite.sut.ListChanges(
				context.Background(), stubChainId, []string{strings.ToLower(collection.Hex())}, 1, 10)

			require.NoError(t, err)
			assert.Equal(t, int32(3), suite.stub.calls.Load())
			assert.Zero(t, suite.logs.FilterMessage("rpc call failed, retrying").Len())
		})
	})

	t.Run("Given a node gateway that fails on a heavy range", func(t *testing.T) {
		t.Parallel()

		overloaded := func(call int32, _ string) rpcAnswer {
			if call == 1 {
				return rpcAnswer{status: http.StatusBadGateway}
			}
			return rpcAnswer{result: []map[string]any{}}
		}

		t.Run("Should split the range at once instead of retrying it", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionChangesBeforeEach(t, newRpcStub(t, overloaded))

			_, err := suite.sut.ListChanges(
				context.Background(), stubChainId, []string{strings.ToLower(collection.Hex())}, 1, 10)

			require.NoError(t, err)
			assert.Equal(t, int32(3), suite.stub.calls.Load())
			assert.Zero(t, suite.logs.FilterMessage("rpc call failed, retrying").Len())
		})
	})

	t.Run("Given a node that refuses a one-block query", func(t *testing.T) {
		t.Parallel()

		refusing := func(int32, string) rpcAnswer {
			return rpcAnswer{err: map[string]any{"code": -32601, "message": "the method eth_getLogs does not exist"}}
		}

		t.Run("Should fail without retrying", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionChangesBeforeEach(t, newRpcStub(t, refusing))

			_, err := suite.sut.ListChanges(
				context.Background(), stubChainId, []string{strings.ToLower(collection.Hex())}, 5, 5)

			assert.ErrorContains(t, err, "does not exist")
			assert.Equal(t, int32(1), suite.stub.calls.Load())
		})
	})

	t.Run("Given a log of a known event in another layout", func(t *testing.T) {
		t.Parallel()

		logs := collectionLogs{collection: collection}
		tokenId := func(value int64) []byte { return common.LeftPadBytes(big.NewInt(value).Bytes(), 32) }
		// An ERC-20 style Transfer: same topic, two indexed arguments instead of three.
		erc20Transfer := logs.entry(0, tokenId(1), contractAbi.Events["Transfer"].ID, owner, owner)
		minted := logs.entry(1, tokenId(7), contractAbi.Events["Minted"].ID, owner)

		t.Run("Should skip that log with a warning and keep the others", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionChangesBeforeEach(t,
				newRateLimitedRpcStub(t, 0, []map[string]any{erc20Transfer, minted}))

			changes, err := suite.sut.ListChanges(
				context.Background(), stubChainId, []string{strings.ToLower(collection.Hex())}, 1, 10)

			require.NoError(t, err)
			assert.Equal(t, []models.AgentRef{
				{CollectionAddress: strings.ToLower(collection.Hex()), TokenId: "7"},
			}, changes.AgentRefs)
			assert.Equal(t, 1, suite.logs.FilterMessage("skipping an unparsable collection log").Len())
		})
	})
}
