#!/bin/sh
# OpenRiot — weather-emoji-plain.sh
# Plain-text weather (Temp + emoji) for Waybar using `stormy`
#
# - Reads optional config for location:
#     $HOME/.config/waybar/weather.conf   (LOCATION="City, CC")
#     $HOME/.config/openriot/weather.conf (LOCATION="City, CC")
#   If LOCATION is unset/empty, stormy will auto-detect by IP.
# - Requires: stormy (pkg_add stormy)
# - Respects weather toggle: if $HOME/.config/openriot/disable-weather exists, prints nothing.
# - Output: a single line, e.g., "24°C ☀︎" (or "" on failure/missing deps)
#
# OpenBSD compatible (POSIX sh, no bash-isms)

set -eu

# Toggle: allow users to disable weather globally
if [ -f "$HOME/.config/openriot/disable-weather" ]; then
  echo ""
  exit 0
fi

have() { command -v "$1" >/dev/null 2>&1; }

# Require stormy; print nothing if unavailable
if ! have stormy; then
  echo ""
  exit 0
fi

# Load config (LOCATION variable) if present
# Prefer Waybar config, fallback to ArchRiot config
CFG=""
if [ -f "$HOME/.config/waybar/weather.conf" ]; then
  CFG="$HOME/.config/waybar/weather.conf"
elif [ -f "$HOME/.config/openriot/weather.conf" ]; then
  CFG="$HOME/.config/openriot/weather.conf"
fi

LOCATION=""
if [ -n "$CFG" ]; then
  # shellcheck disable=SC1090
  . "$CFG" 2>/dev/null || true
  # Support both LOCATION= and export LOCATION=
  LOCATION="${LOCATION:-}"
fi

# Query stormy in "simple" mode; suppress color/tty tricks
OUT_RAW=""
if [ -n "$LOCATION" ]; then
  OUT_RAW="$(NO_COLOR=1 TERM=dumb stormy simple "$LOCATION" 2>/dev/null || true)"
else
  OUT_RAW="$(NO_COLOR=1 TERM=dumb stormy simple 2>/dev/null || true)"
fi

# If stormy failed or returned nothing, print nothing (no error spam)
if [ -z "$OUT_RAW" ]; then
  echo ""
  exit 0
fi

# Strip ANSI escapes (best-effort)
OUT_STRIPPED="$(printf "%s\n" "$OUT_RAW" | sed -E 's/\x1B\[[0-9;]*[A-Za-z]//g')"

# Extract Weather and Temp lines (stormy "simple" typically prints these)
COND="$(printf "%s\n" "$OUT_STRIPPED" | awk '/Weather/{sub(/^.*Weather[[:space:]]+/, ""); print; exit}')"
TEMP="$(printf "%s\n" "$OUT_STRIPPED" | awk '/Temp/{sub(/^.*Temp[[:space:]]+/, ""); print; exit}')"

# Trim whitespace
trim() { printf "%s" "$1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'; }
COND="$(trim "$COND")"
TEMP="$(trim "$TEMP")"

# Fallbacks if stormy format changes
[ -z "$COND" ] && COND="Weather"
[ -z "$TEMP" ] && TEMP="--"

# Emoji mapping from condition text
lw="$(printf "%s" "$COND" | tr '[:upper:]' '[:lower:]')"
EMO="🌡"
case "$lw" in
  *clear*|*sun*) EMO="☀︎" ;;
  *cloud*|*overcast*) EMO="☁︎" ;;
  *rain*|*drizzle*|*shower*) EMO="☔︎" ;;
  *snow*) EMO="❄︎" ;;
  *thunder*|*storm*) EMO="⚡︎" ;;
  *fog*|*mist*|*haze*) EMO="〰" ;;
  *wind*) EMO="⚑" ;;
  *) EMO="🌡" ;;
esac

# Print a single, quiet line (Hyprlock/Waybar friendly)
printf "%s %s\n" "$TEMP" "$EMO"
