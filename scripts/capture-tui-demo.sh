#!/usr/bin/env bash
# Regenerates assets/tui.gif — a VHS recording of `genv tui` driven against a
# throwaway, deterministic demo repo. Self-contained and path-portable: it
# derives the repo root from its own location and builds the demo in a fresh
# temp dir, so it reruns on any checkout.
#
#   bash scripts/capture-tui-demo.sh
#
# Requires: vhs, ttyd, ffmpeg (vhs deps), gifsicle (lossless optimize), go.
#   brew install vhs ttyd gifsicle go
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$(mktemp -d)/genv-tui-demo"
TAPE="$WORK/tui.tape"
RAW="$WORK/raw.gif"
OUT="$REPO/assets/tui.gif"

for t in vhs gifsicle go; do
  command -v "$t" >/dev/null || { echo "missing required tool: $t" >&2; exit 1; }
done

# --- build the real binary onto $WORK/bin so the recording shows a clean
#     `genv tui` on PATH (no wrapper script) ---
mkdir -p "$WORK/bin"
go -C "$REPO" build -o "$WORK/bin/genv" ./cmd/genv
GENV="$WORK/bin/genv"

# --- build a realistic demo repo (plaintext vault ⇒ no passphrase modal) ---
mkdir -p "$WORK"
cd "$WORK"
$GENV init --encrypt=false >/dev/null

$GENV consumer add api    --strategy single --base-dir apps/api    --filename .env >/dev/null
$GENV consumer add web    --strategy single --base-dir apps/web    --filename .env >/dev/null
$GENV consumer add worker --strategy single --base-dir apps/worker --filename .env >/dev/null

$GENV group add database      --title "Database"        >/dev/null
$GENV group add auth          --title "Auth & Security" >/dev/null
$GENV group add payments      --title "Payments"        >/dev/null
$GENV group add observability --title "Observability"   >/dev/null

dws() { # name group secret|plain consumers value
  if [ "$3" = "secret" ]; then $GENV var define "$1" --group "$2" --secret >/dev/null
  else $GENV var define "$1" --group "$2" >/dev/null; fi
  $GENV wire "$1" --vault local --consumers "$4" --shared >/dev/null
  printf '%s' "$5" | $GENV set "$1" --vault local >/dev/null
}

dws DATABASE_URL          database      secret api,worker     "postgres://app@db.internal:5432/app"
dws REDIS_URL             database      plain  api,worker     "redis://cache.internal:6379/0"
dws JWT_SECRET            auth          secret api            "sk_jwt_8f2b1c9d4e"
dws SESSION_SECRET        auth          secret api,web        "sess_2a9f7e1b"
dws STRIPE_SECRET_KEY     payments      secret api,worker     "sk_live_51MxQ2"
dws STRIPE_WEBHOOK_SECRET payments      secret api            "whsec_3bXk9"
dws SENTRY_DSN            observability plain  api,web,worker "https://abc@o12.ingest.sentry.io/456"
dws LOG_LEVEL            observability plain  api,web,worker "info"

$GENV global define NODE_ENV --vault local --value production >/dev/null
$GENV global define PORT     --vault local --runtime          >/dev/null

# --- the screenplay ---
cat > "$TAPE" <<EOF
# genv TUI demo — keyboard-first control over the whole environment.
Output "$RAW"

Set Shell "bash"
Set FontSize 16
Set Width 1360
Set Height 540
Set Padding 18
Set Theme "Catppuccin Mocha"
Set CursorBlink false
Set TypingSpeed 55ms

# --- setup (hidden from the recording) ---
Hide
Type "export PATH=$WORK/bin:\$PATH"
Enter
Type "cd $WORK"
Enter
Type "clear"
Enter
Show

# --- scene ---
Type "genv tui"
Enter
Sleep 3.5s

# walk the variable list — the inspector's wiring matrix updates live
Type "j"
Sleep 700ms
Type "j"
Sleep 700ms
Type "j"
Sleep 700ms
Type "j"
Sleep 900ms

# the main list is tabbed: variables -> globals (static/runtime) -> groups
Type "]"
Sleep 1.6s
Type "]"
Sleep 1s

# back to variables
Type "["
Sleep 300ms
Type "["
Sleep 900ms

# every mutation shows its plan first — open the generate plan
Type "g"
Sleep 3s
Escape
Sleep 1s
EOF

# --- render + lossless optimize ---
mkdir -p "$REPO/assets"
echo "rendering $TAPE ..."
vhs "$TAPE"
echo "optimizing -> $OUT"
gifsicle -O3 "$RAW" -o "$OUT"
echo "done: $OUT ($(du -h "$OUT" | cut -f1))"
