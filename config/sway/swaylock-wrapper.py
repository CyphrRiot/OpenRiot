#!/usr/bin/env python3
"""
OpenRiot Swaylock Background Generator
Generates a lock screen image with time, date, username, hostname, and background.
Works standalone on OpenBSD — no ArchRiot binary dependency.
"""
import os
import subprocess
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
