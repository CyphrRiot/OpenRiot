#!/usr/bin/env python3
"""
OpenRiot Swaylock Background Generator
Generates a lock screen image with time, date, username, hostname, battery, crypto, and background.
Works standalone on OpenBSD — no OpenRiot binary dependency.
"""

import json
import os
import subprocess
import time

from PIL import Image, ImageDraw, ImageFont

BACKGROUND_DIR = os.path.expanduser("~/.local/share/openriot/backgrounds")
DEFAULT_BG = os.path.join(BACKGROUND_DIR, "riot_01.jpg")
OUTPUT = "/tmp/swaylock-bg.png"
W, H = 1920, 1080

BLUE = (125, 166, 255)
PURPLE = (138, 149, 232)
LGRAY = (160, 200, 250)
DGRAY = (8, 9, 12)


def font(size):
    paths = [
        "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/usr/local/share/fonts/dejavu/DejaVuSansMono.ttf",
    ]
    for p in paths:
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()


def find_background():
    if os.path.exists(DEFAULT_BG):
        return DEFAULT_BG
    for i in range(1, 14):
        path = os.path.join(BACKGROUND_DIR, f"riot_{i:02d}.jpg")
        if os.path.exists(path):
            return path
    return None


def main():
    bg = find_background()
    if bg:
        img = Image.open(bg).convert("RGBA").resize((W, H), Image.LANCZOS)
    else:
        img = Image.new("RGBA", (W, H), DGRAY)

    overlay = Image.new("RGBA", (W, H), (*DGRAY, 160))
    img = Image.alpha_composite(img, overlay)
    d = ImageDraw.Draw(img)

    FLG = font(120)
    FMD = font(28)
    FSM = font(18)

    # Time — centered
    t = subprocess.check_output(["date", "+%I:%M %p"], text=True).strip()
    bb = d.textbbox((0, 0), t, font=FLG)
    tw = bb[2] - bb[0]
    d.text(((W - tw) // 2, H // 2 - 160), t, fill=BLUE, font=FLG)

    # Date — below time
    dt = subprocess.check_output(["date", "+%A %B %-d, %Y"], text=True).strip()
    bb = d.textbbox((0, 0), dt, font=FMD)
    dw = bb[2] - bb[0]
    d.text(((W - dw) // 2, H // 2 - 50), dt, fill=BLUE, font=FMD)

    # Username — bottom right
    u = os.getenv("USER", "user")
    bb = d.textbbox((0, 0), u, font=FSM)
    uw = bb[2] - bb[0]
    d.text((W - uw - 30, H - (bb[3] - bb[1]) - 40), u, fill=PURPLE, font=FSM)

    # Hostname — bottom left
    h = subprocess.check_output(["uname", "-n"], text=True).strip()
    bb = d.textbbox((0, 0), h, font=FSM)
    hh = bb[3] - bb[1]
    d.text((30, H - hh - 40), h, fill=PURPLE, font=FSM)

    # Battery status — bottom center
    try:
        result = subprocess.run(
            ["apm", "-l"], capture_output=True, text=True, timeout=2
        )
        if result.returncode == 0:
            percent = result.stdout.strip()
            ac_result = subprocess.run(
                ["apm", "-a"], capture_output=True, text=True, timeout=2
            )
            charging = (
                ac_result.stdout.strip() == "1" if ac_result.returncode == 0 else False
            )
            icon = "⚡" if charging else "🔋"
            bat_text = f"{icon} {percent}%"
            bb = d.textbbox((0, 0), bat_text, font=FSM)
            bw = bb[2] - bb[0]
            d.text(
                ((W - bw) // 2, H - (bb[3] - bb[1]) - 40),
                bat_text,
                fill=LGRAY,
                font=FSM,
            )
    except (
        subprocess.TimeoutExpired,
        FileNotFoundError,
        subprocess.CalledProcessError,
    ):
        pass  # apm not available (desktop machine) — skip silently

    # Crypto price (BTC) — top right
    crypto_text = ""
    cache_file = "/tmp/openriot-crypto-cache.json"
    try:
        # Check cache first
        if os.path.exists(cache_file):
            mtime = os.path.getmtime(cache_file)
            if time.time() - mtime < 300:  # 5 minute TTL
                with open(cache_file) as f:
                    data = json.load(f)
                    btc = data.get("bitcoin", {}).get("usd")
                    if btc:
                        crypto_text = f"₿ ${btc:,.0f}"

        # Fetch fresh if no cache
        if not crypto_text:
            result = subprocess.run(
                [
                    "curl",
                    "-s",
                    "--max-time",
                    "5",
                    "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd",
                ],
                capture_output=True,
                text=True,
                timeout=6,
            )
            if result.returncode == 0:
                data = json.loads(result.stdout)
                btc = data.get("bitcoin", {}).get("usd")
                if btc:
                    crypto_text = f"₿ ${btc:,.0f}"
                    # Cache it
                    with open(cache_file, "w") as f:
                        json.dump(data, f)
    except (
        subprocess.TimeoutExpired,
        FileNotFoundError,
        json.JSONDecodeError,
        subprocess.CalledProcessError,
    ):
        pass  # Network error or curl not found — skip silently

    if crypto_text:
        bb = d.textbbox((0, 0), crypto_text, font=FSM)
        bw = bb[2] - bb[0]
        d.text((W - bw - 30, 30), crypto_text, fill=LGRAY, font=FSM)

    # Save
    out = Image.new("RGB", (W, H), DGRAY)
    if img.mode == "RGBA":
        out.paste(img, mask=img.split()[3])
    else:
        out.paste(img)
    out.save(OUTPUT, "PNG")
    print(f"Done: {OUTPUT}")


if __name__ == "__main__":
    main()
