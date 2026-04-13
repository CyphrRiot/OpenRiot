#!/bin/sh
# OpenRiot - Workspace indicator for polybar
# Shows workspace number with highlight if focused

WORKSPACE="${1:-1}"
FOCUS=$(i3-msg -t get_workspaces | jq -r ".[] | select(.num == $WORKSPACE) | .focused")
if [ "$FOCUS" = "true" ]; then
  echo "%{u#bb9af7}%{+u} $WORKSPACE"
else
  echo " $WORKSPACE"
fi
