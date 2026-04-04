#!/bin/sh
# OpenRiot QEMU test script
# Creates a virtual disk and boots the ISO for testing OpenBSD install

set -e

# Config
ISO_PATH="./isos/openriot.iso"
VM_NAME="openriot"
MEMORY="2048"
DISK_SIZE="5G"
DISK_PATH="${HOME}/.cache/openriot-test.qcow2"

# Create virtual disk if missing
if [ ! -f "$DISK_PATH" ]; then
    echo "Creating virtual disk: $DISK_PATH ($DISK_SIZE)"
    qemu-img create -f qcow2 "$DISK_PATH" "$DISK_SIZE"
fi

echo "Starting QEMU with ${MEMORY}MB RAM..."
echo "Boot order: dc (disk first, then cdrom)"
echo ""
echo "At the OpenBSD prompt:"
echo "  Press 'i' for interactive install"
echo "  Press 'a' for autoinstall"
echo ""

# QEMU for OpenBSD:
# - sd0 = first hard disk (virtio)
# - cd0 = first CD-ROM (ide-cd)
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
    -drive file="$ISO_PATH",id=cd0,if=ide,media=cdrom \
    -boot order=dc \
    -rtc base=utc,clock=host \
    -usb \
    -device usb-tablet
