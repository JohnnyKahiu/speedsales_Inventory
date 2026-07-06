# Inventory Microservice — Memory & CPU Optimization Analysis

**Date:** 2026-06-23  
**Scope:** `/home/johnny/projects/speed_sales/Inventory`  
**Stack:** Go 1.25, pgx v5 connection pool, Kafka consumer, gRPC server, gorilla/mux HTTP

---

## Executive Summary

The service has three systemic problems that compound each other:

1. A global in-memory product cache that is fully re-serialized to disk on *every single write*, including per-sale-item writes in the Kafka consumer.
2. O(n) and O(n²) map scans in hot search paths where direct map lookups are possible.
3. Debug `fmt.Println` / `fmt.Printf` calls left active throughout hot paths.

These combine to make the service significantly more I/O and CPU bound than it needs to be, and introduce a data race between the HTTP handlers and the Kafka goroutine.

---

## Critical Issues

### 1. Full Cache Re-Serialization on Every Write  
**Files:** [pkg/products/storage.go](pkg/products/storage.go) · Lines 358–376, 229–283, 300–319, 331–342

`Pickle()` encodes the *entire* `ProdDB` (all products, codes, VATs) into a gob binary file and writes it to disk. Every function that mutates the cache calls `Pickle()` immediately after:

- `AddProduct()` → 1 `Pickle()` call
- `UpdateLinks()` → 1 `Pickle()` call
- `BulkUpdateLinks()` → **N `Pickle()` calls** (one per code, inside the loop)
- `CacheBal()` → 1 `Pickle()` call — called per item per sale order in the Kafka consumer

For a sales order with 20 items, `ProcessOrder` in [internal/broker/saleOrders.go](internal/broker/saleOrders.go) calls `SaveBal()` once per item (lines 101, 133), which calls `CacheBal()` → `Pickle()`. That means **20 full gob-encode + disk-write cycles per order**, each serializing potentially thousands of products.

**Fix:**
```go
// BulkUpdateLinks — batch all mutations then pickle once
func (arg *ProdDB) BulkUpdateLinks(cts []CodeTranslator) error {
    if cts == nil {
        return nil
    }
    arg.mx.Lock()
    for _, ct := range cts {
        if arg.Codes == nil {
            arg.Codes = make(map[string]CodeTranslator)
        }
        arg.Codes[ct.LinkCode] = ct
    }
    arg.mx.Unlock()
    return arg.Pickle()  // single pickle after all mutations
}
```

For `CacheBal` / `SaveBal`, defer the Pickle call and batch updates from the Kafka consumer:
```go
// In ProcessOrder: accumulate all balance changes, then call Pickle once after the loop
```

---

### 2. Data Race on `ProdMaster.ProductDB`  
**Files:** [pkg/products/storage.go](pkg/products/storage.go) · Line 26, 331–342  
**Files:** [pkg/balances/balance.go](pkg/balances/balance.go) · Lines 222–243

`ProdDB.mx` (a `sync.RWMutex`) is only acquired inside `Pickle()`. All reads and writes to `ProductDB` — from HTTP handlers (`GetByCode`, `SearchDescription`) and from the Kafka goroutine (`SaveBal`) — are **unprotected**. This is a data race that will corrupt the map under concurrent load.

**Fix:** Wrap all reads with `arg.mx.RLock()` / `arg.mx.RUnlock()` and all writes with `arg.mx.Lock()` / `arg.mx.Unlock()`:
```go
func (arg *ProdDB) GetProduct(code string) (StockMaster, bool) {
    arg.mx.RLock()
    defer arg.mx.RUnlock()
    v, ok := arg.ProductDB[code]
    return v, ok
}
```

---

### 3. `GetMultiCodes` — O(n²) Nested Loop  
**File:** [pkg/products/master.go](pkg/products/master.go) · Lines 189–205

```go
// Current — O(keys × products)
for _, val := range ProdMaster.ProductDB {
    for _, code := range keys {
        if val.ItemCode == code {
```

This iterates over every product for every key. The `ProductDB` map is already keyed by `item_code`, so each lookup should be O(1):

```go
func GetMultiCodes(keys []string) ([]StockMaster, error) {
    vals := make([]StockMaster, 0, len(keys))
    for _, code := range keys {
        if v, ok := ProdMaster.ProductDB[code]; ok {
            vals = append(vals, v)
        }
    }
    return vals, nil
}
```

