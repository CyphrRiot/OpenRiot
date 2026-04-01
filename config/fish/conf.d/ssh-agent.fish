# =============================================================================
# SSH Agent Auto-start (OpenBSD / Fish Shell)
# =============================================================================
# Ensures a single ssh-agent per login session.
# Reuses environment from ~/.cache/ssh-agent.env when available.
# Starts a new agent only if the socket is invalid.
# Compatible with OpenBSD and Fish shell.

# Only run in interactive shells
if not status is-interactive
    exit
end

# Require ssh-agent and ssh-add
if not command -q ssh-agent
    exit
end
if not command -q ssh-add
    exit
end

set -l envfile ~/.cache/ssh-agent.env

function __ssh_agent_save_env
    mkdir -p (dirname $envfile)
    printf 'set -gx SSH_AUTH_SOCK %s\nset -gx SSH_AGENT_PID %s\n' $SSH_AUTH_SOCK $SSH_AGENT_PID > $envfile
end

function __ssh_agent_start
    set -l out (ssh-agent -s)
    set -l sock (string match -r 'SSH_AUTH_SOCK=([^;]+)' $out | string replace -r '.*SSH_AUTH_SOCK=([^;]+).*' '$1')
    set -l pid  (string match -r 'SSH_AGENT_PID=([0-9]+)' $out | string replace -r '.*SSH_AGENT_PID=([0-9]+).*' '$1')

    if test -n "$sock" -a -n "$pid"
        set -gx SSH_AUTH_SOCK $sock
        set -gx SSH_AGENT_PID $pid
        __ssh_agent_save_env
        return 0
    end
    return 1
end

set -l have_valid 0

# Case 1: Current env is already valid
if set -q SSH_AUTH_SOCK; and test -S "$SSH_AUTH_SOCK"
    set have_valid 1
else
    # Case 2: Try previously saved env
    if test -f $envfile
        source $envfile 2>/dev/null
        if set -q SSH_AUTH_SOCK; and test -S "$SSH_AUTH_SOCK"
            set have_valid 1
        end
    end
end

# Case 3: Start a new agent if still invalid (login shells only)
if test $have_valid -eq 0
    if status is-login
        __ssh_agent_start >/dev/null 2>/dev/null
        if set -q SSH_AUTH_SOCK; and test -S "$SSH_AUTH_SOCK"
            set have_valid 1
        end
    end
end

# Add default keys if agent is valid and no identities loaded
if test $have_valid -eq 1
    ssh-add -l >/dev/null 2>/dev/null
    if test $status -ne 0
        for k in id_ed25519 id_rsa id_ecdsa
            if test -f ~/.ssh/$k
                ssh-add ~/.ssh/$k >/dev/null 2>/dev/null
            end
        end
    end
end
