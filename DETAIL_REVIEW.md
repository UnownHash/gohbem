# Gohbem — Detailed Code Review

Scope: full source review (`ohbem.go`, `pvp_core.go`, `structs.go`, `utils.go`, `errors.go`, `cpm.go`, `cpm_test.go`, `ohbem_test.go`). Findings grouped by severity. Each item cites `file:line`.

Repo version: `0.12.0` (per `ohbem.go:18`).

---

## 1. Security

### 1.1 No HTTP timeout on `fetchMasterFile` — DoS / hang risk *(high)*
`utils.go:34` — `client := &http.Client{}` has no `Timeout`. A slow or stalled remote (`raw.githubusercontent.com`) can hang the calling goroutine indefinitely. Inside `WatchPokemonData` the watcher goroutine will block on `client.Do`, missing future tick events; if called synchronously by user code it can hang the server.

**Fix**:
```go
client := &http.Client{Timeout: 30 * time.Second}
```
Better: build a single package-level `*http.Client` reused across calls (avoid per-call TCP/TLS pool churn).

### 1.2 Unbounded response body — memory exhaustion *(medium)*
`utils.go:43` — `json.NewDecoder(resp.Body).Decode(&data)` reads without a size cap. A hostile or compromised endpoint can stream gigabytes. Wrap with `io.LimitReader`:
```go
err = json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&data)
```
(MasterFile is ~1–2 MiB; 32 MiB is generous.)

### 1.3 No HTTP status check *(medium)*
`utils.go:35` — non‑2xx responses (404, 503, captive‑portal HTML, GitHub rate limit) flow into `Decode`, returning the generic `ErrMasterFileDecode` and masking the real failure. Add:
```go
if resp.StatusCode != http.StatusOK {
    return PokemonData{}, ErrMasterFileFetch
}
```

### 1.4 `LoadPokemonData` reads arbitrary file with `os.ReadFile` *(low)*
`ohbem.go:34` — entire file slurped into memory; a maliciously huge JSON triggers OOM. Streaming via `os.Open` + `json.NewDecoder` with `LimitReader` mitigates. Acceptable for trusted callers but worth documenting.

### 1.5 `SavePokemonData` writes mode `0644` *(low)*
`ohbem.go:52` — world‑readable. The data is not sensitive, but if the path is shared, prefer `0600`. Also `os.WriteFile` is non-atomic; a crash mid-write leaves a corrupt cache file. Use temp-file + `os.Rename` for atomic replace:
```go
tmp := filePath + ".tmp"
if err := os.WriteFile(tmp, data, 0600); err != nil { return ErrMasterFileSave }
return os.Rename(tmp, filePath)
```

### 1.6 Data races on `Ohbem` shared state *(high)*
Multiple goroutines (caller code + watcher) read/write the same fields with no synchronization:

- `ohbem.go:105` watcher goroutine writes `o.PokemonData = pokemonData` while query methods (`QueryPvPRank`, `CalculateTopRanks`, `FindBaseStats`, `IsMegaUnreleased`, `CalculateCp`) read it concurrently. Race per `go vet -race`.
- `ohbem.go:150,245` lazy default `o.RankingComparator = RankingComparatorDefault` is a write from any caller goroutine. Concurrent writes of identical pointer are still races under the Go memory model.
- `ohbem.go:135` `o.compactRankCache = sync.Map{}` reassigns a `sync.Map` value (`structs.go:18`). Copying / replacing `sync.Map` while another goroutine calls `Load`/`Store` corrupts internal state. `go vet` flags `sync.Map` value copy.
- `ohbem.go:41,106` `o.PokemonData.Initialized = true` written without atomicity; `safetyCheck` reads it.

**Fix**:
- Add `sync.RWMutex` to `Ohbem`; take read lock in query methods, write lock in `Watch`/`Load`/`Fetch`/`ClearCache`.
- Change `compactRankCache sync.Map` → `compactRankCache *sync.Map` so it can be atomically swapped via `atomic.Pointer[sync.Map]`.
- Move `RankingComparator` default into the constructor (see §3.2) so the field is never written after init.