`GetMultiCodes` is called from `Fetch()` and `Details()` in [pkg/products/location.go](pkg/products/location.go) (lines 193, 231) every time a location's stock list is fetched. This makes every location detail request O(n × products).

---

### 4. Full Map Scans on Every Search Request  
**File:** [pkg/products/master.go](pkg/products/master.go) · Lines 211–275

Both `SearchDescription` and `SearchByCategory` iterate over the entire `ProductDB` map on every request. There is no secondary index.

```go
// SearchDescription — O(n) on every keystroke from the UI
for _, val := range ProdMaster.ProductDB {  // iterates ALL products
    ...
    if strings.Contains(strings.ToLower(val.ItemName), ...) {
```

For a catalog of 10,000+ products, each search request scans everything.

**Options (ascending complexity):**
- **Quick win:** Build a sorted `[]string` name index at load time and use `strings.Contains` only on that slice (avoids loading full `StockMaster` structs during iteration).
- **Better:** Maintain an inverted word-to-codes index updated on product add/update.
- **Best for scale:** Use PostgreSQL full-text search (`tsvector`) and remove the in-memory text scan entirely.

---

### 5. `GetSaleLoc` — Database Query Per Sale Item  
**File:** [internal/broker/saleOrders.go](internal/broker/saleOrders.go) · Lines 61–71  
**File:** [pkg/products/location.go](pkg/products/location.go) · Lines 103–134

`ProcessOrder` calls `GetSaleLoc()` for every item in every sale order. This hits the `stock_locations` table with a parameterized query for each item. For an order with 15 items that is 15 sequential DB round-trips before any balance is written.

Stock locations are a near-static lookup table. Cache them in memory at startup (similar to `ProdMaster`) keyed by `store_name + item_code`, or at minimum use a short-TTL in-process map keyed by `(branch, item_code)`.

---

### 6. `ValidateJWT` Creates a New gRPC Connection Per Request  
**File:** [pkg/authentication/authenticate.go](pkg/authentication/authenticate.go) · Lines 59–68

```go
func ValidateJWT(tokenStr string) (User, bool) {
    address := os.Getenv("LOGIN_RPC_ADDR")
    loginSvc, err := grpc.NewLoginService(address)  // new connection every call
```

Creating a gRPC client (and underlying TCP connection) on every authentication call is expensive. gRPC connections are designed to be long-lived and multiplexed.

**Fix:** Initialize one `loginSvc` at startup and reuse it:
```go
var loginSvc *grpc.LoginService  // package-level, initialized in main()

func ValidateJWT(tokenStr string) (User, bool) {
    rights, isValid := loginSvc.ValidateUserToken(...)
```

---

### 7. Prepared Statements Disabled  
**File:** [database/postgresql.go](database/postgresql.go) · Line 71

```go
pgConf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
```

`SimpleProtocol` disables the extended query protocol and prepared statement cache. Every SQL query is re-parsed by PostgreSQL on each execution. For high-frequency queries (`txn_log` inserts, `grn_items` scans, balance aggregates), this wastes CPU on both the Go side and the PostgreSQL server.

Remove this line to use the default `QueryExecModeCacheStatement` mode, which caches prepared statements per connection.

---

## Moderate Issues

### 8. `All()` Calls `sort.Slice` Inside the Loop  
**File:** [pkg/products/master.go](pkg/products/master.go) · Lines 280–299

```go
for _, item := range ProdMaster.ProductDB {
    item.StockCalcs()
    vals = append(vals, item)

    if i >= limit && limit > 0 {
        sort.Slice(vals, ...)  // sorts on every iteration until limit
        return vals, nil
    }
    i++
}
sort.Slice(vals, ...)  // also sorts at the end
```

When a `limit` is set, this sorts the partially-filled slice on each iteration after crossing the limit. Sort after collecting, not during:
```go
for _, item := range ProdMaster.ProductDB {
    item.StockCalcs()
    vals = append(vals, item)
    i++
    if limit > 0 && i >= limit {
        break
    }
}
sort.Slice(vals, func(i, j int) bool { return vals[i].ItemName < vals[j].ItemName })
```

### 9. `Query` Used for Single-Row Results  
**File:** [pkg/purchases/invoice.go](pkg/purchases/invoice.go) · Lines 55–82, 162–190, 312–341

`getGrnNum`, `invoiceIsExists`, and `priceChageExists` all open a `rows` cursor and loop through it for a query that returns at most one row. Use `QueryRow` / `Scan` directly:
```go
// Before
rows, err := database.PgPool.Query(ctx, sql, ...)
for rows.Next() { rows.Scan(&arg.GrnNum) }

// After
err = database.PgPool.QueryRow(ctx, sql, ...).Scan(&arg.GrnNum)
```

