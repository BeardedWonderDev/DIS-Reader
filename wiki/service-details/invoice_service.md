# Invoice Service Reference

This document details the `InvoiceService` exposed by DIS Reader, including data fields, supported filters/sorts, and example usage for integrating invoice lookups into Go applications.

---

## 1. Overview

`InvoiceService` targets `FILEC.IAH` (invoice items) and `FILEC.CUSINV` (whole goods invoices). It lets you:

- Fetch a specific invoice line via document + line identifiers.
- List invoice items with filters on customer, part, posting date, etc.
- Fetch a whole goods invoice line (division/serial sales).
- List whole goods invoices with salesperson, price, and invoice date filters.

Obtain the service through the main DIS reader instance:

```go
invoiceSvc := disSvc.InvoiceService()
```

---

## 2. Data Model (`types.InvoiceItemSpec`)

The spec mirrors all decoded columns documented in `wiki/invoice_items_table.md`.

Selected fields:

| Field | Description |
|-------|-------------|
| `DocumentNumber`, `LineID` | Primary identifiers (`AHDOC`, `AHLINE`). |
| `Division`, `OrderSource`, `CustomerNumber`, `VendorCode`, `PartNumber` | Core dimensions. |
| `PostingDate`, `FormatType`, `ExceptionCode`, `Classification` | Operational metadata. |
| `Description`, `PriceCode`, `Quantity`, `Price`, `Cost`, `Location` | Transaction detail. |
| `TaxCode`, `DiscountRate`, `DiscountIndicator`, `Rebill`, `Posted` | Pricing & tax behavior. |
| `Stocked`, `Warranty`, `SerializedWarranty`, `SerialNumber` | Inventory coverage. |
| `AuditDivision`, `AuditPostingDate`, `AlternateDivision`, etc. | Duplicated audit columns. |
| `ExtendedDescription`, `ProcessCode`, `DemandCode`, `SeasonalCode`, `FreightTaxCode`, `OriginFlag` | Extended semantics (see wiki entry for meaning). |
| `UserField1`, `UserField2` | Custom/reserved fields. |

Optional data is wrapped in pointers so JSON output omits missing values.

## 2.1 Data Model (`types.WholeGoodsInvoiceSpec`)

Whole goods invoices originate from `FILEC.CUSINV`. The exported spec exposes every column with the cleanest label we have today:

| Field | Column | Notes |
|-------|--------|-------|
| `Division` | `DS8LA` | Division code. |
| `LegacyDS9YA` | `DS9YA` | Currently undocumented numeric flag, exposed for parity. |
| `LineItemNumber` | `DS1XQA` | Whole goods line number (converted to `int`). |
| `InvoiceNumber` | `DSHUA` | Trimmed and sanitized (`" #*"` suffixes removed, same as unit service). |
| `LegacyDSYGA` | `DSYGA` | One-character status/flag (unknown meaning). |
| `InvoiceDate` | `DSHVA` | Native DATE column. |
| `LineItemPrice` | `DSHWA` | 9,2 DECIMAL expressed as `float64`. |
| `LineItemDescription` | `DSHXA` | 40-char description. |
| `LegacyDSWVA` | `DSWVA` | Additional numeric flag (unknown). |
| `SoldBy` | `DS22UA` | Salesperson code / initials. |

Any unknown/legacy values stay nullable so callers can ignore them until business meaning is confirmed.

---

## 3. Methods

### 3.1 `GetItem(ctx context.Context, documentNumber, lineID string) (*types.InvoiceItemSpec, error)`

Returns a single invoice line. Example:

```go
item, err := invoiceSvc.GetItem(ctx, "INV12345", "PC0070")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%s qty %.0f sold for %.2f\n", item.PartNumber, item.Quantity, item.Price)
```

### 3.2 `ListItems(ctx context.Context, lp types.ListParams) ([]*types.InvoiceItemSpec, error)`

Build `ListParams` with the invoice parser to ensure safe SQL mapping.

```go
parser := types.InvoiceListParamParser{}

sortBy, _ := types.ParseSortBy("postingDate", parser)
postingDate, _ := time.Parse("2006-01-02", "2024-01-01")

dateFilter, _ := types.ParseFilter(types.Filter{
	Field:    "postingDate",
	Operator: ">=",
	Value:    postingDate,
}, parser)

customerFilter, _ := types.ParseFilter(types.Filter{
	Field:    "customer",
	Operator: "=",
	Value:    "C01153",
}, parser)

lp := types.ListParams{
	Limit:   200,
	SortBy:  sortBy,
	Filters: []types.Filter{dateFilter, customerFilter},
}

items, err := invoiceSvc.ListItems(ctx, lp)
```

### 3.3 `GetWholeGoodsInvoice(ctx context.Context, invoiceNumber, lineItemNumber string) (*types.WholeGoodsInvoiceSpec, error)`

Mirrors `GetItem`, but queries `FILEC.CUSINV` using invoice + line identifiers (line numbers are numeric strings).

```go
wg, err := invoiceSvc.GetWholeGoodsInvoice(ctx, "WG12345 #*", "10")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%s sold %s for %.2f\n", wg.SoldBy, wg.LineItemDescription, wg.LineItemPrice)
```

### 3.4 `ListWholeGoodsInvoices(ctx context.Context, lp types.ListParams) ([]*types.WholeGoodsInvoiceSpec, error)`

Uses the dedicated parser shown below to keep filters aligned with the legacy schema.

