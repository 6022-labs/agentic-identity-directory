#!/usr/bin/env sh
set -eu

# Regenerates the agent-node gateway client from the vendored OpenAPI v3 spec.
# Refresh the spec first with AGENT_NODE_SPEC=<agentic-agent-node>/src/gateway_host/docs/swagger.v3.yaml.

CLIENTS_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
SPEC_PATH="${CLIENTS_DIR}/agent_node_gateway.v3.yaml"
OUTPUT_FILE="${CLIENTS_DIR}/oapi_client.gen.go"
SANITIZED_SPEC="$(mktemp -t agent-node-spec-XXXXXX).yaml"
trap 'rm -f "${SANITIZED_SPEC}"' EXIT

OAPI_CODEGEN_VERSION="v2.5.1"

if [ -n "${AGENT_NODE_SPEC:-}" ]; then
	echo "[0/2] Vendoring ${AGENT_NODE_SPEC}"
	cp "${AGENT_NODE_SPEC}" "${SPEC_PATH}"
fi

# swag v2 emits the Swagger-2.0 `type: file` for multipart uploads, which
# oapi-codegen rejects on a 3.1 spec. Rewrite it to the 3.x binary-string form.
echo "[1/2] Sanitizing multipart file types for oapi-codegen"
perl -0pe 's/(\s*)type: file/${1}type: string${1}format: binary/g' \
	<"${SPEC_PATH}" >"${SANITIZED_SPEC}"

# Only the 6022 discovery card: the directory never calls the rest of the gateway.
echo "[2/2] Generating agent-node client with oapi-codegen ${OAPI_CODEGEN_VERSION}"
(
	cd "${CLIENTS_DIR}"
	go run "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}" \
		-generate types,client \
		-include-tags 6022 \
		-package clients \
		-o "${OUTPUT_FILE}" \
		"${SANITIZED_SPEC}"
)

echo "Done"
