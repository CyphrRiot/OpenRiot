# =============================================================================
# OpenRiot Fish Shell Configuration
# Ported from OpenRiot for OpenBSD
# =============================================================================

# Greeting with fastfetch if available
function fish_greeting
    if command -v fastfetch >/dev/null 2>&1
        sleep 0.1
        command fastfetch --logo-width 20 --logo openbsd
    end
end

# =============================================================================
# Path Configuration
# =============================================================================

# Add local bin directories to PATH
fish_add_path --prepend $HOME/.local/share/openriot/config/bin
fish_add_path --prepend $HOME/.local/bin
fish_add_path --prepend $HOME/.local/bin/openriot

# OpenBSD paths
fish_add_path --prepend /usr/local/bin
fish_add_path --prepend /usr/X11R6/bin

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
    set_color purple
    echo -n "λ "
    set_color normal
    set_color blue
    printf "%s" (string replace $HOME "~" (pwd))
    set_color normal
    printf "%s" (__fish_git_prompt)
    if test $last_status -ne 0
        set_color red
        printf " [%d]" $last_status
        set_color normal
    end
    set_color cyan
    printf " ➤ "
    set_color normal
end

# Right-hand Prompt Function
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

# Vim/Vi aliases to nvim
alias vim='nvim'
alias vi='nvim'

# Directory listing with lsd (better ls)
alias ls='lsd'
alias ll='lsd -l'
alias la='lsd -la'

# Fastfetch with correct logo width
alias fastfetch='command fastfetch --logo-width 20'

# OpenBSD-specific aliases
alias doas='doas'

# Disk usage - show top 10 largest items by size
function dum
    du -sm * | sort -nr | head -10
end

# =============================================================================
# Sway Auto-start (on login TTY)
# =============================================================================

# Auto-start Sway on TTY1
if status is-login && test (tty) = /dev/tty1
    if type -q sway
        exec sway
    else
        echo "Sway not found. Run 'setup.sh' or install sway."
    end
end

# =============================================================================
# OpenRouter LLM Configuration
# =============================================================================

# OpenRouter API key for Neovim plugins (Avante, CodeCompanion)
# Get your free key from https://openrouter.ai/settings
# NOTE: Replace "YOUR_OPENROUTER_API_KEY" with your actual key after install
set -gx OPENROUTER_API_KEY "YOUR_OPENROUTER_API_KEY"
set -gx OPENROUTER_BASE_URL "https://openrouter.ai/api/v1"
