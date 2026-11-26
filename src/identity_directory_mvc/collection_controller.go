package identity_directory_mvc

import (
	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/common/mvc"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const CollectionsEndpoint = "/collections"

type CollectionController struct {
	errorWriter     *mvc.ErrorWriter
	listCollections *use_cases.ListCollections
}

func NewCollectionController(
	errorWriter *mvc.ErrorWriter,
	listCollections *use_cases.ListCollections,
) *CollectionController {
	return &CollectionController{errorWriter: errorWriter, listCollections: listCollections}
}

func (c *CollectionController) RegisterRoutes(app fiber.Router) {
	app.Get(CollectionsEndpoint, c.ListCollections)
}

// @Summary List agent collections
// @Tags Collections
// @Produce json
// @Param chainId query int false "EVM chain id"
// @Param page query int false "page, from 1"
// @Param elementsPerPage query int false "1 to 100, 10 by default"
// @Success 200 {object} responses.ListCollectionsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 400 {object} errors.BadRequestError "invalid_query"
// @Failure 500 {object} errors.InternalError
// @Router /collections [get]
func (c *CollectionController) ListCollections(ctx fiber.Ctx) error {
	request := &requests.ListCollectionsRequest{}
	if err := ctx.Bind().Query(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_query", "query parameters are not valid"))
	}

	response, err := c.listCollections.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}
