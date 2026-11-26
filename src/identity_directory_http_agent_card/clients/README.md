# identity_directory_http_agent_card clients

`oapi_client.gen.go` is the generated HTTP client this adapter uses to read an agent's `/.well-known/6022` card.
It is produced by `oapi-codegen` (pinned `v2.5.1`, `types,client`) from `agent_node_gateway.v3.yaml`, filtered to
the `6022` tag: the directory never calls the rest of the gateway.

`agent_node_gateway.v3.yaml` is `src/gateway_host/docs/swagger.v3.yaml` from `agentic-agent-node`, vendored so
generation never depends on that repo being checked out. Refresh it whenever the 6022 card changes.

Do not edit `oapi_client.gen.go` by hand. To regenerate:

```sh
make agent-node-client                                        # from the vendored spec
AGENT_NODE_SPEC=../agentic-agent-node/src/gateway_host/docs/swagger.v3.yaml \
  make agent-node-client                                      # vendor the upstream spec first
```

`../services/agent_card_requester.go` wraps the generated operation, maps the card to the domain model (keeping
the raw body its signatures cover) and implements `AgentCardRequesterInterface`.
