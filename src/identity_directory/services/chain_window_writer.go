package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
)

type ChainWindowWriter struct {
	unitOfWorkFactory         interfaces.UnitOfWorkFactoryInterface
	agentCollectionRepository repositories.AgentCollectionRepositoryInterface
	agentRepository           repositories.AgentRepositoryInterface
	agentEnsRecordRepository  repositories.AgentEnsRecordRepositoryInterface
	syncCursorRepository      repositories.SyncCursorRepositoryInterface
}

func NewChainWindowWriter(
	unitOfWorkFactory interfaces.UnitOfWorkFactoryInterface,
	agentCollectionRepository repositories.AgentCollectionRepositoryInterface,
	agentRepository repositories.AgentRepositoryInterface,
	agentEnsRecordRepository repositories.AgentEnsRecordRepositoryInterface,
	syncCursorRepository repositories.SyncCursorRepositoryInterface,
) *ChainWindowWriter {
	return &ChainWindowWriter{
		unitOfWorkFactory:         unitOfWorkFactory,
		agentCollectionRepository: agentCollectionRepository,
		agentRepository:           agentRepository,
		agentEnsRecordRepository:  agentEnsRecordRepository,
		syncCursorRepository:      syncCursorRepository,
	}
}

func (w *ChainWindowWriter) Write(ctx context.Context, chainWindow *models.ChainWindow, expectedCursor *uint64) error {
	unitOfWork, err := w.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return fmt.Errorf("start unit of work: %w", err)
	}
	defer func() { _ = unitOfWork.Rollback() }()
	scopedCtx := unitOfWork.Context()

	for _, collection := range chainWindow.Collections {
		if err := w.agentCollectionRepository.Upsert(scopedCtx, collection); err != nil {
			return fmt.Errorf("upsert collection %s: %w", collection.Address, err)
		}
	}
	for _, agent := range chainWindow.Agents {
		if err := w.agentRepository.Upsert(scopedCtx, agent); err != nil {
			return fmt.Errorf("upsert agent %s/%s: %w", agent.CollectionAddress, agent.TokenId, err)
		}
	}
	// Log order: a later value of the same key wins.
	for _, ensRecord := range chainWindow.EnsRecords {
		if err := w.agentEnsRecordRepository.Set(scopedCtx, ensRecord); err != nil {
			return fmt.Errorf("set ens record %s: %w", ensRecord.Key, err)
		}
	}
	if err := w.syncCursorRepository.Advance(
		scopedCtx,
		chainWindow.ChainId,
		expectedCursor,
		chainWindow.ToBlock,
	); err != nil {
		return err
	}

	return unitOfWork.Commit()
}