### 1.7 Watcher cannot be restarted *(medium)*
`ohbem.go:124` — `StopWatchingPokemonData` closes `o.watcherChan` but never sets it to `nil`. Subsequent `WatchPokemonData` returns `ErrWatcherStarted` (`ohbem.go:60`). Also a second `StopWatchingPokemonData` call would `close` a closed channel and panic. Fix:
```go
func (o *Ohbem) StopWatchingPokemonData() error {
    if o.watcherChan == nil { return ErrNilChannel }
    close(o.watcherChan)
    o.watcherChan = nil
    return nil
}
```

### 1.8 `errors.go` exports mutable error vars *(informational)*
`errors.go:6+` — `var Err… = errors.New(...)` is the standard Go pattern but allows callers to reassign. If you want immutability use `errors.New` returns wrapped behind a function, or accept the convention.

### 1.9 `User-Agent` header concatenation *(informational)*
`utils.go:32` — `fmt.Sprintf("Gohbem/%s", VERSION)` with `VERSION` const is safe today; if `VERSION` ever derives from user/runtime input, sanitize for CR/LF to prevent header injection.

---

## 2. Correctness / Logic Bugs

### 2.1 `FindBaseStats` overwrites `masterForm` in the wrong else *(bug, high)*
`ohbem.go:602–610`:
```go
if _, ok := masterPokemon.TempEvolutions[evolution]; ok && evolution != 0 {
    masterEvolution = masterPokemon.TempEvolutions[evolution]
} else {
    masterForm = Form{Attack: masterPokemon.Attack, ...}   // <-- clobbers form
}
```
The `else` branch resets `masterForm` to base stats even when the form *was* found earlier (lines 592–600). This silently strips per-form base stats whenever `evolution` is unset or unknown. Tests pass because none of the tested cases hit a "form found, evolution missing" path with mismatched form/base stats.

**Fix**: drop the entire else branch; `masterEvolution` remains zero and the existing fallthrough at lines 612–626 already handles it:
```go
if me, ok := masterPokemon.TempEvolutions[evolution]; ok && evolution != 0 {
    masterEvolution = me
}
```

Also note: this lookup ignores `masterForm.TempEvolutions` even though `Form` has its own (`structs.go:107`). Should mirror `QueryPvPRank` / `CalculateCp`: try `masterForm.TempEvolutions[evolution]` first, then fall back to `masterPokemon.TempEvolutions[evolution]`.

### 2.2 `FilterLevelCaps` mutates a value copy, not the result slice *(bug, high)*
`ohbem.go:644–673`:
```go
last = result[len(result)-1]   // copy
...
if last.Pokemon != 0 && last.Pokemon == entry.Pokemon ... {
    last.Cap = entry.Cap            // <-- writes to local copy only
    if entry.Capped { last.Capped = true }
}
```
`last` is a `PokemonEntry` value, not a pointer. The merge mutations (`last.Cap`, `last.Capped`) never propagate back into `result`. The collapsed entry keeps the *first* `Cap` it saw and never inherits `Capped=true` from a later level cap. Tests assert only `len(output)` so the bug is invisible.

**Fix**:
```go
ref := &result[len(result)-1]
if ref.Pokemon != 0 && ref.Pokemon == entry.Pokemon && ... {
    ref.Cap = entry.Cap
    if entry.Capped { ref.Capped = true }
} else {
    result = append(result, entry)
}
```
And drop the separate `last` var entirely — re-derive from `result` each iteration.

### 2.3 `CalculateTopRanks` master league: missing 15/15/15 *(bug, medium)*
`ohbem.go:314` — `for stamina := ivFloor; stamina < 15; stamina++`. The loop excludes `stamina == 15`, so the canonical 15/15/15 entry is never produced for the master league. Compare with `QueryPvPRank` (`ohbem.go:496`) where the special path is gated by `stamina < 15` *because* the trivial 15/15/15 case is handled elsewhere — but in `CalculateTopRanks` there is no elsewhere.

Verify against tests; the master test cases for `CalculateTopRanks` are commented out (`ohbem_test.go:110–111`), so this is uncovered.

