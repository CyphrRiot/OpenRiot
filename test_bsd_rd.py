#!/usr/bin/env python3
"""
Explore bsd.rd structure to understand how to inject auto_install.conf
"""

import gzip
import os

BSD_RD = "/home/grendel/Code/OpenRiot/.work/iso_contents/7.9/amd64/bsd.rd"
WORK_DIR = "/home/grendel/Code/OpenRiot/.work/rd_extract"


def extract_bsd_rd():
    """Decompress bsd.rd and save for analysis"""
    os.makedirs(WORK_DIR, exist_ok=True)

    with gzip.open(BSD_RD, "rb") as f:
        data = f.read()

    out_path = os.path.join(WORK_DIR, "bsd.rd.decompressed")
    with open(out_path, "wb") as f:
        f.write(data)

    print(f"Decompressed bsd.rd: {len(data)} bytes ({len(data) / 1024 / 1024:.1f} MB)")
    print(f"Saved to: {out_path}")
    return data


def find_ufs_magic(data):
    """Find UFS superblock magic numbers"""
    print("\n=== Searching for UFS superblock magic ===")
    patterns = [
        (b"\x00\x19\x19\x19", "UFS2 magic (big-endian)"),
        (b"\x19\x00\x00\x00", "UFS2 alt magic"),
    ]

    for pattern, desc in patterns:
        offsets = []
        start = 0
        while True:
            idx = data.find(pattern, start)
            if idx == -1:
                break
            offsets.append(idx)
            start = idx + 1
            if len(offsets) > 10:
                break

        if offsets:
            print(f"  {desc}: found at {offsets[:5]}")

    # Check standard UFS2 superblock offset (65536 = 128 * 512)
    for offset in [65536, 131072, 196608]:
        if offset < len(data):
            chunk = data[offset : offset + 16]
            print(f"  Offset {offset}: {chunk.hex()}")


def find_strings(data, min_len=4):
    """Find interesting strings in the data"""
    print("\n=== Searching for path strings ===")

    strings = []
    current = b""

    for byte in data[: min(len(data), 2000000)]:
        if 32 <= byte <= 126:
            current += bytes([byte])
        else:
            if len(current) >= min_len:
                strings.append(current)
            current = b""

    interesting = []
    for s in strings:
        try:
            decoded = s.decode("ascii")
            if any(
                decoded.startswith(p)
                for p in ["/etc/", "/bin/", "/sbin/", "/usr/", "/root/", "/var/"]
            ):
                interesting.append(decoded)
        except:
            pass

    print(f"  Found {len(interesting)} path strings")
    for s in interesting[:20]:
        print(f"    {s}")


def main():
    if not os.path.exists(BSD_RD):
        print(f"ERROR: {BSD_RD} not found")
        return

    print(f"=== Exploring bsd.rd ===")
    print(f"File: {BSD_RD}")
    print(f"Size: {os.path.getsize(BSD_RD)} bytes")

    decompressed_path = os.path.join(WORK_DIR, "bsd.rd.decompressed")
    if os.path.exists(decompressed_path):
        print("Using existing decompressed file...")
        with open(decompressed_path, "rb") as f:
            data = f.read()
    else:
        data = extract_bsd_rd()

    print(f"\nDecompressed size: {len(data)} bytes")

    find_ufs_magic(data)
    find_strings(data)

    print("\n=== Done ===")


if __name__ == "__main__":
    main()
