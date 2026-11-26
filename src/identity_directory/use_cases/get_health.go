package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

type GetHealth struct{}

func NewGetHealth() *GetHealth {
	return &GetHealth{}
}

func (u *GetHealth) Execute(_ context.Context) (*responses.GetHealthResponse, error) {
	return &responses.GetHealthResponse{Status: "ok"}, nil
}
