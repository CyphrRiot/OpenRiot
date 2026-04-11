#!/bin/sh
# OpenRiot - Polybar Active Window Title
# Shows title of currently focused window

get_focused_title() {
    i3-msg -t get_tree 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
except:
    sys.exit(0)

def find_focused(node):
    if node.get('focused', False) and node.get('type') == 'con' and node.get('window'):
        return node.get('name', '')
    for n in node.get('nodes', []) + node.get('floating_nodes', []):
        result = find_focused(n)
        if result:
            return result
    return ''

# Search in focused workspace first
for node in data.get('nodes', []) + data.get('floating_nodes', []):
    if node.get('type') == 'workspace' and node.get('focused', False):
        title = find_focused(node)
        if title:
            # Truncate to 50 chars
            if len(title) > 50:
                title = title[:47] + '...'
            print(title)
            sys.exit(0)
        break

# Fallback: find any focused window
title = find_focused(data)
if title:
    if len(title) > 50:
        title = title[:47] + '...'
    print(title)
" 2>/dev/null
}

title="$(get_focused_title)"
if [ -n "$title" ]; then
    printf "  %s" "$title"
fi
