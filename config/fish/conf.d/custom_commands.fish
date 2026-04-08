# =============================================================================
# Custom Commands Loader
# =============================================================================
# This file automatically loads custom commands from ~/.config/fish/custom_commands.fish
# if it exists. This follows Fish shell best practices for modular configuration.

# Load custom commands if the file exists
set -l custom_commands_file "$HOME/.config/fish/custom_commands.fish"
if test -f "$custom_commands_file"
    source "$custom_commands_file"
end
