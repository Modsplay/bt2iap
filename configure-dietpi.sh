#!/bin/bash
set -euo pipefail
cd /opt/bt2iap
backup=/root/bt2iap-before-setup
mkdir -p "$backup"
for file in /etc/asound.conf /etc/bluetooth/main.conf /boot/config.txt /boot/cmdline.txt /boot/firmware/config.txt /boot/firmware/cmdline.txt; do
 if [ -f "$file" ] && [ ! -f "$backup$file" ]; then mkdir -p "$backup$(dirname "$file")"; cp -a "$file" "$backup$file"; fi
done
install -m644 alsa/asound.conf /etc/asound.conf
install -m644 modules-load.d/bt2iap.conf /etc/modules-load.d/bt2iap.conf
[ -f /etc/default/bt2iap ] || install -m644 etc/default/bt2iap.example /etc/default/bt2iap
for unit in systemd/*.service; do install -m644 "$unit" /etc/systemd/system/; done
install -Dm644 systemd/bluealsa.service.d/override.conf /etc/systemd/system/bluealsa.service.d/override.conf
chmod +x scripts/*.sh bluetooth/pair-agent.sh
# Edit the existing INI sections instead of appending duplicate sections.
python3 - <<'PY'
import configparser
p='/etc/bluetooth/main.conf'
c=configparser.ConfigParser(strict=False);c.optionxform=str;c.read(p)
for section,values in {'General':{'Name':'bt2iap','Class':'0x240404','DiscoverableTimeout':'0','PairableTimeout':'0','JustWorksRepairing':'always'},'Policy':{'AutoEnable':'true'}}.items():
 if not c.has_section(section):c.add_section(section)
 for k,v in values.items():c.set(section,k,v)
with open(p,'w') as f:c.write(f)
PY
mkdir -p /etc/systemd/system/pair-agent.service.d
cat > /etc/systemd/system/pair-agent.service.d/dietpi.conf <<'EOF'
[Service]
ExecStart=
ExecStart=/usr/bin/python3 /opt/bt2iap/bluetooth/pair-agent.py
EOF
install -m755 /opt/bt2iap/usb-mode /usr/local/sbin/bt2iap-usb-mode
systemctl disable --now bluealsa-aplay.service || true
systemctl unmask bluetooth.service dbus.service dbus.socket
systemctl daemon-reload
modprobe snd-aloop
systemctl enable --now bluetooth.service bluealsa.service
systemctl enable ipod-gadget.service ipod-session.service audio-bridge.service audio-loopback.service pair-agent.service
systemctl restart bluetooth.service bluealsa.service
systemctl start pair-agent.service
