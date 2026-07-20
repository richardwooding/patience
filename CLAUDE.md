# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

patience is a multi-variant solitaire game: a pure-Go rules engine under an
Ebitengine front end. Eight configurations across five rule families —
Klondike draw-1/3, FreeCell, Spider 1/2/4 suits, Golf, Pyramid. One codebase
runs native (`go run .`, the dev loop) and as WebAssembly on GitHub Pages
(https://richardwooding.github.io/patience/play/).

## Commands

```sh
go test ./internal/solitaire/                # entire rules core, fully headless
go test ./internal/solitaire/ -run TestUndoRoundTrip -v   # one test
go run .                                     # native window (macOS)
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o docs/play/patience.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/play/
(cd docs/play && python3 -m http.server 8080)   # ?v=freecell etc. deep-links a variant
go vet ./... && gofmt -l .                   # CI also runs golangci-lint (latest) — errcheck and unused are strict
```

CI runs `go test` under `xvfb-run` because importing Ebitengine panics at init
on a display-less Linux runner. This matters the moment `internal/ui` gains a
test file; `internal/solitaire` must never need it (see below). PRs also run
the codemetrics complexity ratchet — any function a PR *changes* must stay
under cognitive complexity 15.

## Architecture and invariants

**`internal/solitaire` is a pure rules engine with zero Ebitengine imports.**
Keep it that way — every test in it runs headlessly, and the seeded-deal
goldens depend on that.

- **A card is one byte**: `Card uint8` — rank in bits 0–3, suit in 4–5,
  face-up in bit 6. Decks come from `DeckSpec{Decks, Suits}` (Spider 1-suit =
  8×♠, always 104 cards); `Shuffled(spec, seed)` is a Fisher–Yates over
  `rand/v2` PCG(seed, seed). Golden tests pin exact deals per seed — if a
  shuffle change is intentional, regenerate the goldens knowingly.
- **Variants are hooks, not data tables.** Each implements `Rules`
  (`Layout/Deal/CanPickUp/CanDrop/AfterMove/TapStock/Won/SafeMoves/
  AutoCompleteReady`); pile indices are fixed per variant by `Layout()`.
  Shared predicates (`descAltColor`, `descSameSuit`, `isRun`,
  `foundationDrop`) live in rules.go.
- **Undo is a snapshot per move**, deep-copied piles. Chosen over reverse
  deltas because Spider's AfterMove chains (run removal + flips) and Klondike
  recycles make deltas a bug farm. `TestUndoRoundTrip` plays 400 random moves
  per config and DeepEquals its way back — it must stay green.
- **Spider's AfterMove is a fixpoint loop**: flip exposed cards, remove
  complete same-suit K→A runs to the next empty foundation pile, repeat. Run
  removal is a *normal pile move*, so undo and win detection need no special
  cases. TapStock refuses while any column is empty (`ErrEmptyColumn`).
- **FreeCell supermoves**: capacity `(freeCells+1) << emptyCols`, halved when
  the destination itself is an empty column (the `ti != dst` exclusion in
  `capacity()` — there's a dedicated off-by-one test).
- **Golf** (`golf.go`): 7 columns of 5, one foundation built ±1 rank (no
  wrap), stock deals to the foundation. A clean fit for the Move model — every
  move is a tableau top onto the single Foundation pile.
- **Pyramid** (`pyramid.go`): 28 slots in a 7-row triangle, row-major from
  `pyramidBase` (pile 3). A slot is `exposed` when both covering child piles
  are empty (precomputed `children[28][2]`). Pairs summing to 13 are played by
  *dropping* one exposed card onto another (or from the waste); AfterMove
  sweeps any slot left holding a sum-13 pair to the foundation, and a lone King
  drops straight to the foundation. This reuses the ordinary Move machinery, so
  undo/win need no special cases. In the UI, slots are laid out row-major so
  the existing `hitCard`/draw order handles the overlap (lower rows draw on top
  and win hit-tests); empty pyramid slots set `hideSlot` so a cleared card
  leaves a gap, not an outline.
- **Safe auto-send rule** (`safeFoundationSends`): a card may auto-send when
  rank ≤ 2 or both opposite-color foundations have reached rank−1; the floor
  is 0 until at least 2 same-color foundations are started. Double-tap uses
  any-legal send; the auto-complete loop uses safe-only.

**`internal/ui` details worth knowing:**

- Card sprites (71×96, Windows 3.x dims) are generated programmatically at
  init — corner indices via bitmapfont glyphs, center pips from hand-drawn
  16×16 `uint16` bitmasks in pips.go drawn at 2×. No image assets.
- Fan offsets (18 face-up / 8 face-down / 15 waste) compress dynamically with
  floors (10/4) so Spider's deep columns never overflow the 600px canvas —
  `fannedPositions` in layout.go.
- Drop targeting is by **max rect-overlap of the dragged card**, not pointer
  position (better on touch); touch hits are inflated 6px. The drag machine
  lives in tablescene.go: pressed → 4px threshold → dragging ghost stack →
  drop or snap-back flight.
- Flights (anim.go) are presentation-only: the model updates first, the
  sprite flies after. `hiddenDepths` keeps in-flight cards from double-drawing.
- The win cascade draws bouncing cards onto a persistent never-cleared
  offscreen image — trails cost nothing (chipdeck phosphor-layer pattern).
- `?v=<variant-id>` deep-links a variant (autostart_js.go); it's also how
  headless screenshots reach table scenes (headless Chrome needs
  `--use-angle=swiftshader --enable-unsafe-swiftshader` for WebGL2).
- Stats are build-tagged: localStorage (js) / `os.UserConfigDir()` (native).
  The persisted shape is `{variants, daily}`; `ensure()` migrates legacy files
  that were a bare `map[string]Entry`. Daily streak math lives in the pure
  `applyWin` helper (unit-tested); `RecordDailyWin` just persists its result.
- **Daily deals**: `solitaire.DayNumber(now)` is the local-calendar day count
  from a FROZEN epoch (2026-01-01); `solitaire.DailySeed(id, day)` is a frozen
  FNV-1a + splitmix64 mix — both guarded by committed golden tests, because the
  seed is shared across players and baked into share links. **Never change the
  epoch or the mix.** The menu toggles daily mode (`D`); `newDailyScene`
  seeds from `DailySeed`; `?v=<id>&d=<day>` deep-links a specific daily.
- Share blurb: pure `internal/share.Text` (emoji OK — it goes to the clipboard,
  not the bitmap-font canvas, which stays ASCII); `copyToClipboard` is
  build-tagged (navigator.clipboard on js, stdout natively).
- `Game.Hint()` ranks legal card moves (expose face-down > safe send >
  build on a non-empty pile > any) and drives the `H` key / hint button,
  which blinks `accentOutline` on the move's source and destination. The
  dead-game notice comes from `AnyLegalMove`, recomputed only when
  `MoveCount` changes (it clones the game, so it stays off the per-frame path).

## Deploy

Push to `main` touching `docs/**`, `internal/**`, or `main.go` triggers
`pages.yml`: it builds the wasm into `docs/play/` (the wasm and `wasm_exec.js`
are gitignored, never committed) and deploys `docs/` to GitHub Pages. The
landing page is gloam-styled with vendored `gloam.css`/`gloam.js`, kept in
sync by the weekly `gloam-sync.yml` PR workflow.
