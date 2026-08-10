package pagegeneric

const (
	// PagorminatorClause is the key used to store the pagination clause in the GORM statement context.
	PagorminatorClause = "pagorminator:clause"
	// PagorminatorCursorWhereSQL stores the cursor pagination where SQL to strip from count queries.
	PagorminatorCursorWhereSQL = "pagorminator:cursor:where:sql"
	// PagorminatorCursorWhereVars stores the cursor pagination where vars to strip from count queries.
	PagorminatorCursorWhereVars = "pagorminator:cursor:where:vars"
)