**Fix**: change the bound to `<= 15` (the inner equality check `calculateHp(...,stamina,...) == maxHp` is true at stamina==15 trivially, so it self-handles).

### 2.4 `cacheKey` magic radix and overflow risk *(low)*
`ohbem.go:142`:
```go
cacheKey := int64(cpCap*999*999*999 + stats.Attack*999*999 + stats.Defense*999 + stats.Stamina)
```
The expression is evaluated in `int` (platform-dependent: 32-bit on 386/arm). On 32-bit platforms, `cpCap*999*999*999` overflows for any `cpCap >= 2`. Even on 64-bit, `999` is a strange radix — if a stat ever reaches `999` (unlikely but unguarded) collisions silently corrupt cache hits.

**Fix**: use bit packing — Pokémon stats fit comfortably in 16 bits, `cpCap` in ~14:
```go
cacheKey := int64(cpCap)<<48 | int64(stats.Attack)<<32 | int64(stats.Defense)<<16 | int64(stats.Stamina)
```
Faster (no multiplications), no overflow, no collision.

### 2.5 `WatchPokemonData` writes new data even when save fails *(logic)*
`ohbem.go:105–115`: `o.PokemonData = pokemonData` happens *before* the cache file save. If the save fails, in-memory data is the new version but the cache file still points at the old one. On next process start, `LoadPokemonData(MasterFileCachePath)` loads stale data while the service had been running on fresh data. Either:
- Save first, swap second, or
- Log loudly and accept (current behavior with no swap rollback is fine if documented).

### 2.6 `calculateCpMultiplier` half-level branch — verify rounding *(needs validation, low)*
`pvp_core.go:15–20`:
```go
baseCpm := float64(float32(0.5903 + float64(baseLevel)*0.005))
...
nextCpm := float64(float32(0.5903 + float64(baseLevel+1)*0.005))
return math.Sqrt((baseCpm*baseCpm + nextCpm*nextCpm) / 2)
```
Used only for level > 55 (which the masterfile multiplier table doesn’t supply). The double `float64(float32(...))` is intentional (Niantic’s client uses single-precision multipliers), but Levels 1–55 use the precomputed table from `cpm.go` and Levels >55 do not exist in current Pokémon GO (caps are 50/51). Unreachable code; consider removing or asserting.

### 2.7 `calculatePvPStat` binary search midpoint *(verify)*
`pvp_core.go:52`:
```go
mid := math.Ceil(lowest+highest) / 2
```
Reads as `Ceil(sum) / 2`, not `Ceil(sum/2)`. Because `lowest`/`highest` step in 0.5, `lowest+highest` is always a multiple of 0.5, so `Ceil` only acts on `*.5` sums. The result still lands on a 0.5 grid. Behavior matches existing tests, but the parenthesization is non-obvious — add `(lowest+highest)/2` rounded to nearest 0.5 with explicit helper, or comment why.

### 2.8 `IsMegaUnreleased` parameter naming mismatch *(doc)*
`ohbem.go:630` — second parameter is named `evolution` but tests pass *form* values (`{150, 2, true}` is Mewtwo Mega-Y). Rename to clarify or align with TempEvolution map keys.

### 2.9 `QueryPvPRank` recursion has no cycle guard *(defensive)*
`ohbem.go:553` recursively calls `o.QueryPvPRank` for each `evolution.Pokemon`. If MasterFile data is ever malformed (cyclic evolutions), this stack‑overflows. Track a small visited set on the stack or cap recursion depth.

### 2.10 Fields written via positional struct literal *(fragile)*
`ohbem.go:525,527`:
```go
pushAllEntries(&PokemonStats{masterForm.Attack, masterForm.Defense, masterForm.Stamina, false}, 0)
```
Positional init breaks silently if `PokemonStats` field order changes (`structs.go:119`). Use named fields:
```go
&PokemonStats{Attack: masterForm.Attack, Defense: masterForm.Defense, Stamina: masterForm.Stamina}
```
(`Unreleased` defaults to false anyway.)

### 2.11 `CalculateCp` doesn’t validate IV/level range *(consistency)*
`ohbem.go:357` — unlike `QueryPvPRank` (`ohbem.go:400`), `CalculateCp` accepts negative IVs, IVs > 15, and level < 1 silently. Either validate consistently or document the divergence.

