package services_test

import (
	"context"
	"math/big"
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenListingAgentRefsTestingSuite struct {
	sut *services.AgentCollectionRequester
}

func WhenListingAgentRefsBeforeEach(t *testing.T, stub *rpcStub) *WhenListingAgentRefsTestingSuite {
	sut, err := services.NewAgentCollectionRequester(zap.NewNop(), stub.ethClientFactory(t), stub.chainsSettings())
	require.NoError(t, err)
	return &WhenListingAgentRefsTestingSuite{sut: sut}
}

func TestWhenListingAgentRefs(t *testing.T) {
	t.Parallel()

	collection := strings.ToLower(common_tests.RandomEthAddress().Hex())

	t.Run("Given a collection claiming more agents than a collection can hold", func(t *testing.T) {
		t.Parallel()

		hugeNextTokenId := hexutil.Encode(common.LeftPadBytes(new(big.Int).Lsh(big.NewInt(1), 255).Bytes(), 32))

		t.Run("Should fail with ErrContractUnreadable instead of enumerating", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingAgentRefsBeforeEach(t, newRateLimitedRpcStub(t, 0, hugeNextTokenId))

			agentRefs, err := suite.sut.ListAgentRefs(context.Background(), stubChainId, collection, 10)

			assert.Nil(t, agentRefs)
			assert.ErrorIs(t, err, apperrors.ErrContractUnreadable)
		})
	})
}
