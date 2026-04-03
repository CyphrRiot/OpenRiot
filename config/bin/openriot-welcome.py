#!/usr/bin/env python3
"""
OpenRiot Welcome Screen - GTK3 based welcome dialog for OpenBSD/Sway
"""

import os
import sys

try:
    import gi

    gi.require_version("Gtk", "3.0")
    gi.require_version("Gdk", "3.0")
    from gi.repository import Gdk, GLib, Gtk
except ImportError:
    print("GTK3 not available, falling back to terminal")
    sys.exit(1)


class WelcomeWindow:
    def __init__(self):
        self.home_dir = os.path.expanduser("~")
        self.welcomed_file = os.path.join(self.home_dir, ".openriot-welcomed")

        # Check if already welcomed
        if os.path.exists(self.welcomed_file):
            sys.exit(0)

        self.window = Gtk.Window()
        self.window.set_title("Welcome to OpenRiot")
        self.window.set_default_size(700, 500)
        self.window.set_position(Gtk.WindowPosition.CENTER)
        self.window.set_resizable(False)
        self.window.set_modal(False)
        self.window.set_keep_above(False)
        self.window.set_type_hint(Gdk.WindowTypeHint.NORMAL)

        # Set dark background
        self.window.modify_bg(Gtk.StateType.NORMAL, Gdk.Color(0, 0, 0))

        # Main container
        main_box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=20)
        main_box.set_margin_start(40)
        main_box.set_margin_end(40)
        main_box.set_margin_top(40)
        main_box.set_margin_bottom(30)

        # Title
        title = Gtk.Label()
        title.set_markup(
            '<span size="32000" weight="bold" color="#89b4fa">Welcome to OpenRiot!</span>'
        )
        title.set_halign(Gtk.Align.CENTER)
        main_box.pack_start(title, False, False, 10)

        # Version
        version_file = os.path.join(
            os.path.expanduser("~"), ".local/share/openriot/VERSION"
        )
        version_str = (
            open(version_file).read().strip() if os.path.exists(version_file) else "0.7"
        )
        version = Gtk.Label()
        version.set_markup(
            '<span size="14000" color="#a6adc8">v{} on OpenBSD 7.9</span>'.format(
                version_str
            )
        )
        version.set_halign(Gtk.Align.CENTER)
        main_box.pack_start(version, False, False, 0)

        # Separator
        sep = Gtk.Separator(Gtk.Orientation.HORIZONTAL)
        main_box.pack_start(sep, False, False, 20)

        # Quick Start section
        quick_start = Gtk.Label()
        quick_start.set_markup(
            '<span size="16000" weight="bold" color="#cba6f7">Quick Start</span>\n\n'
            '<span size="12000" color="#cdd6f4">'
            "• Super + D       → Open fuzzel app launcher\n"
            "• Super + Enter   → Open terminal (foot)\n"
            "• Super + Q       → Close focused window\n"
            "• Super + 1-9     → Switch workspaces\n"
            "• Super + Shift + Q → Quit Sway\n"
            "</span>"
        )
        quick_start.set_halign(Gtk.Align.START)
        quick_start.set_line_wrap(True)
        main_box.pack_start(quick_start, False, False, 10)

        # Separator
        sep2 = Gtk.Separator(Gtk.Orientation.HORIZONTAL)
        main_box.pack_start(sep2, False, False, 20)

        # Links
        links = Gtk.Label()
        links.set_markup(
            '<span size="12000" color="#89b4fa">'
            "📖 Documentation: https://openriot.org\n"
            "🐛 Issues: https://github.com/CyphrRiot/OpenRiot/issues"
            "</span>"
        )
        links.set_halign(Gtk.Align.CENTER)
        main_box.pack_start(links, False, False, 10)

        # Buttons
        button_box = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=20)
        button_box.set_halign(Gtk.Align.CENTER)

        gotit_btn = Gtk.Button(label="Got it!")
        gotit_btn.set_size_request(120, 40)
        gotit_btn.connect("clicked", self.on_gotit_clicked)

        button_box.pack_start(gotit_btn, False, False, 0)
        main_box.pack_start(button_box, False, False, 20)

        self.window.add(main_box)
        self.window.connect("destroy", self.on_gotit_clicked)
        self.window.show_all()

    def on_gotit_clicked(self, widget=None):
        # Mark as welcomed
        try:
            with open(self.welcomed_file, "w") as f:
                f.write("1")
        except Exception:
            pass
        Gtk.main_quit()


def main():
    WelcomeWindow()
    Gtk.main()


if __name__ == "__main__":
    main()
