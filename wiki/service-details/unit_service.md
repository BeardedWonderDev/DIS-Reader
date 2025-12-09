# Unit Service Reference

This document describes the `UnitService` provided by DIS Reader: available methods, DTO fields, sorting/filtering options, and sample code snippets for integration.

---

## 1. Overview

`UnitService` wraps DIS table `FILEC.DMUNITM1` and exposes helpers for:

- Fetching a single unit by unit number.
- Listing units with powerful search, sort, and filter combinations.

Access the service via:

```go
unitSvc := disSvc.UnitService()
```

---

## 2. Data Model (`types.UnitSpec`)

`UnitSpec` fields include identifying info, product metadata, financials, and customer/sale details. Key fields (non-exhaustive):

| Field | Description |
|-------|-------------|
| `UnitID` | Unique unit number. |
| `Year`, `Make`, `Model`, `Description` | Basic product data. |
| `Serial`, `Engine`, `HorsePower`, `EngineHours`, `Color` | Equipment attributes (pointers when missing). |
| `Location`, `Account`, `Status`, `ProductCode` | Operational metadata. |
| `CreatedAt`, `SoldAt` | AS/400 dates converted to `time.Time`. |
| `InvoiceNumber`, `SoldTo`, `SoldBy`, `SoldByName` | Sales attribution. |
| `RevenueAmount`, `Cost`, `SuggestedList`, `DealerList` | Pricing. |
| `WarrantyCode`, `TradedOnUnit`, `GLAccount` | Accounting details. |
| `FlooringDueDate`, `RentalStartDate`, `RentalRevenue`, etc. | Flooring & rental portfolio data. |

All fields carry JSON tags; pointer fields omit when empty.

---

## 3. Methods

### 3.1 `GetByUnitNumber(ctx context.Context, unitNum string) (*types.UnitSpec, error)`

Queries `FILEC.DMUNITM1` for the exact unit number.

```go
unit, err := unitSvc.GetByUnitNumber(ctx, "123456")
if err != nil {
	log.Fatal(err)
}
fmt.Println(unit.Description, unit.Location)
```

### 3.2 `ListUnits(ctx context.Context, lp types.ListParams) ([]*types.UnitSpec, error)`

Returns a slice of `UnitSpec` after applying full-text search, filters, sort, and limit. See the next section for building `ListParams`.

```go
parser := types.UnitListParamParser{}

sortBy, _ := types.ParseSortBy("", parser)        // defaults to "UNIT"
filters := []types.Filter{}

lp := types.ListParams{
	Limit:   50,
	SortBy:  sortBy,
	Filters: filters,
	Query:   "SKID STEER", // optional
}

units, err := unitSvc.ListUnits(ctx, lp)
```

---

## 4. Sorting & Filtering Reference

`types.UnitListParamParser` governs supported fields.

### 4.1 Sortable Fields

| Logical Field | Column | Notes |
|---------------|--------|-------|
| `unit` | `UNIT` | Default sort. |
| `year` | `YEAR` | Numeric as string. |
| `make`, `model`, `location` | `MAKE`, `MODEL`, `LOCATION` | Text columns. |
| `indate` | `INDATE` | Converts AS/400 date before ordering. |
| `soldat` | `SOLDAT` | Date. |
| `soldto` | `SOLDTO` | Customer account. |
| `amount` | `AMOUNT` | Revenue numeric. |
| `cost` | `COST` | Cost numeric. |
| `soldby` | `SOLDBY` | Salesperson. |

`types.ParseSortBy` maps logical names to actual columns; fallback is parser default.

### 4.2 Filters

Allowed logical fields map to:

| Field | Column | Value Type |
|-------|--------|------------|
| `unit` | `UNIT` | string |
| `year` | `YEAR` | int |
| `make` | `MAKE` | string |
| `model` | `MODEL` | string |
| `productCode` | `PRODCT` | string |
| `location` | `LOCATION` | string |
| `indate` | `INDATE` | `time.Time` |
| `soldat` | `SOLDAT` | `time.Time` |
| `soldto` | `SOLDTO` | string |
| `amount` | `AMOUNT` | float64 |
| `cost` | `COST` | float64 |
| `soldby` | `SOLDBY` | string |

Use `types.ParseFilter` to validate/escape values:

```go
parser := types.UnitListParamParser{}
filter, err := types.ParseFilter(types.Filter{
	Field:    "soldto",
	Operator: "=",
	Value:    "C01153",
}, parser)

lp.Filters = append(lp.Filters, filter)
```

- `Operator` can be `=`, `<>`, `LIKE`, `IN`, `>`, `<`, etc.
- For `IN`, set `Value` to `[]string` and the helper will format `('A','B')`.
- Date filters require `time.Time`. Use `time.Parse("2006-01-02", "...")`.

### 4.3 Full-Text Search (`ListParams.Query`)

When `Query` is non-empty, `UnitRepository` searches `UNIT`, `MAKE`, and `MODEL` columns with `LIKE '%term%'`.

---

## 5. Usage Patterns

- **Pagination**: Combine `Limit` with `SortBy` + filters. `types.ListParams.UseCursorPagination()` is available if you implement cursor-based flows.
- **Reporting APIs**: Marshal `UnitSpec` directly to JSON; optional fields (pointers) drop cleanly when nil.
- **Error Handling**: Repository errors include raw SQL issues (bad filters, connection errors). Wrap with context in calling layers if needed.

---

## 6. Troubleshooting Tips

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Empty results despite known unit | Unit number lacks padding or formatting | Trim/uppercase values before calls (`strings.TrimSpace`). |
| `expected time.Time for indate filter` | Filter value not parsed to `time.Time` | Parse string to `time.Time` first. |
| SQL errors when using `IN` | Provided a raw string `"('A','B')"` | Prefer `[]string` and let helper format it. |
| Slow queries | Large `Limit` without indexes | Narrow filters, add `Query`, or limit to specific columns. |

---

Use this reference to build precise inventory endpoints or analytics pipelines without diving into the internal repository code. Update the table above if new columns or behaviors are added to `UnitService`.***
