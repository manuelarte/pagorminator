package pagegeneric

//nolint:gochecknoglobals // global variables to specify why there is no next page.
var (
	NoNextPage                 = NextPossible(false)
	NoTotalElements            = NextPossible(false)
	PreviousCursorValuesNotSet = NextPossible(false)
)

type (
	NextPossible bool
)
