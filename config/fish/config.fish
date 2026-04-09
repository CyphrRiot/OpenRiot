# =============================================================================
# OpenRiot Fish Shell Configuration
# Ported from OpenRiot for OpenBSD
# =============================================================================

# Set UTF-8 locale for foot/wayland compatibility
set -gx LANG en_US.UTF-8
set -gx LC_ALL en_US.UTF-8

# XDG directories — ensure history persists
set -gx XDG_DATA_HOME $HOME/.local/share
if not test -d $XDG_DATA_HOME/fish
    mkdir -p $XDG_DATA_HOME/fish
end

# XDG_RUNTIME_DIR — required by Wayland compositors
set -gx XDG_RUNTIME_DIR /tmp/$USER-runtime
if not test -d $XDG_RUNTIME_DIR
    mkdir -p $XDG_RUNTIME_DIR
end

# Greeting with fastfetch if available
function fish_greeting
    if command -v fastfetch >/dev/null 2>&1
        # Run fastfetch with timeout to prevent hanging on font/locale issues
        timeout 5 fastfetch --logo-width 20 --logo openbsd_small 2>/dev/null
    end
end

# =============================================================================
# Path Configuration
# =============================================================================

# Add OpenRiot scripts to PATH
fish_add_path --prepend $HOME/.local/share/openriot/config/bin

# OpenBSD paths
fish_add_path --prepend /usr/local/bin

# =============================================================================
# Git Prompt Configuration
# =============================================================================

set -g __fish_git_prompt_showdirtystate yes
set -g __fish_git_prompt_showstashstate yes
set -g __fish_git_prompt_showuntrackedfiles yes
set -g __fish_git_prompt_showupstream yes
set -g __fish_git_prompt_color_branch yellow
set -g __fish_git_prompt_color_upstream_ahead purple
set -g __fish_git_prompt_color_upstream_behind red
set -g __fish_git_prompt_color_dirtystate ff8c00
set -g __fish_git_prompt_char_dirtystate "*"
set -g __fish_git_prompt_char_stagedstate "+"
set -g __fish_git_prompt_char_untrackedfiles "?"
set -g __fish_git_prompt_char_stashstate "^"
set -g __fish_git_prompt_char_upstream_ahead "+"
set -g __fish_git_prompt_char_upstream_behind -
set -g __fish_git_prompt_char_upstream_equal ""

# =============================================================================
# Enhanced Prompt Function
# =============================================================================

function fish_prompt
    set -l last_status $status
    set_color purple
    echo -n '🐡 '
    set_color cyan
    printf "%s" (string replace $HOME "~" (pwd))
    set_color normal
    if test $last_status -ne 0
        set_color red
        printf " [%d]" $last_status
        set_color normal
    end
    set_color cyan
    printf "❯ "
    set_color normal
end

# NOTE: Sway autostart is handled in the Sway config itself, NOT here.
# fish is a login shell — it should NEVER exec sway, or SSH sessions die.
# On TTY1, xenodm launches sway directly. On SSH, you get a shell.
function fish_right_prompt
    set_color purple
    printf "("
    set_color cyan
    printf "%s" (hostname)
    set_color purple
    printf ") "
    set_color blue
    printf "%s" (date "+%H:%M:%S")
    set_color normal
end

# =============================================================================
# Aliases & Functions
# =============================================================================

# Vim/Vi aliases to helix
alias vim='hx'
alias vi='hx'
alias helix='hx'

# Directory listing with lsd (better ls)
alias ls='lsd'
alias ll='lsd -l'
alias la='lsd -la'

# Fastfetch with correct logo width
alias fastfetch='command fastfetch --logo-width 20 --logo openbsd_small'

# OpenBSD-specific aliases
alias doas='doas'

# Disk usage - show top 10 largest items by size
function dum
    du -sm * | sort -nr | head -10
end

# =============================================================================
# OpenRouter LLM Configuration
# =============================================================================

# OpenRouter API key for Neovim plugins (Avante, CodeCompanion)
# Get your free key from https://openrouter.ai/settings
# NOTE: Replace "YOUR_OPENROUTER_API_KEY" with your actual key after install
set -gx OPENROUTER_API_KEY "YOUR_OPENROUTER_API_KEY"
set -gx OPENROUTER_BASE_URL "https://openrouter.ai/api/v1"
