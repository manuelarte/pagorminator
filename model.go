package pagorminator

import (
	"errors"
	"github.com/manuelarte/pagorminator/internal"
)

var (
	ErrPageCantBeNegative = errors.New("page number can't be negative")
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	ErrSizeNotAllowed     = errors.New("size is not allowed")
)

var _ PageRequest = internal.PageRequestImpl{}

// PageRequest Struct that contains the pagination information
type PageRequest interface {
	GetPage() int
	GetSize() int
	GetOffset() int
	GetTotalPages() int
	GetTotalElements() int
	IsUnPaged() bool
}

// PageRequestOf Creates a PageRequest with the page and size values
func PageRequestOf(page, size int) (PageRequest, error) {
	if page < 0 {
		return nil, ErrPageCantBeNegative
	}
	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}
	if page > 0 && size == 0 {
		return nil, ErrSizeNotAllowed
	}
	return &internal.PageRequestImpl{Page: page, Size: size}, nil
}

// UnPaged Create an unpaged request (no pagination is applied)
func UnPaged() PageRequest {
	return &internal.PageRequestImpl{Page: 0, Size: 0}
}
