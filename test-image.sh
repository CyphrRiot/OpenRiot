#!/bin/sh
# Boot the installed OpenBSD system from the qcow2 disk
# Does NOT use the ISO - boots directly from the installed system

set -e

# Config
DISK_PATH="${HOME}/.cache/openriot-test.qcow2"
VM_NAME="openriot-installed"
MEMORY="2048"

# Check if disk exists
if [ ! -f "$DISK_PATH" ]; then
    echo "Error: $DISK_PATH not found"
    echo "Run 'make isotest' first to install OpenBSD"
    exit 1
fi

echo "Booting installed OpenBSD from $DISK_PATH"
echo ""
echo "Login as root or your user."
echo "To start OpenRiot desktop: openriot"
echo ""

exec qemu-system-x86_64 \
    -name "$VM_NAME" \
    -machine type=q35,accel=kvm \
    -cpu host \
    -smp 2 \
    -m "$MEMORY" \
    -device virtio-vga-gl \
    -display gtk,gl=on \
    -netdev user,id=net0 \
    -device virtio-net-pci,netdev=net0 \
    -device virtio-blk-pci,drive=hd0 \
    -drive file="$DISK_PATH",format=qcow2,id=hd0,if=none \
    -boot order=c \
    -rtc base=utc,clock=host \
    -usb \
    -device usb-tablet