---

## 3. Optimizations

### 3.1 Hash → byte compare in watcher *(easy win)*
`ohbem.go:91–101` marshals both old and new payloads then SHA-256s them. SHA‑256 is unnecessary; equal bytes is sufficient and twice as fast:
```go
if bytes.Equal(newData, oldData) { continue }
```
Even better: cache the previous marshalled bytes (or its hash) on the `Ohbem` struct so each tick only marshals the *new* fetched payload. Saves ~50% of the work per tick.

Drop the unused import `crypto/sha256` afterwards.

### 3.2 Lazy `RankingComparator` default — hoist to constructor *(perf + race fix)*
`ohbem.go:149,244` runs the nil check on every call. Set once in `FetchPokemonData`/`LoadPokemonData` (or expose `NewOhbem(...)` factory). Eliminates the per-call branch *and* fixes §1.6.

### 3.3 `calculateRanksCompact` is the hot loop — micro-opts *(perf)*
`pvp_core.go:188` inner triple loop is called per stat triple per level cap. Easy wins:
- Move `calculatePvPStat` allocation: `out` is already a pointer write, good. But `multiplier := calculateCpMultiplier(level)` (`pvp_core.go:34`) is called from `calculateCp` inside `calculatePvPStat`'s binary-search loop AND once for the final level — recompute fine, but cache the `multiplier*multiplier` product to skip a mul.
- `calculateCp` inside binary search: ignore the `< 10` clamp during search — the clamp only affects the very low end where the search has long since narrowed. Saves a branch per iteration. (Verify with tests first.)
- Pre-allocate the `[4096]PvPRankingStats` once in `Ohbem` (sync.Pool) instead of `new(...)` per call. With cache disabled (master league) this allocates ~16 KiB per call.
- For default comparator path, avoid `sort.Sort` interface calls (one method dispatch per `Less`/`Swap`). A monomorphic `sort.Slice` over a typed slice or hand-rolled introsort is measurably faster on 4096 items in this hot path. Even better: for the default comparator, the keys are `(value, attack)` — sort with `slices.SortFunc` (Go 1.21+) which is faster than `sort.Sort`.

### 3.4 `roundFloat(x, 5)` — eliminate `math.Pow` *(easy)*
`utils.go:13` — `math.Pow(10, float64(precision))` for a constant 5 in every caller. Replace with a constant:
```go
const roundFactor5 = 100000.0
func roundPercent(v float64) float64 { return math.Round(v*roundFactor5) / roundFactor5 }
```
or specialize `roundFloat` with a switch on `precision`. The current generic version dominates `Percentage` computation cost.

### 3.5 `containsInt` linear scan *(low)*
`utils.go:18` — fine for short slices (`CostumeOverrideEvolutions` rarely > 5). Leave as is unless profiling indicates otherwise.

### 3.6 `calculateAllRanksCompact` cache key build *(see §2.4)*
Bit packing also wins ~3 ns per call.

### 3.7 `QueryPvPRank` sort of `combinationIndexKeys` *(perf)*
`ohbem.go:465–471` allocates a slice and sorts to iterate `combinationIndex` in ascending level order. But the level keys come from `o.LevelCaps` (already user-ordered) plus optional `MaxLevel`. Iterate `o.LevelCaps` directly with a presence check; append `MaxLevel` if present. Avoids the per-call alloc + sort.

### 3.8 `CalculateTopRanks` parallel slice `lastRank` / `lastRankIdx` *(cleanup)*
`ohbem.go:250–303` keeps two parallel slices growing in lock-step. Combine into one struct slice for clarity and cache locality:
```go
type lastEntry struct { ranking Ranking; idx int }
var last []lastEntry
```

### 3.9 Rebuild `*http.Client` per call *(low)*
`utils.go:34` — once §1.1 is in place, also share the client across calls (package var or field on `Ohbem`).

