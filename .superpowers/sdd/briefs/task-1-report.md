# Task 1 Report: Payload Registry

## Files Created
- `internal/payloads/registry.go` — types (`StealthLevel`, `Category`, `Payload`), the
  `Registry` data (5 categories: sqli, xxe, dos, integrity, ml_poison, 16 payloads total),
  and helper functions `GetCategory`, `GetPayload`, `ProfileIDs`, `GetProfile`.
- `internal/payloads/registry_test.go` — 16 test functions covering category/payload
  presence, lookups (including unknown/malformed IDs), SQLi template content, impact
  value validity, all 5 profiles, payload ID uniqueness, and JSON marshaling to confirm
  `Template` is excluded from output.

## Design Notes
- All payloads use `Stealth: []int{1,2,3,4,5}` (all levels compatible) since the brief
  didn't specify per-payload stealth restrictions — this satisfies "All stealth levels
  should be []int slices" while keeping every payload usable across delivery mechanisms.
  `stealth_recon` profile filters to payloads compatible with level 4 (Shadowed) or 5
  (Polyglot), which currently includes all payloads since none restrict Stealth.
- Payloads without a literal template in the brief (billion_laughs, dos.*, integrity.*,
  ml_poison.*) have `Template: ""` with a comment noting they're generated dynamically,
  matching the brief's "(generated dynamically)" notes.
- `GetPayload` parses "category.payload" by splitting on the first `.`.
- Profiles are built once via `buildProfiles()` into a package-level map, derived from
  `Registry` contents so `sqli_focus`/`xxe_focus`/`dos_focus`/`full_audit` stay in sync
  automatically if the registry changes.

## Test Results

Command: `go vet ./internal/payloads/... && go test ./internal/payloads/... -v`

```
=== RUN   TestRegistryCategoriesDefined
--- PASS: TestRegistryCategoriesDefined (0.00s)
=== RUN   TestCategoryColors
--- PASS: TestCategoryColors (0.00s)
=== RUN   TestGetCategoryUnknown
--- PASS: TestGetCategoryUnknown (0.00s)
=== RUN   TestGetPayloadLookup
--- PASS: TestGetPayloadLookup (0.00s)
=== RUN   TestGetPayloadUnknownOrMalformed
--- PASS: TestGetPayloadUnknownOrMalformed (0.00s)
=== RUN   TestSQLiTemplates
--- PASS: TestSQLiTemplates (0.00s)
=== RUN   TestImpactValuesAreValid
--- PASS: TestImpactValuesAreValid (0.00s)
=== RUN   TestProfileIDs
--- PASS: TestProfileIDs (0.00s)
=== RUN   TestGetProfileSQLiFocus
--- PASS: TestGetProfileSQLiFocus (0.00s)
=== RUN   TestGetProfileXXEFocus
--- PASS: TestGetProfileXXEFocus (0.00s)
=== RUN   TestGetProfileDoSFocus
--- PASS: TestGetProfileDoSFocus (0.00s)
=== RUN   TestGetProfileFullAudit
--- PASS: TestGetProfileFullAudit (0.00s)
=== RUN   TestGetProfileStealthRecon
--- PASS: TestGetProfileStealthRecon (0.00s)
=== RUN   TestGetProfileUnknown
--- PASS: TestGetProfileUnknown (0.00s)
=== RUN   TestPayloadIDsUniqueWithinCategory
--- PASS: TestPayloadIDsUniqueWithinCategory (0.00s)
=== RUN   TestTemplateNotExposedInJSON
--- PASS: TestTemplateNotExposedInJSON (0.00s)
PASS
ok  	github.com/novusedge/hirejack/internal/payloads	0.004s
```

16/16 tests passing. `gofmt -l` clean. `go build ./...` and `go test ./...` (full repo) pass.

## Concerns
- The working tree had a large set of pre-existing uncommitted changes unrelated to this
  task (frontend files, go.mod/go.sum, main.go, justfile, etc.) already staged/modified
  before this task began. This commit stages only the two new payloads files to avoid
  bundling unrelated changes.
- No per-payload stealth-level differentiation was specified in the brief beyond the
  `stealth_recon` profile filter, so all payloads currently declare full 1-5 compatibility.
  If later tasks need payloads that are *only* deliverable at low stealth (e.g. `basic`
  raw payloads that can't be polyglot-wrapped), this may need revisiting — flagging for
  visibility, not blocking.
