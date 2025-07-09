# DIS Keystone IAH Table – Master Field Reference

This document outlines the decoded field meanings for the `FILEC.IAH` table within the DIS Keystone system. It is based on reverse engineering of invoice UIs, internal data patterns, and historical Keystone analysis.

## ✅ Decoded Fields

| Column     | Meaning                                                                 |
|------------|-------------------------------------------------------------------------|
| `AHDIV`    | Division Code (e.g., `C` = Coastal Tractor)                             |
| `AHAORS`   | Order Source (`S` = Sales, `A` = POS/Parts Counter)                     |
| `AHAS#`    | Customer or Job Number (e.g., `CASHC`, `C01153`)                         |
| `AHVEND`   | Vendor Code (e.g., `FNH`, `TRI`, `PAW`)                                  |
| `AHPART`   | Part Number or SKU                                                      |
| `AHPDTE`   | Posting Date (YYYYMMDD)                                                 |
| `AHDOC`    | Document/Invoice Number                                                 |
| `AHLINE`   | Line Number (e.g., `PC0070`)                                            |
| `AHFRMT`   | Format Type (`P` = Part line)                                           |
| `AHECDE`   | Tax/Exception Code (`a` = override type; maps to TaxDP column)          |
| `AHDESC`   | Item Description (e.g., `ARM`, `SOFTWARE`)                              |
| `AHCLSS`   | Classification (`Y` = Inventory Tracked)                                |
| `AHPCDE`   | Price Code (`L` = List, `C` = Contract)                                 |
| `AHCOST`   | Cost to dealership (2 decimal places)                                   |
| `AHLOC`    | Bin/Storage Location (e.g., `TIL01A`, `SOFTWARE`)                       |
| `AHQTY`    | Quantity                                                                |
| `AHPRCE`   | Extended Sale Price                                                     |
| `AHTAX`    | Tax Code (e.g., `C\`I`)                                                |
| `AHDRTE`   | Discount Rate (decimal format)                                          |
| `AHMEMO`   | Notes or Comments (appears in UI Notes section)                         |
| `AHRBY`    | Rebill Flag (`1` indicates chargeable)                                  |
| `AHPOST`   | Posting Status Flag (likely `Y` when posted)                            |
| `AHSTOK`   | Stocked Item Flag                                                       |
| `AHWARR`   | Warranty Flag                                                           |
| `AHSWAR`   | Serialized Warranty Indicator                                           |
| `AHSERL`   | Serial Number (blank unless item is serialized)                         |
| `AHDIV1`   | Repeat of Division Code (audit or cross-divisional logic)               |
| `AHVND1`   | Repeat of Vendor Code                                                   |
| `AHPDT1`   | Repeat of Posting Date                                                  |
| `AHDIV2`   | Possibly alternate division for split invoice                           |
| `AHPDT2`   | Possibly original invoice date or alt posting date                      |

## ⚠️ Partially Decoded / Inferred Fields

| Column     | Meaning                                                                 |
|------------|-------------------------------------------------------------------------|
| `AHDISC`   | Discount Indicator (unused or flag)                                      |
| `AHSPRT`   | Possibly extended description or unused                                  |
| `AHPROC`   | Possibly internal process flag                                           |
| `AHDMND`   | Possibly demand source or request flag                                   |
| `AHSEA`    | Possibly seasonal or category indicator                                  |
| `AHFTAX`   | Possibly Freight Taxable Indicator                                       |
| `AHWORI`   | Unknown – possibly write-in or origin tracking flag                      |
| `AHFIL1`   | User-defined or legacy switch (unused in sample data)                    |
| `AHFIL2`   | Secondary user-defined or reserved field                                 |

## 🧪 Notes & Observations

- `AHTAX` and `AHECDE` together decode the TaxDP column from POS screens.
- `AHLINE` aligns with the POS line number (e.g., `0070`, `0080`).
- Pricing margins (AHCOST vs AHPRCE) match internal markup patterns.
- Some duplicated fields (AHDIV1/2, AHPDT1/2, AHVND1) exist for accounting traceability.

## 🔧 Remaining Unknowns

These fields require further investigation, especially via service invoice data or serialized unit handling:
- `AHWORI`
- `AHFIL1`, `AHFIL2`
- `AHSEA`, `AHDMND`

---

**Next Step:** Convert this schema into a Go struct with type-safe annotations and potential enum values where applicable.