### 3.10 `json.Marshal` of full `PokemonData` for change detection *(see §3.1)*
After fetching, you have raw bytes from `resp.Body`. Decode into `PokemonData` *and* keep the raw bytes (read into buffer first via `io.ReadAll(io.LimitReader(...))`, then `json.Unmarshal`). Compare raw incoming bytes against the previous raw bytes — no remarshal needed.

### 3.11 `sync.Map` is not the best fit for `compactRankCache` *(consider)*
`structs.go:18` — `sync.Map` shines for write-once-read-many keys with disjoint key sets across goroutines. Here keys are derived from `(cpCap, stats)` — heavily shared. A `map[int64]map[int]compactCacheValue` guarded by `sync.RWMutex` (or sharded map) often outperforms `sync.Map` for this access pattern. Benchmark before changing — `BenchmarkCalculateAllRanksCompactCached` is your yardstick.

---

## 4. Deduplication / Cleanup

### 4.1 Repeated "lookup form / fallback to base / lookup evolution" pattern
The same triple‑lookup logic appears in `CalculateTopRanks` (lines 198–242), `CalculateCp` (357–388), `QueryPvPRank` (404–427 + 524–528), and `FindBaseStats` (584–626). Extract a helper:
```go
func (o *Ohbem) resolveStats(pokemonId, form, evolution int) (PokemonStats, Form, Pokemon, bool) {
    mp, ok := o.PokemonData.Pokemon[pokemonId]
    if !ok { return PokemonStats{}, Form{}, Pokemon{}, false }
    mf, ok := mp.Forms[form]
    if !ok || form == 0 {
        mf = Form{Attack: mp.Attack, Defense: mp.Defense, Stamina: mp.Stamina,
                  Little: mp.Little, Evolutions: mp.Evolutions,
                  TempEvolutions: mp.TempEvolutions, CostumeOverrideEvolutions: mp.CostumeOverrideEvolutions}
    }
    var stats PokemonStats
    if me, ok := mf.TempEvolutions[evolution]; ok && evolution != 0 && me.Attack != 0 {
        stats = me
    } else if me, ok := mp.TempEvolutions[evolution]; ok && evolution != 0 {
        stats = me
    } else if mf.Attack != 0 {
        stats = PokemonStats{Attack: mf.Attack, Defense: mf.Defense, Stamina: mf.Stamina}
    } else {
        stats = PokemonStats{Attack: mp.Attack, Defense: mp.Defense, Stamina: mp.Stamina}
    }
    return stats, mf, mp, true
}
```
Eliminates ~120 lines and forces a single, testable resolution rule (which would have caught §2.1 / §2.8).

### 4.2 Commented-out `calculateRanks` / `TestCalculateRanks` *(cleanup)*
`pvp_core.go:75–113` and `pvp_core_test.go:158–238` are large commented blocks. Either delete or move to a separate experimental file. Dead code rots and confuses readers. Git history preserves it.

### 4.3 `goland:noinspection` comment *(cleanup)*
`utils.go:39` — IDE-specific marker. Either suppress globally in a config or use Go's idiom: just `defer resp.Body.Close()` is fine; staticcheck/`errcheck` won’t flag a deferred Close in stdlib usage.

### 4.4 Lazy `RankingComparator` default duplicated *(see §3.2)*
Two copies (`ohbem.go:149` and `:244`).

### 4.5 `else` after `return` in `StopWatchingPokemonData` *(style)*
`ohbem.go:124–131` — `if cond { return … } else { close(...) }` simplifies to early return.

### 4.6 `containsInt` could be `slices.Contains` *(Go 1.21+)*
`utils.go:18` — replace with stdlib `slices.Contains` once go.mod allows.

### 4.7 Unused `Index` field on `Ranking` *(cleanup)*
`structs.go:71` — `Index int` is in the JSON shape with `omitempty` but never set anywhere in code. Either remove or document.

### 4.8 Inconsistent error variable scope
`errors.go` mixes errors that reflect user input (`ErrQueryInputOutOfRange`, `ErrMissingPokemon`), runtime state (`ErrMasterFileUnloaded`, `ErrLeaguesMissing`), and remote IO (`ErrMasterFileFetch`). Group with comments or split into `errors_input.go` / `errors_io.go` / etc.

---

