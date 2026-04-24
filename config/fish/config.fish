# =============================================================================
# OpenRiot Fish Shell Configuration
# Ported from OpenRiot for OpenBSD
# =============================================================================

# Set UTF-8 locale for foot/X11 compatibility
set -gx LANG en_US.UTF-8
set -gx LC_ALL en_US.UTF-8

# XDG directories — ensure history persists
set -gx XDG_DATA_HOME $HOME/.local/share
if not test -d $XDG_DATA_HOME/fish
    mkdir -p $XDG_DATA_HOME/fish
end

# XDG_RUNTIME_DIR — required by X11
set -gx XDG_RUNTIME_DIR /tmp/$USER-runtime
if not test -d $XDG_RUNTIME_DIR
    mkdir -p $XDG_RUNTIME_DIR
end

# No greeting - use fastfetch instead
set -g fish_greeting ""

# Show motd then fastfetch (interactive shells only - skip SSH/rsync)
if status is-interactive
    if test -f $HOME/.local/share/openriot/install/motd
        cat $HOME/.local/share/openriot/install/motd
    end
    fastfetch
end

# =============================================================================
# Path Configuration
# =============================================================================

# Add OpenRiot scripts to PATH
fish_add_path --prepend $HOME/.local/bin
fish_add_path --prepend $HOME/.local/share/openriot/config/bin

# Add OpenRiot binary to PATH (for CLI commands)
fish_add_path --prepend $HOME/.local/share/openriot/install

# Go workspace
set -gx GOPATH "$HOME/.go"
set -gx GOMODCACHE "$HOME/.go/pkg/mod"
fish_add_path $GOPATH/bin

# OpenBSD paths
fish_add_path --prepend /usr/local/bin
fish_add_path --prepend /usr/bin
fish_add_path --prepend /bin
fish_add_path --prepend /usr/sbin
fish_add_path --prepend /sbin

# =============================================================================
# Git Prompt Configuration
# =============================================================================

set -g __fish_git_prompt_showdirtystate yes
set -g __fish_git_prompt_showstashstate yes
set -g __fish_git_prompt_showuntrackedfiles yes
set -g __fish_git_prompt_showupstream yes
set -g __fish_git_prompt_color_branch 9d7cd8
set -g __fish_git_prompt_color_upstream_ahead purple
set -g __fish_git_prompt_color_upstream_behind red
set -g __fish_git_prompt_color_dirtystate 565f89
set -g __fish_git_prompt_char_dirtystate "●"
set -g __fish_git_prompt_char_stagedstate "→"
set -g __fish_git_prompt_char_untrackedfiles "☡"
set -g __fish_git_prompt_char_stashstate "↩"
set -g __fish_git_prompt_char_upstream_ahead "+"
set -g __fish_git_prompt_char_upstream_behind -
set -g __fish_git_prompt_char_upstream_equal ""

# =============================================================================
# Enhanced Prompt Function
# =============================================================================

function fish_prompt
    set -l last_status $status
    set -l host (hostname -s 2>/dev/null || echo "unknown")
    set_color bb9af7
    echo -n '  '
    set_color 705e97
    printf "%s " $host
    set_color 7dcfff
    printf "%s" (string replace $HOME "~" (pwd))
    if command git rev-parse --git-dir >/dev/null 2>&1
        printf "%s" (__fish_git_prompt)
    end
    if test $last_status -ne 0
        set_color red
        printf " [%d]" $last_status
        set_color normal
    end
    set_color bb9af7
    printf " ❯ "
    set_color normal
end

# NOTE: i3 autostart is handled in the i3 config itself, NOT here.
# fish is a login shell — it should NEVER exec i3, or SSH sessions die.
# On TTY1, xenodm launches ~/.xsession which starts i3. On SSH, you get a shell.
function fish_right_prompt
    # Disabled - uncomment to enable
    # printf "%s" (date "+%H:%M:%S")
end

# =============================================================================
# Aliases & Functions
# =============================================================================

# Vim/Vi aliases to helix (only if vim is not installed)
if not type -q vim
    alias vim='hx'
    alias vi='hx'
end

# Helix alias
alias helix='hx'

# Signal Messenger (gurk)
alias signal '~/.local/share/openriot/config/bin/gurk'

# Directory listing with lsd (better ls)
alias ls='lsd'
alias ll='lsd -l'
alias la='lsd -la'

# more sucks without this
alias more='more -e'

# Disk usage - show top 10 largest items by size
function dum
    du -sm * | sort -nr | head -10
end

# =============================================================================
# History Configuration
# =============================================================================
set -x FISH_HISTORY_SAVE 1
set -x FISH_HISTORY_LIMIT 5000

# NOTE: Don't manually save history on every command - fish auto-saves.
# Manual save causes race conditions and file corruption.
# function __save_history --on-event fish_postexec
#     history save
# end

# =============================================================================
# Abbreviations & Aliases
# =============================================================================

# OpenRiot path - type "or" then space to expand
abbr -a -g or --position anywhere --set-cursor='%' ~/.local/share/openriot/

# Create the function properly
abbr -a -g !! --position anywhere 'commandline -t -- (history | head -n 1)'

# Ctrl+O for !! (escape doesn't work well with !!)
bind \co 'commandline -i (history | head -1)'

# Signal Messenger
alias signal '~/.local/share/openriot/config/bin/gurk'

# Fix Firefux
set -x MOZ_ENABLE_WAYLAND 0 # Force X11 (sometimes helps clipboard)
set -x MOZ_DISABLE_CONTENT_SANDBOX 1
set -x MOZ_USE_XINPUT2 0
