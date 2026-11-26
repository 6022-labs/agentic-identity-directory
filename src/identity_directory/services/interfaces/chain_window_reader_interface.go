package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type ChainWindowReaderInterface interface {
	// Read collects what the block range changed; the state is read at headBlock.
	Read(ctx context.Context, chainId, fromBlock, toBlock, headBlock uint64) (*models.ChainWindow, error)
}
