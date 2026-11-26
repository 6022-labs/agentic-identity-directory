package responses

// PaginateResponse is embedded by every paginated list response.
type PaginateResponse struct {
	Page            uint `json:"page"`
	ElementsPerPage uint `json:"elementsPerPage"`
	TotalElements   uint `json:"totalElements"`
}
