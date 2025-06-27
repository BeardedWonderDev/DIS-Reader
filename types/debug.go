package types

type ProgressStatus struct {
	RunID            string
	TotalQueries     int
	CompletedQueries int
	PercentComplete  float64
}

type TableEvent struct {
	TableName   string
	EventType   string // "table_created" or "row_inserted"
	RowCount    int
	ColumnCount int
	SampleRow   map[string]interface{}
}
