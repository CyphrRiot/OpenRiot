#!/bin/sh
# Polybar i3 workspaces with app icons - OpenRiot
# Maps window classes to Nerd Font icons

# Case-insensitive match
CLASS="$(echo "$1" | tr '[:upper:]' '[:lower:]')"

case "$CLASS" in
    # === Terminals ===
    foot) echo "󰽒" ;;
    alacritty) echo "󰞷" ;;
    kitty) echo "󰄛" ;;
    ghostty) echo "󰊠" ;;
    wezterm) echo "󰞷" ;;
    konsole) echo "" ;;
    gnome-terminal|xfce4-terminal|xterm) echo "" ;;
    xfce4-terminal) echo "" ;;

    # === Browsers ===
    firefox) echo "󰈹" ;;
    firefox-esr) echo "󰈹" ;;
    chromium|chrome) echo "󰊯" ;;
    thorium) echo "󰊯" ;;
    librewolf) echo "󰈹" ;;
    floorp) echo "󰈹" ;;
    microsoft-edge|edge) echo "󰇩" ;;
    qutebrowser) echo "󰈹" ;;
    epiphany) echo "󰈹" ;;
    falkon) echo "󰈹" ;;
    waterfox) echo "󰈹" ;;
    brave) echo "󰊯" ;;

    # === File Managers ===
    thunar) echo "󰝰" ;;
    nautilus|nemo) echo "󰝰" ;;
    dolphin) echo "󰝰" ;;
    pcmanfm) echo "󰝰" ;;
    pcmanfm-qt) echo "󰝰" ;;
    lf) echo "󰝰" ;;
    ranger) echo "󰝰" ;;
    yazi) echo "󰝰" ;;
    nnn) echo "󰝰" ;;
   车上) echo "󰝰" ;;
   车上-desktop) echo "󰝰" ;;

    # === Communication ===
    telegram-desktop|tdesktop|telegram) echo "󰘦" ;;
    discord) echo "" ;;
    vesktop) echo "" ;;
    slack) echo "󰔟" ;;
    signal) echo "󰍡" ;;
    zoom) echo "󰕧" ;;
    teams) echo "󰊹" ;;
    thunderbird) echo "󰊫" ;;
    geary) echo "󰊫" ;;
    evolution) echo "󰊫" ;;
    mailspring) echo "󰊫" ;;
    flares) echo "󰇢" ;;
    flare) echo "󰇢" ;;

    # === Development ===
    code|code-oss) echo "󰨞" ;;
    vscodium) echo "󰨞" ;;
    helix|hx) echo "󰛞" ;;
    havoc) echo "󰊠" ;;
    vim|nvim) echo "󰈙" ;;
    sublime|subl) echo "󰅳" ;;
    jetbrains-idea|pycharm|idea) echo "" ;;
    android-studio) echo "󰀴" ;;
    gedit|gedit) echo "󰷈" ;;
    atom) echo "󰌔" ;;
    brackets) echo "󰌔" ;;
    zed) echo "󰛞" ;;
    lapce) echo "󰛞" ;;
    joe) echo "󰢝" ;;
    nano) echo "󰢝" ;;
    micro) echo "󰢝" ;;
    geany) echo "󰈙" ;;
    codeblocks) echo "󰘦" ;;
    clion) echo "󰛞" ;;
    goland) echo "󰛞" ;;
    pycharm) echo "󰛞" ;;
    rider) echo "󰛞" ;;
    rubymine) echo "󰛞" ;;
    webstorm) echo "󰛞" ;;
    phpstorm) echo "󰛞" ;;
    datagrip) echo "󰛞" ;;
    rust-analyzer) echo "󱘚" ;;

    # === Media Players ===
    spotify) echo "󰽴" ;;
    celluloid|mpv) echo "󰕼" ;;
    vlc) echo "󰕼" ;;
    totem) echo "󰕼" ;;
    audacious) echo "󰕼" ;;
    smplayer) echo "󰕼" ;;
    smplayer) echo "󰕼" ;;
    strawberry) echo "󰽴" ;;
    lxmusic) echo "󰽴" ;;
    rhythmbox) echo "󰝚" ;;
    elisa) echo "󰽴" ;;
    lollypop) echo "󰝚" ;;
    cebra) echo "󰎆" ;;
    pomodoro) echo "🍅" ;;
    cmus) echo "󰽴" ;;
    cantata) echo "󰽴" ;;
    ncmpcpp) echo "󰽴" ;;
    mpd) echo "󰽴" ;;
    arandr) echo "󰍹" ;;
    playerctl) echo "󰐌" ;;

    # === System Monitoring ===
    btop|btop) echo "󰍹" ;;
    htop|htop) echo "󰍹" ;;
    gotop) echo "󰍹" ;;
    bashtop) echo "󰍹" ;;
    ytop) echo "󰍹" ;;
    gtop) echo "󰍹" ;;
    gnome-system-monitor) echo "󰍹" ;;
    lxtask) echo "󰍹" ;;
    conky) echo "󰍹" ;;

    # === Image & Graphics ===
    gimp) echo "" ;;
    krita) echo "" ;;
    inkscape) echo "" ;;
    blender) echo "󰂫" ;;
    darktable) echo "󰓰" ;;
    rawtherapee) echo "󰓰" ;;
    eog|eyeofgnome) echo "󰋨" ;;
    gwenview) echo "󰋨" ;;
    shotwell) echo "󰄄" ;;
    digikam) echo "󰄄" ;;
    gthumb) echo "󰄄" ;;
    viewnior) echo "󰄄" ;;
    ristretto) echo "󰄄" ;;
    imv) echo "󰋨" ;;
    feh) echo "󰋨" ;;
    sxiv) echo "󰋨" ;;
    pqiv) echo "󰋨" ;;
    showfoto) echo "󰄄" ;;
    pix) echo "󰄄" ;;
    simplescan) echo "󰈿" ;;
    cheese) echo "󰑮" ;;
    webcamoid) echo "󰑮" ;;
    guvcview) echo "󰑮" ;;

    # === Video Editing ===
    kdenlive) echo "" ;;
    pitivi) echo "󰒒" ;;
    openshot) echo "" ;;
    shotcut) echo "" ;;
    blender) echo "󰂫" ;;
    obs|obsidian) echo "󰑋" ;;
    obs-studio) echo "󰑋" ;;
    vokoscreen) echo "󰑋" ;;
    simplescreenrecorder) echo "󰑋" ;;
    kazam) echo "󰑋" ;;

    # === Games ===
    steam) echo "󰓓" ;;
    lutris) echo "󰓓" ;;
    heroic) echo "󰓓" ;;
    bottles) echo "󰓓" ;;
    mangohud) echo "󰓓" ;;
    gamemode) echo "󰓓" ;;
    prismlauncher) echo "󰓓" ;;
    multimc) echo "󰓓" ;;
    minecraft) echo "󰓓" ;;
    prusaslicer|orca-slicer|ultimaker-cura) echo "󰹛" ;;

    # === Office ===
    libreoffice|libreoffice-writer) echo "󰈙" ;;
    libreoffice-calc) echo "󰈙" ;;
    libreoffice-impress) echo "󰈙" ;;
    libreoffice-draw) echo "󰈙" ;;
    libreoffice-math) echo "󰈙" ;;
    libreoffice-startcenter) echo "󰏆" ;;
    soffice) echo "󰏆" ;;
    abiword) echo "󰈙" ;;
    gnome-documents) echo "󰈙" ;;
    calligra) echo "󰈙" ;;

    # === PDF ===
    evince) echo "󰐅" ;;
    zathura) echo "󰐅" ;;
    okular) echo "󰐅" ;;
    qpdfview) echo "󰐅" ;;
    atril) echo "󰐅" ;;
    xreader) echo "󰐅" ;;
    mupdf) echo "󰐅" ;;
    apvlv) echo "󰐅" ;;
    pdfpc) echo "󰐅" ;;
    foliate) echo "󰐅" ;;
    bookworm) echo "󰐅" ;;
    calibre) echo "󰐅" ;;
    fbreader) echo "󰐅" ;;

    # === Settings & System ===
    gnome-control-center|gnome-settings) echo "󰒓" ;;
    xfce4-settings-manager|xfce4-settings) echo "󰒓" ;;
    gnome-tweaks) echo "󰒓" ;;
    systemsettings) echo "󰒓" ;;
    lxappearance) echo "󰀻" ;;
    qt5ct) echo "󰀻" ;;
    qt6ct) echo "󰀻" ;;
    gnome-software) echo "󰏗" ;;
    muon) echo "󰏗" ;;
    pkgsrc) echo "󰏗" ;;
    gnome-disks) echo "󰋊" ;;
    gnome-disk-utility) echo "󰋊" ;;
    gparted) echo "" ;;
    gnome-system-log) echo "󰁯" ;;
    gnome-logs) echo "󰁯" ;;
    gnome-usage) echo "󰍹" ;;
    gnome-power-stats) echo "󰠭" ;;
    gnome-power-stat) echo "󰠭" ;;
    gnome-about) echo "󰋖" ;;
    gnome-extensions) echo "󰔎" ;;
    gnome-font-viewer) echo "󰛖" ;;
    gnome-fonts) echo "󰛖" ;;
    dconf-editor) echo "󰒃" ;;
    gnome-connections) echo "󰢹" ;;
    vinagre) echo "󰢹" ;;
    remmina) echo "󰢹" ;;
    virt-manager) echo "󰢹" ;;
    virtualbox) echo "󰌗" ;;
    gnome-calendar) echo "󰃭" ;;
    gnome-clocks) echo "󰔗" ;;
    gnome-calculator) echo "󰪚" ;;
    gnome-weather) echo "󰖕" ;;
    gnome-maps) echo "󰟡" ;;
    gnome-contacts) echo "󰀉" ;;
    gnome-characters) echo "󰋽" ;;
    gnome-dictionary) echo "󰋽" ;;
    polari) echo "󰯯" ;;
    gnome-text-editor) echo "󰷈" ;;
    gnome-todo) echo "󰗃" ;;
    baobab) echo "󰋊" ;;
    d-feet) echo "󰯾" ;;
    gnome-screenshot) echo "󰄀" ;;
    flameshot) echo "󰄀" ;;
    scrot) echo "󰄀" ;;
    spectacle) echo "󰄀" ;;
    gnome-sound-recorder) echo "󰏘" ;;
    pavucontrol) echo "󱡫" ;;
    pulsemixer) echo "󱡫" ;;
    alsa) echo "󰕔" ;;
    blueman-manager) echo "󰂯" ;;
    bluetooth) echo "󰂯" ;;
    nm-connection-editor) echo "󰤨" ;;
    networkmanager) echo "󰤨" ;;
    nmtui) echo "󰤨" ;;
    iwctl) echo "󰤨" ;;
    connman) echo "󰤨" ;;
    gnome-network) echo "󰤨" ;;

    # === Utilities ===
    file-roller) echo "󰗄" ;;
    engrampa) echo "󰗄" ;;
    ark) echo "󰗄" ;;
    xarchiver) echo "󰗄" ;;
    7z) echo "󰗄" ;;
    transmission) echo "󰇚" ;;
    transmission-gtk) echo "󰇚" ;;
    qbittorrent) echo "󰇚" ;;
    deluge) echo "󰇚" ;;
    rtorrent) echo "󰇚" ;;
    vuze) echo "󰇚" ;;
    fragments) echo "󰇚" ;;
    ncdc) echo "󰇚" ;;
    gnome-terminal|preferences) echo "�贤" ;;
    gnome-terminal) echo "󰒏" ;;
    xfce4-terminal) echo "󰒏" ;;
    gnome-appfolders) echo "󰰀" ;;
    nautilus) echo "󰝰" ;;
    doublecmd) echo "󰝰" ;;
    midnight-commander|midnight-commander) echo "󰝰" ;;
    mc) echo "󰝰" ;;
    crashplan) echo "󰒝" ;;
    timeshift) echo "󰣜" ;;
    grsync) echo "�、背" ;;
    deja-dup) echo "�、背" ;;
    syncthing) echo "�、背" ;;
    restic) echo "�、背" ;;
    rclone) echo "�、背" ;;
    borg) echo "�、背" ;;
    duplicati) echo "�、背" ;;

    # === AI / CLI Tools ===
    crush|crush-cli) echo "󰚩" ;;
    openai) echo "󰚩" ;;
    chatgpt) echo "󰚩" ;;
    groq) echo "󰚩" ;;
    ollama) echo "󰚩" ;;
    lm-studio) echo "󰚩" ;;
    jan) echo "󰚩" ;;
    text-generation-webui) echo "󰚩" ;;
    localai) echo "󰚩" ;;

    # === Other ===
    *) echo "" ;;
esac
