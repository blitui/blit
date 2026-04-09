#!/usr/bin/env bash
# gen-gifs.sh — Run VHS on every .tape script under scripts/vhs/ and write
#               the resulting GIFs to docs/gifs/.
#
# Usage:
#   ./scripts/gen-gifs.sh              # render all tapes
#   ./scripts/gen-gifs.sh table form   # render only named tapes (no .tape ext)
#
# Requirements:
#   vhs must be installed: brew install vhs
#                     or:  go install github.com/charmbracelet/vhs@latest
#
# GIF size budget: each output GIF must be < 1.5 MB.
# If a GIF exceeds the budget, reduce Framerate or PlaybackSpeed in its .tape.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TAPES_DIR="$REPO_ROOT/scripts/vhs"
GIFS_DIR="$REPO_ROOT/docs/gifs"

# ── Dependency check ──────────────────────────────────────────────────────────
if ! command -v vhs &>/dev/null; then
  echo ""
  echo "  ERROR: 'vhs' is not installed or not on PATH."
  echo ""
  echo "  Install it with one of:"
  echo "    brew install vhs                                      # macOS / Linux (Homebrew)"
  echo "    go install github.com/charmbracelet/vhs@latest        # Go toolchain"
  echo "    scoop install vhs                                     # Windows (Scoop)"
  echo ""
  echo "  After installing, re-run: ./scripts/gen-gifs.sh"
  echo ""
  exit 1
fi

mkdir -p "$GIFS_DIR"

# ── Tape selection ────────────────────────────────────────────────────────────
if [[ $# -gt 0 ]]; then
  TAPES=()
  for name in "$@"; do
    tape="$TAPES_DIR/${name%.tape}.tape"
    if [[ ! -f "$tape" ]]; then
      echo "WARNING: tape not found: $tape" >&2
      continue
    fi
    TAPES+=("$tape")
  done
else
  mapfile -t TAPES < <(find "$TAPES_DIR" -name "*.tape" | sort)
fi

if [[ ${#TAPES[@]} -eq 0 ]]; then
  echo "No .tape files found under $TAPES_DIR" >&2
  exit 1
fi

# ── Render loop ───────────────────────────────────────────────────────────────
SIZE_BUDGET_BYTES=$((1536 * 1024))   # 1.5 MB
PASS=0
FAIL=0

for tape in "${TAPES[@]}"; do
  name="$(basename "$tape" .tape)"
  echo "  rendering $name ..."
  if vhs "$tape" --output "$GIFS_DIR/${name}.gif" 2>&1; then
    gif="$GIFS_DIR/${name}.gif"
    if [[ -f "$gif" ]]; then
      size=$(wc -c < "$gif")
      kb=$(( size / 1024 ))
      if (( size > SIZE_BUDGET_BYTES )); then
        echo "  WARNING: $name.gif is ${kb} KB — exceeds 1.5 MB budget!"
        echo "           Reduce Framerate or PlaybackSpeed in $tape"
        FAIL=$(( FAIL + 1 ))
      else
        echo "  OK       $name.gif (${kb} KB)"
        PASS=$(( PASS + 1 ))
      fi
    else
      echo "  OK       $name (no gif produced — vhs may have used inline Output)"
      PASS=$(( PASS + 1 ))
    fi
  else
    echo "  FAILED   $name" >&2
    FAIL=$(( FAIL + 1 ))
  fi
done

echo ""
echo "  Done: $PASS passed, $FAIL failed."
echo "  GIFs written to: $GIFS_DIR/"

if (( FAIL > 0 )); then
  exit 1
fi
