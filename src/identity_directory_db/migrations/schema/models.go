// Package schema lists the gorm models the migrations cover; the loader and the migration tests both read it.
package schema

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

func Models() []any {
	return []any{
		&models.AgentCollection{},
		&models.Agent{},
		&models.AgentAddress{},
		&models.AgentEnsRecord{},
		&models.SyncCursor{},
	}
}
