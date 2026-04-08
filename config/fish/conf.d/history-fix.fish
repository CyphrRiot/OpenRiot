# Fix corrupted fish history on shell startup
# Common issue from unclean shutdowns or concurrent writes
set -l hist_file "$HOME/.local/share/fish/fish_history"

if not test -f "$hist_file"
    exit 0
end

# Try merge first — fish handles most corruption internally
history merge 2>/dev/null
if test $status -eq 0
    exit 0
end

# History is corrupted — try to salvage good entries
set -l tmp_file (mktemp)
python3 -c "
import sys, os
path = os.path.expanduser('$hist_file')
with open(path, 'r') as f:
    lines = f.readlines()
out = []
skip = False
for line in lines:
    if skip:
        if line.startswith('- cmd:'):
            skip = False
            out.append(line)
        continue
    if line.startswith('- cmd:'):
        out.append(line)
    elif line.startswith('  '):
        if out:
            out.append(line)
    else:
        skip = True
if out:
    with open('$tmp_file', 'w') as f:
        f.writelines(out)
" 2>/dev/null

if test -s "$tmp_file"
    cp "$tmp_file" "$hist_file"
end
rm -f "$tmp_file"