## 5. Logic / API Recommendations

### 5.1 Add `New(config)` constructor *(API)*
Currently callers build `Ohbem{...}` directly and the library has to defend against missing fields lazily on every call (`safetyCheck`, comparator default, cache init). A constructor:
```go
func New(opts Options) (*Ohbem, error) { ... }
```
- centralizes defaults (comparator, watcher interval, cache),
- returns `*Ohbem` to make the mutex semantics obvious,
- removes need for lazy mutation in hot paths,
- enables marking unexported fields explicitly.

### 5.2 Receivers should be pointer everywhere *(consistency)*
All receivers are `*Ohbem` already; with the cache & watcher state, callers must always pass pointers. Document this in README and in the type doc comment.

### 5.3 Watcher API *(ergonomics)*
- Provide `WatchPokemonDataContext(ctx context.Context)` so users can cancel via context instead of a `Stop` method (Go idiom). Drop the bool channel.
- Surface fetch errors via a callback or error channel rather than a `Logger.Print` string.

### 5.4 `Logger` should accept levels *(observability)*
`structs.go:37` — single `Print(string)` flattens info, warning, error. Either:
- Adopt `log/slog` (`*slog.Logger`), or
- Pass `(level, msg, fields...)`.

### 5.5 Document concurrency contract
README (`README.md:55`) shows `ohbem := gohbem.Ohbem{...}` — a value, not pointer. With `sync.Map` / channel state, mixing values and pointers risks copying. Add a "Concurrency" section: must-be-pointer, methods are safe under the rules of §1.6 once the mutex is added.

### 5.6 `CalculateTopRanks` "broken" tag in README *(action)*
`README.md:92` flags the function as broken. The recent `len-fixes` branch (`695642a fix: uncomment and fix CalculateTopRanks…`) indicates work in progress. Once the master-league bug (§2.3) and the commented capped tests (`ohbem_test.go:110–111`) are addressed, drop the "broken" tag.

### 5.7 `DisableCache` semantics *(API)*
`structs.go:13` — `DisableCache bool`. Inverted-default flags read awkwardly. Prefer `EnableCache bool` defaulting to false, *or* make caching the default and offer a `WithoutCache()` option.

### 5.8 MasterFile URL hardcoded *(API)*
`utils.go:11` — `MasterFileURL` is a top-level `const`. Make it overridable on `Ohbem` (e.g. `MasterFileURL string`) so users can self-host or air-gap.

### 5.9 Error wrapping *(idiomatic)*
All errors are sentinel-only (`errors.go`). The actual underlying cause (network error, file path, JSON offset) is discarded. Wrap with `fmt.Errorf("fetch masterfile: %w", err)` and keep the sentinel via `errors.Is`.

### 5.10 `safetyCheck` duplicates `o != nil` assumption *(defensive)*
`utils.go:51` — passing a nil `*Ohbem` panics before `safetyCheck` runs. Add `if o == nil { return ErrNotInitialized }` if `*Ohbem` is the public type.

---

## 6. Test Gaps

- `CalculateTopRanks` master league not exercised (related to §2.3).
- `FilterLevelCaps` only asserts `len(output)`, not `Cap` / `Capped` (related to §2.2).
- No `-race` run in CI (`.github/workflows/test.yml` not inspected here, but adding `go test -race ./...` will surface §1.6 immediately).
- No fuzz tests for `QueryPvPRank` IV bounds.
- No test for `WatchPokemonData` start→stop→start sequence (related to §1.7).
- No test for `FetchPokemonData` error paths (timeout, non-200, malformed JSON).

---

## 7. Priority Punch‑list