This avoids allocating the rows iterator and its internal buffer.

### 10. Balance Aggregation Done in SQL String Concatenation  
**File:** [pkg/products/storage.go](pkg/products/storage.go) · Lines 77–81

```sql
CONCAT('{"',location_id, '": ', SUM(qty_in - qty_out), '}') as balance
```

This builds a JSON string in SQL that is then `json.Unmarshal`-ed in Go. The balance subquery only returns one `(item_code, location_id)` row per item because it aggregates before joining. For items with balances in multiple locations, this only captures one location.

Use `jsonb_object_agg(location_id::text, SUM(qty_in - qty_out))` to return a proper JSONB object covering all locations in one pass, or fetch balances separately and map them in Go.

### 11. No HTTP Server Timeouts  
**File:** [app.go](app.go) · Line 261

```go
http.ListenAndServe(address+":"+port, r)
```

The non-TLS path uses the default `http.Server` with no read, write, or idle timeouts. Slow clients can hold open goroutines and connections indefinitely.

```go
srv := &http.Server{
    Addr:         address + ":" + port,
    Handler:      r,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}
srv.ListenAndServe()
```

---

## Low-Priority / Cleanup

### 12. Debug `fmt.Println` in Hot Paths  
**Files:** Multiple

The following diagnostic prints are active in production code on the hot request path:

| File | Line | Content |
|------|------|---------|
| [pkg/products/master.go](pkg/products/master.go) | 139 | `defer fmt.Printf("stockMaster GetByCode took %v")` — timing on every lookup |
| [pkg/products/master.go](pkg/products/master.go) | 181 | `fmt.Println("stock master combo_items =", ...)` — printed on every product fetch |
| [pkg/products/storage.go](pkg/products/storage.go) | 231 | `defer fmt.Printf("AddProduct took %v")` |
| [pkg/products/location.go](pkg/products/location.go) | 104 | `fmt.Println("\n\t store_name =", ...)` — every sale item location lookup |
| [pkg/products/location.go](pkg/products/location.go) | 171, 281 | SQL query printed to stdout on every location fetch |
| [pkg/count/stock_count.go](pkg/count/stock_count.go) | 63 | `fmt.Printf` on every count update |

Replace with structured logging at `DEBUG` level gated by a flag or environment variable so they can be enabled in development but are no-ops in production.

### 13. `error` from `time.Parse` Discarded in `GetGrn`  
**File:** [pkg/purchases/invoice.go](pkg/purchases/invoice.go) · Lines 96–97

```go
invdate, err := time.Parse(layout, arg.InvDate)
recvdate, err := time.Parse(layout, arg.RecvDate)  // first err overwritten
```

The parse error from `InvDate` is silently overwritten by the second assignment. Both should be checked.

### 14. Config File Read with Unchecked `json.Unmarshal`  
**File:** [app.go](app.go) · Line 86

```go
byteValue, _ := io.ReadAll(jsonFile)
json.Unmarshal(byteValue, &arg)
```

Both the `io.ReadAll` error and the `json.Unmarshal` error are discarded. A malformed config silently leaves `configs` as zero-value, resulting in confusing runtime behavior rather than a clear startup error.

---

## Prioritized Action Plan

| Priority | Issue | Est. Impact |
|----------|-------|-------------|
| P0 | Fix data race on `ProdMaster.ProductDB` | Correctness / crash prevention |
| P0 | Batch `Pickle()` calls — one per operation, not per item | 10–20× reduction in disk I/O per order |
| P1 | Fix `GetMultiCodes` O(n²) → O(1) map lookups | Eliminates per-location O(n) scan |
| P1 | Cache `stock_locations` in memory | Eliminates N DB round-trips per order |
| P1 | Re-use gRPC login service connection | Eliminates TCP dial cost per request |
| P2 | Remove `QueryExecModeSimpleProtocol` | Enables prepared statement caching |
| P2 | Fix `All()` sort inside loop | Minor CPU on catalog endpoints |
| P2 | Replace `Query` with `QueryRow` for single-row queries | Minor allocations |
| P3 | Add HTTP server timeouts | Resilience under slow/abusive clients |
| P3 | Remove / gate debug `fmt.Println` prints | Reduces stdout I/O and string formatting overhead |
| P3 | Fix balance JSON aggregation in SQL | Correctness — multi-location balances |
