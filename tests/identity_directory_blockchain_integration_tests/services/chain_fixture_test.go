package services_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/metrics"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_blockchain_integration_tests/fixtures"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const anvilChainId = uint64(31337)

// chainFixture is a deployed 6022 stack on the shared anvil chain, with requesters wired to it.
type chainFixture struct {
	t      *testing.T
	stack  *fixtures.AgentStack
	client *ethclient.Client
	rpcUrl string
}

func newChainFixture(t *testing.T) *chainFixture {
	anvilAccountSettings := container_builders.GetAnvilContainerBuilder().Create()
	client, err := ethclient.Dial(anvilAccountSettings.HttpUrl)
	if err != nil {
		t.Fatalf("failed to dial anvil: %v", err)
	}
	t.Cleanup(client.Close)

	stack, err := fixtures.DeployAgentStack(context.Background(), client, anvilAccountSettings)
	if err != nil {
		t.Fatalf("failed to deploy the contract stack: %v", err)
	}
	return &chainFixture{t: t, stack: stack, client: client, rpcUrl: anvilAccountSettings.HttpUrl}
}

func (f *chainFixture) chainsSettings(rpcUrl string, maxAddressesPerLogQuery int) *settings.ChainsSettings {
	return &settings.ChainsSettings{anvilChainId: {
		RpcHttpUrl:              rpcUrl,
		RequestTimeout:          5 * time.Second,
		MaxAddressesPerLogQuery: maxAddressesPerLogQuery,
	}}
}

func (f *chainFixture) ethClientFactory(chainsSettings *settings.ChainsSettings) *factory.EthClientFactory {
	return factory.NewEthClientFactory(chainsSettings, metrics.NewBlockchainRpcMetrics())
}

func (f *chainFixture) agentContractsSettings() *settings.AgentContractsSettings {
	return &settings.AgentContractsSettings{anvilChainId: {
		AgentCollectionsManager: f.stack.Manager,
		AgentEnsRegistry:        f.stack.AgentEnsRegistry,
	}}
}

func (f *chainFixture) blockchainRequester() *services.BlockchainRequester {
	chainsSettings := f.chainsSettings(f.rpcUrl, 10)
	return services.NewBlockchainRequester(zap.NewNop(), f.ethClientFactory(chainsSettings), chainsSettings)
}

func (f *chainFixture) agentCollectionsManagerRequester() *services.AgentCollectionsManagerRequester {
	chainsSettings := f.chainsSettings(f.rpcUrl, 10)
	requester, err := services.NewAgentCollectionsManagerRequester(
		zap.NewNop(), f.ethClientFactory(chainsSettings), chainsSettings, f.agentContractsSettings())
	if err != nil {
		f.t.Fatalf("failed to build the manager requester: %v", err)
	}
	return requester
}

func (f *chainFixture) agentCollectionRequesterThrough(
	rpcUrl string,
	maxAddressesPerLogQuery int,
) *services.AgentCollectionRequester {
	chainsSettings := f.chainsSettings(rpcUrl, maxAddressesPerLogQuery)
	requester, err := services.NewAgentCollectionRequester(
		zap.NewNop(),
		f.ethClientFactory(chainsSettings),
		chainsSettings,
	)
	if err != nil {
		f.t.Fatalf("failed to build the collection requester: %v", err)
	}
	return requester
}

func (f *chainFixture) agentCollectionRequester() *services.AgentCollectionRequester {
	return f.agentCollectionRequesterThrough(f.rpcUrl, 10)
}

func (f *chainFixture) agentEnsRegistryRequester() *services.AgentEnsRegistryRequester {
	chainsSettings := f.chainsSettings(f.rpcUrl, 10)
	requester, err := services.NewAgentEnsRegistryRequester(
		zap.NewNop(), f.ethClientFactory(chainsSettings), chainsSettings, f.agentContractsSettings())
	if err != nil {
		f.t.Fatalf("failed to build the ens registry requester: %v", err)
	}
	return requester
}

func (f *chainFixture) currentBlock() uint64 {
	blockNumber, err := f.client.BlockNumber(context.Background())
	if err != nil {
		f.t.Fatalf("failed to read the head block: %v", err)
	}
	return blockNumber
}

func (f *chainFixture) must(err error) {
	if err != nil {
		f.t.Fatalf("chain setup failed: %v", err)
	}
}

// rangeRefusingRpcUrl proxies anvil but refuses every eth_getLogs over more than maxBlocks, like a capped provider.
func (f *chainFixture) rangeRefusingRpcUrl(maxBlocks uint64) (string, *atomic.Int32) {
	target, err := url.Parse(f.rpcUrl)
	if err != nil {
		f.t.Fatalf("failed to parse the anvil url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	var refusals atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		if strings.Contains(string(body), "eth_getLogs") && f.blockSpan(string(body)) > maxBlocks {
			refusals.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + f.requestId(string(body)) +
				`,"error":{"code":-32602,"message":"block range too large"}}`))
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	f.t.Cleanup(server.Close)
	return server.URL, &refusals
}

func (f *chainFixture) blockSpan(body string) uint64 {
	from := f.hexField(body, `"fromBlock":"`)
	to := f.hexField(body, `"toBlock":"`)
	return to - from + 1
}

func (f *chainFixture) hexField(body, prefix string) uint64 {
	start := strings.Index(body, prefix)
	if start < 0 {
		return 0
	}
	rest := body[start+len(prefix):]
	value, err := strconv.ParseUint(strings.TrimPrefix(rest[:strings.Index(rest, `"`)], "0x"), 16, 64)
	if err != nil {
		f.t.Fatalf("failed to parse %s: %v", prefix, err)
	}
	return value
}

func (f *chainFixture) requestId(body string) string {
	start := strings.Index(body, `"id":`) + len(`"id":`)
	end := strings.IndexAny(body[start:], ",}")
	return body[start : start+end]
}