```go
wgParser := types.WholeGoodsInvoiceListParamParser{}

soldByFilter, _ := types.ParseFilter(types.Filter{
	Field:    "soldBy",
	Operator: "=",
	Value:    "JDAVIS",
}, wgParser)

after, _ := time.Parse("2006-01-02", "2024-07-01")
dateFilter, _ := types.ParseFilter(types.Filter{
	Field:    "invoiceDate",
	Operator: ">=",
	Value:    after,
}, wgParser)

sortBy, _ := types.ParseSortBy("invoiceDate", wgParser)

lp := types.ListParams{
	Limit:   100,
	SortBy:  sortBy,
	Filters: []types.Filter{soldByFilter, dateFilter},
}

rows, err := invoiceSvc.ListWholeGoodsInvoices(ctx, lp)
```

---

## 4. Sorting & Filtering Reference

### 4.1 Sortable Fields

Provided by `InvoiceListParamParser.GetAllowedSortByColumns()`:

| Logical Field | Column | Notes |
|---------------|--------|-------|
| `document` | `AHDOC` | Invoice number. |
| `line` | `AHLINE` | Line ID (e.g., `PC0070`). |
| `postingDate` | `AHPDTE` | AS/400 numeric date converted before sorting. |
| `customer` | `AHAS#` | Customer/job number. |
| `vendor` | `AHVEND` | Vendor code. |
| `part` | `AHPART` | Part/SKU. |
| `division` | `AHDIV` | Division letter. |
| `cost` | `AHCOST` | Cost. |
| `price` | `AHPRCE` | Extended sale price. |

Use `types.ParseSortBy` to validate names; fallback is parser default (`postingDate`).

### 4.2 Allowed Filters

`InvoiceListParamParser.GetAllowedFilterColumns()` maps:

| Field | Column | Value Type |
|-------|--------|------------|
| `document` | `AHDOC` | string |
| `line` | `AHLINE` | string |
| `postingDate` | `AHPDTE` | `time.Time` |
| `customer` | `AHAS#` | string |
| `vendor` | `AHVEND` | string |
| `part` | `AHPART` | string |
| `division` | `AHDIV` | string |
| `orderSource` | `AHAORS` | string |
| `formatType` | `AHFRMT` | string |
| `class` | `AHCLSS` | string |
| `priceCode` | `AHPCDE` | string |

Create filters via `types.ParseFilter` to guarantee proper quoting/`IN` formatting.

Special cases:
- For `postingDate`, supply `time.Time`.
- For numeric comparisons (cost/price/discountRate), use floats and convert with `types.InvoiceListParamParser.ParseValue`.

### 4.3 Search Query

If `ListParams.Query` is non-empty, `Repository.ListInvoiceItems` applies a wildcard search across `AHDOC`, `AHPART`, and `AHAS#`, using proper escaping for `%` and `_`.

### 4.4 Whole Goods Sortable Fields

Provided by `WholeGoodsInvoiceListParamParser.GetAllowedSortByColumns()`:

| Field | Column | Notes |
|-------|--------|-------|
| `invoiceNumber` | `DSHUA` | Whole goods invoice number (raw). |
| `lineItem` | `DS1XQA` | Numeric line item identifier. |
| `invoiceDate` | `DSHVA` | Proper DATE column. |
| `division` | `DS8LA` | Division/location code. |
| `price` | `DSHWA` | DECIMAL(9,2). |
| `soldBy` | `DS22UA` | Salesperson / closer. |
| `ds9ya`, `dsyga`, `dswva` | `DS9YA`, `DSYGA`, `DSWVA` | Raw legacy columns for advanced filters. |

### 4.5 Whole Goods Filters

`WholeGoodsInvoiceListParamParser.GetAllowedFilterColumns()` covers:

| Field | Column | Value Type |
|-------|--------|------------|
| `invoiceNumber` | `DSHUA` | string |
| `lineItem` | `DS1XQA` | integer |
| `invoiceDate` | `DSHVA` | `time.Time` |
| `division` | `DS8LA` | string |
| `price` | `DSHWA` | float |
| `description` | `DSHXA` | string |
| `soldBy` | `DS22UA` | string |
| `ds9ya` | `DS9YA` | integer |
| `dsyga` | `DSYGA` | string |
| `dswva` | `DSWVA` | integer |

Use the parser when calling `types.ParseFilter`, so numbers/dates convert before SQL rendering.

Whole goods `ListParams.Query` fans out across `DSHUA`, `DSHXA`, and `DS22UA` (invoice number, description, salesperson) with the same escaping rules as the standard invoice search.

---

## 5. Usage Scenarios

| Scenario | Approach |
|----------|----------|
| Fetching all lines for a document | Filter on `document = 'INV123'`, optionally `Limit` large enough to cover lines. |
| Customer history | Filter `customer`, optionally `postingDate >= lastYear`. |
| Vendor analysis | Filter `vendor` + date range, sort by `postingDate DESC`. |
| High-value line items | Filter `price >= 10000`, sort by `price DESC`. |
| Whole goods sales leaderboard | Use `ListWholeGoodsInvoices`, filter by `invoiceDate`, group by `SoldBy` in application code. |

Combine filters and sorts to feed reporting APIs or dashboards.

---

## 6. Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| `expected time.Time for postingDate filter` | Value left as string | Parse to `time.Time` before calling `ParseFilter`. |
| Special characters in `Query` yield SQL errors | Manual `%` or `_` not escaped | Use `ListParams.Query`; repository handles escaping automatically. |
| Duplicate rows | Table includes duplicates per search criteria | Add extra filters (line, division) or dedupe in application logic. |
| Performance concerns | Wide scans on `AHPART`/`AHDOC` | Reduce `Limit`, add targeted filters, or chunk by posting date. |

---

Keep this reference handy whenever you add new invoice-centric endpoints, background jobs, or analytics. Update the tables above if the underlying schema or parser capabilities evolve.***
