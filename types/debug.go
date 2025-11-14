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

type DebugSearchOutputMode string

const (
	DebugSearchOutputSQLite DebugSearchOutputMode = "sqlite"
	DebugSearchOutputCSV    DebugSearchOutputMode = "csv"
)

type DebugSearchOptions struct {
	OutputMode DebugSearchOutputMode
	OutputPath string
}
