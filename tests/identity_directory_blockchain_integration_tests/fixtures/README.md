# Contract fixtures

`artifacts/*.json` are trimmed hardhat artifacts (`{contractName, abi, bytecode}`) the integration tests deploy onto a throwaway anvil chain. The abigen bindings carry no bytecode, so they cannot be deployed.

- `AgentCollectionsManager`, `AgentCollectionsDescriptor`, `AgentCollectionV1`, `AgentCollectionCreatorV1`: compiled from `agentic-smart-contracts`.
- `AgentEnsRegistry`: taken from `agentic-smart-contracts/deployments/amoy/AgentEnsRegistry.json`.

## Refreshing

Only needed when a contract changes in a way the tests rely on. Keep only the three fields:

```sh
python3 -c "
import json, sys
a = json.load(open(sys.argv[1]))
json.dump({'contractName': sys.argv[3], 'abi': a['abi'], 'bytecode': a['bytecode']}, open(sys.argv[2], 'w'), indent=2)
" <artifact or deployment json> artifacts/<Name>.json <Name>
```
