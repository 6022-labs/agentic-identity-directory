package requests

import "math"

const (
	defaultPage            = uint(1)
	defaultElementsPerPage = uint(10)
	maxElementsPerPage     = uint(100)
	// maxPage keeps the offset within an int, so a huge page reads as empty instead of a negative OFFSET.
	maxPage = uint(math.MaxInt32)
)

// PaginateRequest pages a list; Sanitize defaults and clamps every field rather than rejecting it.
type PaginateRequest struct {
	Page            *uint `query:"page"            json:"-"`
	ElementsPerPage *uint `query:"elementsPerPage" json:"-"`
}

func (r *PaginateRequest) Sanitize() {
	if r.Page == nil || *r.Page < 1 {
		page := defaultPage
		r.Page = &page
	}
	if *r.Page > maxPage {
		page := maxPage
		r.Page = &page
	}
	if r.ElementsPerPage == nil || *r.ElementsPerPage < 1 {
		elementsPerPage := defaultElementsPerPage
		r.ElementsPerPage = &elementsPerPage
	}
	if *r.ElementsPerPage > maxElementsPerPage {
		elementsPerPage := maxElementsPerPage
		r.ElementsPerPage = &elementsPerPage
	}
}

// Offset is the number of elements the pages before this one hold; call after Sanitize.
func (r *PaginateRequest) Offset() int {
	return int((*r.Page - 1) * *r.ElementsPerPage)
}
