# ssh-agent disabled on OpenBSD (requires dbus-session which is not reliably available)
# Re-enable manually if needed by adding keys to ssh-agent
exit 0

# Original implementation kept for reference (disabled due to fish printf issues on OpenBSD):
# The printf redirection approach had issues with variable expansion timing in fish shell.
