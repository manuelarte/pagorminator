package internal

type PaginationResponse interface {
	SetTotalElements(totalElements int64) error
}
