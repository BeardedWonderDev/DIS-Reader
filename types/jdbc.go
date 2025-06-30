package types

type JDBCResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ResultRow is a map representing a row from a JDBC query result.
type ResultRow map[string]interface{}
