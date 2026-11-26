package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type ChainWindowWriterInterface interface {
	// Write stores the window and moves the cursor to its last block, atomically.
	Write(ctx context.Context, chainWindow *models.ChainWindow, expectedCursor *uint64) error
}
