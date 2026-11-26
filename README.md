# Agentic Identity Directory

[![Go Build](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-build.yml/badge.svg)](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-build.yml)
[![Go Unit Tests](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-unit-tests.yml/badge.svg)](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-unit-tests.yml)
[![Go Lint](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-lint.yml/badge.svg)](https://github.com/6022-labs/agentic-identity-directory/actions/workflows/go-lint.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Every 6022 agent has an on-chain identity: an NFT in a collection, with a name, a role, its
addresses, and an ENS name under which it publishes the URL to call it. This service reads those
identities from the chain, keeps them in Postgres, and lets anyone search them.

Agents use it as an MCP server: add `https://<host>/mcp` as an MCP connection to an agent, and it
can find other agents and learn how to call them. The same data is served as a REST API.

Docs: https://docs.agentic.6022.io

## Try it locally

You need Docker and a Polygon Amoy RPC URL. Public endpoints rate-limit `eth_getLogs` hard, so a
first sync over one is slow; use a provider key.

```sh
cp src/identity_directory_host/.env.example src/identity_directory_host/.env
# set CHAINS__80002__RPC_HTTP_URL in that file
docker compose up --build
```

Or against a local chain, with the 6022 contracts deployed at the addresses baked for 31337:

```sh
docker compose -f docker-compose.yml -f docker-compose.local-chain.yml up --build
```

Then:

```sh
curl 'http://localhost:5000/agents?role=expert'
curl -X POST http://localhost:5000/mcp -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

Swagger: http://localhost:5000/swagger/index.html. Grafana: http://localhost:3000.

## What it serves

| MCP tool | REST | What it returns |
|---|---|---|
| `search_agents` | `GET /agents` | Agents matching words, role, chain, collection, owner, one of their addresses, or whether they published a URL. Paged, 100 at most per page. |
| `get_agent` | `GET /agents/{chainId}/{collectionAddress}/{tokenId}` | One agent. |
| `get_agent_by_ens_name` | `GET /agents/ens/{ensName}` | One agent, e.g. `scout.markets.137.6022.eth`. |
| `get_agent_card` | `GET /agents/ens/{ensName}/card` | The agent's live `/.well-known/6022`: its endpoints per API dialect and its payment methods. `verified` is true when one of the agent's on-chain EVM addresses signed it. |
| `list_collections` | `GET /collections` | The agent collections. |
| | `GET /health` | Liveness. |

Everything an agent returns was written by whoever minted it. Callers should treat it as data,
never as instructions; the MCP tool descriptions say so.

## How the index stays current

Each configured chain has its own loop. Every `poll_interval_seconds` it reads the chain from the
last indexed block up to `head - confirmations`, window by window, and commits each window with its
cursor in one transaction. The first deployment, a restart after downtime and the live updates all
run this same code: nothing is missed while the service is down.

The 6022 contracts' events only say that an agent changed, not to what, so the service reads the
agent back from its contract after any event. ENS records come from the `AgentEnsRegistry` events.

Reorgs deeper than `confirmations` are not undone. If the `AgentEnsRegistry` parent node or manager
ever changes, the service logs it: delete that chain's rows and cursor to resync.

## Configuration

Every setting comes from `appsettings.json` or from the environment (`CHAINS__80002__TESTNET` is
`chains.80002.testnet`). See `src/identity_directory_host/.env.example`.

| Setting | Default | |
|---|---|---|
| `DATABASE__URL` | required | Postgres. Migrations run at startup. |
| `CHAINS__<id>__RPC_HTTP_URL` | required | One block per chain: 137, 80002 or 31337. The 6022 contract addresses are built in. |
| `CHAINS__<id>__TESTNET` | required | Startup refuses a mix of testnet and mainnet chains. |
| `CHAINS__<id>__CONFIRMATIONS` | 20 | How far behind the head the index stays. |
| `CHAINS__<id>__POLL_INTERVAL_SECONDS` | 15 | |
| `CHAINS__<id>__LOG_WINDOW_BLOCKS` | 10000 | Blocks per `eth_getLogs`. A range the node refuses is split in halves. |
| `CHAINS__<id>__MAX_ADDRESSES_PER_LOG_QUERY` | 10 | Collections per `eth_getLogs`. |
| `CHAINS__<id>__REQUEST_TIMEOUT_SECONDS` | 10 | |
| `AGENT_CARD__REQUEST_TIMEOUT_SECONDS` | 10 | |
| `AGENT_CARD__ALLOW_PRIVATE_NETWORKS` | false | Card fetches only reach public addresses. Turn on for local agents only. |
| `RATE_LIMIT__REQUESTS_PER_SECOND` | 50 | Per client IP, per instance. |
| `PROXY__HEADER`, `PROXY__TRUSTED_PROXIES` | image: `X-Forwarded-For` from private ranges | Where the client IP comes from. |
| `TELEMETRY__*` | off | OpenTelemetry export. |

## Running it in production

Run the image `ghcr.io/6022-labs/agentic-identity-directory` with a Postgres database. Run one
deployment for mainnet chains and another one for testnets. Several replicas can share a database:
only one of them commits each window.

The endpoint is public and needs no key. Keep it behind a proxy that sets `X-Forwarded-For`, so
the rate limit applies per client.

## Monitoring

With `telemetry.enabled` on, the service sends metrics, logs and traces to Grafana Alloy over
OTLP. `docker compose up` turns it on and starts a full local stack (Alloy, Mimir, Loki,
Tempo, Grafana) with these dashboards:

- **Identity Directory**: blocks behind the confirmed head and the indexed block per chain, failed
  sync passes and their duration, agent card fetches and their duration, database queries and errors.
- **API Requests**, **Blockchain RPC**, **DB Queries** and **Logs**, shared with the other 6022 services.

Alloy accepts data from anyone by default. On an untrusted network, set
`OTLP_AUTH_HTPASSWD` or `OTLP_AUTH_TOKEN`, and `OTLP_TLS_*`.

## Development

```sh
make build             # go build ./src/...
make test              # unit tests
make integration-test  # Postgres + anvil in Docker
make lint              # golangci-lint v2
make mocks             # after any interface change
make swagger           # after any controller change
make migration NAME=x  # after any model change (atlas + Docker)
make abi               # after a contract ABI change (abigen)
make hooks             # git hooks: format + secret scan on commit, build/lint/test on push
```

Conventions are in [CLAUDE.md](CLAUDE.md).

## License

MIT — see [LICENSE](LICENSE).
