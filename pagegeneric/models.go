package pagegeneric

const (
	// NoNextPage is a constant that represents that there is no next page.
	NoNextPage = PrevNextPossible(false)
	// NoPrevPage is a constant that represents that there is no previous page.
	NoPrevPage = PrevNextPossible(false)
	// NoTotalElements is a constant that represents that the total elements are not set,
	// so we can't know whether there is a next page.
	NoTotalElements = PrevNextPossible(false)
	// PreviousCursorValuesNotSet is a constant that represents that the previous cursor values
	// are not set, so we can't know whether there is a next page.
	PreviousCursorValuesNotSet = PrevNextPossible(false)
)

type (
	// PrevNextPossible represents whether the previous and next page are available.
	PrevNextPossible bool
)
