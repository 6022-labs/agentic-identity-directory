package services_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	metrics_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_blockchain_mocks/metrics_mocks/interfaces_mocks"
	"go.uber.org/mock/gomock"
)

const stubChainId = uint64(80002)

// rpcAnswer is one JSON-RPC reply; a non-zero status answers that HTTP status instead.
type rpcAnswer struct {
	status int
	result any
	err    map[string]any
}

// rpcStub replies to the n-th call (from 1) with what answer returns for it.
type rpcStub struct {
	calls  atomic.Int32
	server *httptest.Server
}

func newRpcStub(t *testing.T, answer func(call int32, body string) rpcAnswer) *rpcStub {
	stub := &rpcStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var request struct {
			Id json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(body, &request)
		reply := answer(stub.calls.Add(1), string(body))
		if reply.status != 0 {
			http.Error(w, http.StatusText(reply.status), reply.status)
			return
		}
		response := map[string]any{"jsonrpc": "2.0", "id": request.Id}
		if reply.err != nil {
			response["error"] = reply.err
		} else {
			response["result"] = reply.result
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(stub.server.Close)
	return stub
}

// newRateLimitedRpcStub answers 429 to the first rateLimited calls, then result.
func newRateLimitedRpcStub(t *testing.T, rateLimited int32, result any) *rpcStub {
	return newRpcStub(t, func(call int32, _ string) rpcAnswer {
		if call <= rateLimited {
			return rpcAnswer{status: http.StatusTooManyRequests}
		}
		return rpcAnswer{result: result}
	})
}

func (s *rpcStub) chainsSettings() *settings.ChainsSettings {
	return &settings.ChainsSettings{stubChainId: {
		RpcHttpUrl: s.server.URL, RequestTimeout: 5 * time.Second, MaxAddressesPerLogQuery: 10,
	}}
}

func (s *rpcStub) ethClientFactory(t *testing.T) *factory.EthClientFactory {
	mockBlockchainRpcMetrics := metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t))
	mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	return factory.NewEthClientFactory(s.chainsSettings(), mockBlockchainRpcMetrics)
}
