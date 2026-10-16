#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo "Run as root" >&2; exit 1; }
[[ -x /boot/dietpi/func/dietpi-set_swapfile ]] || { echo "DietPi required" >&2; exit 1; }
backup=/root/bt2iap-before-write-reduction
mkdir -p "$backup"
for file in /etc/fstab /boot/dietpi.txt; do
  [[ ! -f "$file" || -e "$backup/$(basename "$file")" ]] || cp -a "$file" "$backup/"
done
/boot/dietpi/func/dietpi-set_swapfile 1 zram
mkdir -p /etc/systemd/journald.conf.d
cat > /etc/systemd/journald.conf.d/zz-bt2iap-ram.conf <<'EOF'
[Journal]
Storage=volatile
RuntimeMaxUse=8M
RuntimeMaxFileSize=2M
RuntimeKeepFree=32M
MaxLevelStore=info
ForwardToSyslog=no
EOF
systemctl restart systemd-journald
swapon --show
findmnt /
