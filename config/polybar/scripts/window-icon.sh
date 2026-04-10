#!/bin/sh
# Polybar i3 workspaces with app icons - OpenRiot
# Format: ● 󰊠 󰈹   ○   ◉ 󰝰   ○

# Map window classes to icons
case "$1" in
    # Terminals
    foot) echo "󰊠" ;;
    alacritty) echo "󰞷" ;;
    kitty) echo "󰄛" ;;
    gnome-terminal|xfce4-terminal|konsole) echo "" ;;

    # Browsers
    firefox|Firefox) echo "󰈹" ;;
    chrome|Chromium) echo "󰊯" ;;
    microsoft-edge) echo "󰇩" ;;

    # File Managers
    thunar|Thunar) echo "󰝰" ;;
    nautilus|Nautilus|nemo) echo "󰝰" ;;
    dolphin|Dolphin) echo "󰝰" ;;

    # Communication
    telegram|[Tt]elegram-desktop) echo "󰘦" ;;
    discord|Discord) echo "" ;;
    signal) echo "󰍡" ;;

    # Development
    code|Code|VSCode) echo "󰨞" ;;
    code-oss|VSCodium) echo "󰨞" ;;
    sublime|subl) echo "󰅳" ;;
    vim|nvim) echo "󰈙" ;;
    jetbrains-idea|pycharm) echo "" ;;

    # Media
    spotify|Spotify) echo "󰽴" ;;
    mpv) echo "󰥠" ;;
    vlc) echo "󰕼" ;;
    celluloid) echo "󰕼" ;;
    obs|OBS) echo "󰑋" ;;
    cider|Cider) echo "󰎆" ;;

    # System/Settings
    gnome-control-center|org.gnome.Settings) echo "󰒓" ;;
    gnome-tweaks|org.gnome.tweaks) echo "󰒓" ;;
    pavucontrol) echo "󱡫" ;;
    btop|htop|btop) echo "󰍹" ;;
    gnome-system-monitor|org.gnome.SystemMonitor) echo "󰍹" ;;
    blueman-manager|blueman) echo "󰂯" ;;
    nm-connection-editor) echo "󰤨" ;;
    gnome-disks|org.gnome.DiskUtility) echo "󰋊" ;;
    gnome-software|org.gnome.Software) echo "󰏗" ;;
    file-roller|org.gnome.FileRoller) echo "󰗄" ;;
    gnome-screenshot) echo "󰄀" ;;
    gnome-calculator|org.gnome.Calculator) echo "󰪚" ;;
    gnome-text-editor|org.gnome.TextEditor) echo "󰷈" ;;
    gnome-maps|org.gnome.Maps) echo "󰟡" ;;
    gnome-weather|org.gnome.Weather) echo "󰖕" ;;
    evolution|Evolution) echo "󰊫" ;;
    thunderbird|[Tt]hunderbird) echo "󰊫" ;;
    geary|org.gnome.Geary) echo "󰊫" ;;
    geary) echo "󰊫" ;;
    betterbird|eu.betterbird.Betterbird) echo "󰊫" ;;

    # Design/Creativity
    gimp) echo "" ;;
    krita) echo "" ;;
    kdenlive) echo "" ;;
    inkscape) echo "" ;;
    blender) echo "󰂫" ;;
    prusa-slicer|prusaslicer|orca-slicer) echo "󰹛" ;;
    prusaslicer) echo "󰹛" ;;
    ultimaker-cura) echo "󰹛" ;;

    # Office/Documents
    libreoffice-writer) echo "󰈙" ;;
    libreoffice-calc) echo "󰈙" ;;
    libreoffice-startcenter) echo "󰏆" ;;
    evince|org.gnome.Evince) echo "󰈦" ;;
    zathura|org.pwmt.zathura) echo "󰈦" ;;
    org.gnome.Papers) echo "󰈦" ;;
    mousepad) echo "󰈙" ;;
    ghostwriter|org.kde.ghostwriter) echo "󰷈" ;;

    # Games/Entertainment
    steam) echo "󰓓" ;;
    lutris) echo "󰓓" ;;
    heroic) echo "󰓓" ;;
    bottles) echo "󰓓" ;;

    # Utilities
    virt-manager) echo "󰢹" ;;
    remmina) echo "󰢹" ;;
    virtualbox|VirtualBox) echo "󰌗" ;;
    vbox) echo "󰌗" ;;
    gnome-contacts|org.gnome.Contacts) echo "󰀉" ;;
    transmission|transmission-gtk) echo "󰇚" ;;
    fragments|de.haeckerfelix.Fragments) echo "󰇚" ;;
    fragments|de.haeckerfelix.Fragments) echo "󰇚" ;;
    baobab|org.gnome.baobab) echo "󰋊" ;;
    baobab|org.gnome.baobab) echo "󰋊" ;;
    gnome-logs|org.gnome.Logs) echo "󰁯" ;;
    gnome-logs) echo "󰁯" ;;
    gnome-power-stats|org.gnome.PowerStats) echo "󰠭" ;;
    gnome-power-stats) echo "󰠭" ;;
    gnome-about|org.gnome.About) echo "󰋖" ;;
    gnome-usage|org.gnome.Usage) echo "󰍹" ;;
    gnome-connections|org.gnome.Connections) echo "󰢹" ;;
    gnome-extensions|org.gnome.Extensions) echo "󰔎" ;;
    gnome-font-viewer|org.gnome.FontViewer) echo "󰛖" ;;
    gnome-font-viewer) echo "󰛖" ;;
    eog|EyeOfGnome) echo "󰋨" ;;
    eog) echo "󰋨" ;;
    shotwell|shotwell) echo "󰄄" ;;
    shotwell) echo "󰄄" ;;
    simple-scan) echo "󰈿" ;;
    cheese) echo "�拍了" ;;
    polari) echo "󰯯" ;;
    Polari) echo "󰯯" ;;
    quadrapassel) echo "󰘫" ;;
    swell-foop) echo "󰘫" ;;
    five-or-more) echo "󰘫" ;;
    iagno) echo "󰘫" ;;
    lightsoff) echo "󰛔" ;;
    gnome-chess) echo "" ;;
    gnome-mahjongg) echo "" ;;
    gnome-mines) echo "" ;;
    gnome-nibbles) echo "" ;;
    gnome-robots) echo "󰫔" ;;
    gnome-taquin) echo "󰘫" ;;
    gnome-tetravex) echo "󰘫" ;;
    tali) echo "🎲" ;;
    hitori) echo "󰌥" ;;
    gnome-sudoku) echo "�超时" ;;
    atom) echo "󰌔" ;;
    Atom) echo "󰌔" ;;
    zoom|Zoom) echo "󰕧" ;;
    totem) echo "󰕧" ;;
    eog) echo "󰋨" ;;
    imv) echo "󰋨" ;;
    feh) echo "󰋨" ;;
    mpv) echo "󰥠" ;;
    smplayer) echo "󰕼" ;;
    strawberry) echo "󰽴" ;;
    lxmusic) echo "󰽴" ;;
    audacious) echo "󰕼" ;;
    rhythmbox) echo "󰝚" ;;
    elisa) echo "󰽴" ;;
    lollypop) echo "󰝚" ;;
    pomodoro|pomodoro) echo "🍅" ;;
    gnome-calendar) echo "󰃭" ;;
    gnome-clocks) echo "󰔗" ;;
    gnome-characters) echo "󰋽" ;;
    gnome-calculator) echo "󰪚" ;;
    gnome-dictionary) echo "󰋽" ;;
    dconf-editor) echo "󰒃" ;;
    gdb) echo "󰟡" ;;
    valgrind) echo "󰟡" ;;
    strace) echo "󰟡" ;;
    gnome-system-log) echo "󰁯" ;;
    system-log) echo "󰁯" ;;
    gnome-usage) echo "󰍹" ;;
    gnome-power-profile-switcher) echo "󰠭" ;;
    gsmartcontrol) echo "󰋊" ;;
    gnome-disk-utility) echo "󰋊" ;;
    gnome-fonts) echo "󰛖" ;;
    d-feet) echo "󰯾" ;;
    devhelp) echo "󰌤" ;;
    ninja) echo "󰜌" ;;
    meson) echo "󰜌" ;;
    cmake) echo "󰜌" ;;
    autogen) echo "󰜌" ;;
    automake) echo "󰜌" ;;
    gnome-subtitles) echo "�繁" ;;
    gnome-sound-recorder) echo "󰏘" ;;
    sound-recorder) echo "󰏘" ;;
    pitivi) echo "󰒒" ;;
    dash-to-dock) echo "󰎀" ;;
    caffeine) echo "󰎌" ;;
    clipboard-indicator) echo "󰐑" ;;
    hide-top-bar) echo "󰒨" ;;
    arc-menu) echo "󰌇" ;;
    openweather) echo "󰖕" ;;
    netspeed) echo "󰤨" ;;
    cpufreq) echo "󰓅" ;;
    audio-switcher) echo "󰕩" ;;
    places-and-devices) echo "󰉹" ;;
    drop-by) echo "󰈅" ;;
    gsconnect) echo "󰟢" ;;
    emoji-clipboard) echo "󰐑" ;;
    unite) echo "󰎟" ;;
    gravity-switcher) echo "󰋜" ;;
    horizontal-workspaces) echo "󰫍" ;;
    auto-move-windows) echo "󰫍" ;;
    workspace-indicator) echo "󰫍" ;;
    vienna) echo "󰊫" ;;
    geary) echo "󰊫" ;;
    mailnag) echo "󰊫" ;;
    gnome-mail) echo "󰊫" ;;
    msmtp) echo "󰊫" ;;
    mailspring) echo "󰊫" ;;
    newton) echo "󰊫" ;;
    protonmail-bridge) echo "󰊫" ;;
    protonmail) echo "󰊫" ;;
    thunderbird) echo "󰊫" ;;
    sylpheed) echo "󰊫" ;;
    claws-mail) echo "󰊫" ;;
    kmail) echo "󰊫" ;;
    kontact) echo "󰊫" ;;
    evolution) echo "󰊫" ;;
    neomutt) echo "󰊫" ;;
    mutt) echo "󰊫" ;;
    alpine) echo "󰊫" ;;
    gnome-todo) echo "󰗃" ;;
    todo) echo "󰗃" ;;
    gnome-contacts) echo "󰀉" ;;
    dia) echo "󰃐" ;;
    draw.io) echo "󰃐" ;;
    inkscape) echo "" ;;
    darktable) echo "󰓰" ;;
    rawtherapee) echo "󰓰" ;;
    shotwell) echo "󰄄" ;;
    digikam) echo "󰄄" ;;
    f-spot) echo "󰄄" ;;
    gthumb) echo "󰄄" ;;
    gwenview) echo "󰄄" ;;
    showfoto) echo "󰄄" ;;
    pix) echo "󰄄" ;;
    viewnior) echo "󰄄" ;;
    ristretto) echo "󰄄" ;;
    mcomix) echo "󰄄" ;;
    comix) echo "󰄄" ;;
    foliate) echo "󰐅" ;;
    bookworm) echo "󰐅" ;;
    calibre) echo "󰐅" ;;
    fbreader) echo "󰐅" ;;
    lucidor) echo "󰐅" ;;
    sushi) echo "󰐅" ;;
    pdfmod) echo "󰐅" ;;
    pdftk) echo "󰐅" ;;
    pdfinfo) echo "󰐅" ;;
    pdfarranger) echo "󰐅" ;;
    pdfchains) echo "󰐅" ;;
    pdfshuffler) echo "󰐅" ;;
    xreader) echo "󰐅" ;;
    atril) echo "󰐅" ;;
    evince) echo "󰐅" ;;
    okular) echo "󰐅" ;;
    qpdfview) echo "󰐅" ;;
    zathura) echo "󰐅" ;;
    mupdf) echo "󰐅" ;;
    flamegraph) echo "󰜌" ;;
    perf) echo "󰜌" ;;
    sysprof) echo "󰜌" ;;
    avahi-daemon) echo "󰤨" ;;
    bluetooth) echo "󰂯" ;;
    network) echo "󰤨" ;;
    wifi) echo "󰤨" ;;
    sharing) echo "�把手" ;;
    update) echo "󰏗" ;;
    software) echo "󰏗" ;;
    appstream) echo "󰏗" ;;
    *) echo "" ;;
esac
