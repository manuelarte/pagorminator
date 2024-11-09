package internal

type PageRequestImpl struct {
	Page          int
	Size          int
	TotalPages    int
	TotalElements int
}

func (p PageRequestImpl) GetOffset() int {
	return (p.Page - 1) * p.Size
}

func (p PageRequestImpl) GetPage() int {
	return p.Page
}

func (p PageRequestImpl) GetSize() int {
	return p.Size
}

func (p PageRequestImpl) GetTotalPages() int {
	return p.TotalPages
}

func (p PageRequestImpl) GetTotalElements() int {
	return p.TotalElements
}

func (p PageRequestImpl) IsUnPaged() bool {
	return p.Size == 0 && p.Page == 0
}
