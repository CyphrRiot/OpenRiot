# Fix corrupted fish history on shell startup
# Common issue: entries missing blank line separators
set -l hist_file "$HOME/.local/share/fish/fish_history"

if not test -f "$hist_file"
    exit 0
end

# Check if fix is needed — look for missing blank lines between entries
set -l needs_fix (python3 -c "
with open('$hist_file') as f:
    lines = f.readlines()
bad = False
for i, line in enumerate(lines):
    if line.startswith('- cmd:'):
        # Previous non-blank line should be blank or file start
        if i > 0 and lines[i-1].strip() != '':
            bad = True
            break
print('yes' if bad else 'no')
" 2>/dev/null)

if test "$needs_fix" = "no"
    exit 0
end

# Rebuild history with proper blank line separators
set -l tmp_file (mktemp)
python3 -c "
import sys
with open('$hist_file') as f:
    lines = f.readlines()

entries = []
current = []
for line in lines:
    if line.startswith('- cmd:'):
        if current:
            entries.append(current)
        current = [line]
    elif current:
        current.append(line)

if current:
    entries.append(current)

with open('$tmp_file', 'w') as f:
    for entry in entries:
        # Trim trailing blank lines from each entry
        while entry and entry[-1].strip() == '':
            entry.pop()
        if entry:
            f.writelines(entry)
            f.write('\n')
" 2>/dev/null

if test -s "$tmp_file"
    cp "$tmp_file" "$hist_file"
    echo "[fish] History repaired"
end
rm -f "$tmp_file"