| # | Item | Severity | Effort |
|---|------|----------|--------|
| 1 | Add HTTP timeout + LimitReader + status check (§1.1–1.3) | High | S |
| 2 | Fix data races: mutex + atomic cache pointer (§1.6) | High | M |
| 3 | Fix `FindBaseStats` else clobber (§2.1) | High | XS |
| 4 | Fix `FilterLevelCaps` value-copy mutation (§2.2) | High | XS |
| 5 | Restartable watcher + nil-out channel (§1.7) | Medium | XS |
| 6 | Master-league `<= 15` loop bound (§2.3) | Medium | XS |
| 7 | Replace SHA-256 with byte compare; reuse last bytes (§3.1, §3.10) | Medium | S |
| 8 | Bit-pack cacheKey (§2.4) | Medium | XS |
| 9 | Atomic file write for SavePokemonData (§1.5) | Low | XS |
| 10 | Extract `resolveStats` helper (§4.1) | Low | M |
| 11 | Move comparator default to constructor (§3.2) | Low | S |
| 12 | Drop dead commented blocks (§4.2) | Low | XS |
| 13 | `go test -race` in CI | High | XS |
| 14 | Fuzz tests for IV/level bounds (§6) | Low | M |

---

## 8. Summary

The library is small, focused, and the hot path is well shaped — `calculateRanksCompact` shows real attention to allocation and sort cost. Two real correctness bugs (`FindBaseStats` else clobber, `FilterLevelCaps` value-copy mutate) are masked by gaps in test assertions; both are one-line fixes. The most pressing risks are the **HTTP client without timeout** and **unsynchronized shared state** between the watcher goroutine and query callers — both will bite under production load. The proposed refactor (constructor + mutex + helper resolver) will simultaneously fix the races, eliminate ~120 lines of duplication, and make further optimizations (atomic-pointer cache swap, sync.Pool for the 16 KiB rank arena) safe to apply.

---

## 9. Status (post-fix pass — 2026-05-05)

