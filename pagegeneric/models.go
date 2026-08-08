package pagegeneric

const (
	// NoNextPage is a constant that represents that there is no next page.
	NoNextPage = NextPossible(false)
	// NoTotalElements is a constant that represents that the total elements are not set,
	// so we can't know whether there is a next page.
	NoTotalElements = NextPossible(false)
	// PreviousCursorValuesNotSet is a constant that represents that the previous cursor values
	// are not set, so we can't know whether there is a next page.
	PreviousCursorValuesNotSet = NextPossible(false)
)

type (
	// NextPossible represents whether the next page is available.
	NextPossible bool
)
