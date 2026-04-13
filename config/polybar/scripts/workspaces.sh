#!/bin/sh
# Polybar i3 workspaces with app icons - OpenRiot
# Format: ● 󰊠 󰈹
# Usage: workspaces.sh [workspace_num]

# Get window classes for a workspace
get_window_classes() {
    ws_num="$1"
    i3-msg -t get_tree 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
except:
    sys.exit(0)

def find_windows(node, ws_num):
    if node.get('type') == 'workspace' and node.get('num') == ws_num:
        return collect_windows(node)
    results = []
    for n in node.get('nodes', []) + node.get('floating_nodes', []):
        results.extend(find_windows(n, ws_num))
    return results

def collect_windows(node):
    wins = []
    if node.get('window'):
        cp = node.get('window_properties', {})
        cls = cp.get('class', '')
        if cls:
            wins.append(cls)
    for c in node.get('nodes', []) + node.get('floating_nodes', []):
        wins.extend(collect_windows(c))
    return wins

for n in data.get('nodes', []) + data.get('floating_nodes', []):
    for cls in find_windows(n, $ws_num):
        print(cls)
" 2>/dev/null
}

# Map window class to icon
map_icon() {
    class="$1"
    ~/.local/share/openriot/install/openriot --window-icon "$class" 2>/dev/null
}

# Get workspace state
get_state() {
    ws_num="$1"
    workspaces="$(i3-msg -t get_workspaces 2>/dev/null)"
    
    # Extract focused and urgent for this workspace
    ws_line="$(echo "$workspaces" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    for ws in data:
        if ws.get('num') == $ws_num:
            focused = ws.get('focused', False)
            urgent = ws.get('urgent', False)
            print('focused' if focused else 'unfocused', 'urgent' if urgent else 'normal')
            break
except:
    pass
" 2>/dev/null)"
    
    state="$(echo "$ws_line" | awk '{print $1}')"
    has_urgent="$(echo "$ws_line" | awk '{print $2}')"
    
    if [ "$has_urgent" = "urgent" ]; then
        echo "urgent"
    elif [ "$state" = "focused" ]; then
        echo "focused"
    else
        echo "unfocused"
    fi
}

# Output one workspace
output_workspace() {
    ws_num="$1"
    state="$(get_state "$ws_num")"
    
    # Get icons for all windows in this workspace
    icons=""
    for cls in $(get_window_classes "$ws_num"); do
        icon=$(map_icon "$cls")
        if [ -n "$icon" ]; then
            if [ -n "$icons" ]; then
                icons="$icons $icon"
            else
                icons="$icon"
            fi
        fi
    done
    
    # Determine indicator based on state AND whether has apps
    if [ "$state" = "focused" ]; then
        indicator=""
    elif [ "$state" = "urgent" ]; then
        indicator=""
    elif [ -n "$icons" ]; then
        # Inactive with apps
        indicator=""
    else
        # Inactive empty
        indicator=""
    fi
    
    # Dim inactive apps to match window-title color (fg3)
    if [ "$state" = "unfocused" ] && [ -n "$icons" ]; then
        icons="%{T0}%{F#565f89}$icons%{F-}%{T-}"
    fi
    
    # Unfocused indicators use window-title color (fg3)
    if [ "$state" = "unfocused" ]; then
        indicator="%{F#565f89}$indicator%{F-}"
    fi
    
    if [ -n "$icons" ]; then
        echo "$indicator $icons"
    else
        echo "$indicator"
    fi
}

# Main
if [ -n "$1" ]; then
    # Single workspace requested
    output_workspace "$1"
else
    # All workspaces
    for i in 1 2 3 4; do
        if [ -z "$output" ]; then
            output="$(output_workspace $i)"
        else
            output="$output   $(output_workspace $i)"
        fi
    done
    echo "${output:-         }"
fi