| Item | Status | Notes |
|------|--------|-------|
| 1.1 HTTP timeout | ✅ done | Package-level `httpClient` with 30s `Timeout` (`utils.go`). |
| 1.2 LimitReader 32 MiB | ✅ done | `io.LimitReader(resp.Body, 32<<20)` in `fetchMasterFile`. |
| 1.3 HTTP status check | ✅ done | Non-200 → `ErrMasterFileFetch`. |
| 1.4 LoadPokemonData OOM | ⏭ skipped | Local file is trusted-caller scope; HTTP path covered by 1.2. |
| 1.5 SavePokemonData mode/atomic | ⏭ skipped | Per request. Watcher path uses tmp+Rename anyway (see 2.5). |
| 1.6 Data races | ✅ done | `sync.RWMutex` on `Ohbem`; `compactRankCache` is `atomic.Pointer[sync.Map]`; comparator default set under lock. |
| 1.7 Restartable watcher | ✅ done | `StopWatchingPokemonData` nils channel; second stop returns `ErrNilChannel`; covered by `TestWatchPokemonDataRestart`. |
| 1.8 Mutable error vars | ⏭ informational | Convention kept. |
| 1.9 User-Agent injection | ✅ done | `VERSION` is a `const`; no runtime input flows into header. |
| 2.1 `FindBaseStats` else clobber | ✅ done | Replaced with `resolveStats` helper (no else-clobber path). |
| 2.2 `FilterLevelCaps` value copy | ✅ done | Mutations applied via `&result[len-1]`. Regression test `TestFilterLevelCapsMerge`. |
| 2.3 Master league `<= 15` | ✅ done | Loop bound updated; `CalculateTopRanks` master-league path now produces 15/15/15. |
| 2.4 Cache key bit-pack | ✅ done | `cacheKey()` packs `(cpCap, attack, defense, stamina)` into one int64. |
| 2.5 Watcher save before swap | ✅ done | Tmp-file + `os.Rename`; on save failure, in-memory swap is skipped. |
| 2.6 `calculateCpMultiplier` half-level | ⏭ skipped | Per request. |
| 2.7 Binary search midpoint | ⏭ verify only | Behavior preserved; left as is. |
| 2.8 `IsMegaUnreleased` arg name | ✅ done | Renamed second arg to `tempEvolution`; doc clarifies it is a TempEvolution key. |
| 2.9 Recursion cycle guard | ✅ done | Depth bound (`maxEvolutionDepth=8`) on `queryPvPRankInternal`. |
| 2.10 Positional struct literal | ✅ done | `resolveStats` builds `PokemonStats` with named fields throughout. |
| 2.11 `CalculateCp` IV/level validation | ✅ done | Same `ErrQueryInputOutOfRange` gate as `QueryPvPRank`. |
| 3.1 Hash → byte compare | ⏭ skipped | Per request. |
| 3.2 Comparator default in init | ✅ done | Hoisted into `LoadPokemonData` / `FetchPokemonData` under write lock; lazy fallback kept inside `calculateAllRanksCompact` for safety but no longer the primary path. |
| 3.3 Hot loop micro-opts | ✅ partial | `sync.Pool` for the 16 KiB `[4096]PvPRankingStats` arena (`releaseRankArena`); `sort.Sort` interface replaced with `slices.SortFunc`. |
| 3.4 `roundFloat(x, 5)` constant | ✅ done | Fast path with `roundFactor5` in `utils.go`. |
| 3.5 `containsInt` linear scan | ✅ done | Removed; replaced with `slices.Contains`. |
| 3.6 cacheKey perf | ✅ done | Covered by 2.4. |
| 3.7 `QueryPvPRank` sort | ✅ done | Iterates `o.LevelCaps` directly + optional `MaxLevel`; no per-call alloc/sort. |
| 3.8 Parallel slice cleanup | ✅ done | Replaced `lastRank` / `lastRankIdx` with one `[]lastEntry`. |
| 3.9 Reuse `*http.Client` | ✅ done | Package-level `httpClient`. |
| 3.10 Keep raw bytes | ⏭ skipped | Tied to 3.1. |
| 3.11 sync.Map vs RWMutex map | ⏭ deferred | Needs benchmark; cache is now `atomic.Pointer[sync.Map]` so this can be revisited safely. |
| 4.1 `resolveStats` helper | ✅ done | New helper in `ohbem.go`; used by `CalculateTopRanks`, `CalculateCp`, `QueryPvPRank` (`queryPvPRankInternal`), `FindBaseStats`. |
| 4.2 Dead commented blocks | ✅ done | Removed from `pvp_core.go` and `pvp_core_test.go`. |
| 4.3 `goland:noinspection` | ✅ done | Removed; deferred Close kept. |
| 4.4 Comparator default duplicate | ✅ done | Centralized via 3.2. |
| 4.5 `else` after return in Stop | ✅ done | Early-return form. |
| 4.6 `slices.Contains` | ✅ done | All call sites. |
| 4.7 Unused `Index` on `Ranking` | ✅ done | Field removed. |
| 4.8 Errors regrouped | ✅ done | Grouped into Input / Runtime / I/O / Internal blocks in `errors.go`. |
| 5.6 Drop "broken" tag | ✅ done | README heading updated. |
| 5.8 MasterFileURL configurable | ✅ done | New `Ohbem.MasterFileURL` field; `fetchMasterFile(url)` accepts override; default fallback preserved; documented in README. |
| 6.* Test gaps | ✅ done | Master-league `TestCalculateTopRanks` cases re-enabled; `TestFilterLevelCapsMerge` asserts Cap+Capped propagation; `TestWatchPokemonDataRestart` covers start→stop→start + double-stop; `TestFetchPokemonDataNon200` / `TestFetchPokemonDataMalformed` cover error paths via `httptest`; `TestSavePokemonDataAtomic` round-trips a save; `FuzzQueryPvPRankBounds` and `FuzzCalculateCpBounds` cover IV/level bounds; CI workflow now runs `go test -race`. |

### Verification

- `go build ./...` clean.
- `go vet ./...` clean.
- `go test -race ./...` passes (`ok github.com/UnownHash/gohbem`).
- VERSION bumped to `0.13.0` to reflect API additions (`Ohbem.MasterFileURL`, removal of `Ranking.Index`, `IsMegaUnreleased` arg rename).

### Items intentionally left

- 1.4 (file size cap), 1.5 (Save mode/atomic), 1.8 (mutable error vars), 2.6 / 2.7 (CPM / midpoint deep-dive), 3.1 / 3.10 (raw-byte change detection), 3.11 (sync.Map alternative — needs benchmark), 5.x items not requested (constructor, context-based watcher, Logger levels, error wrapping, `DisableCache` rename, etc.).
