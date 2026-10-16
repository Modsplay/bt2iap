#!/bin/bash
# Run after reboot, with the Pi unplugged. Changes last until the next reboot.
# Temporarily disables laptop USB audio drivers and Apple usbmuxd auto-probing.
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run with sudo'; exit 1; }
if lsusb -d 05ac:1297 | grep -q .; then
 echo 'Unplug the Pi before preparing the laptop.'; exit 1
fi
systemctl mask --runtime --now usbmuxd.service
modprobe -r snd_usb_audio
modprobe snd_usb_audio enable=0
echo 'Ready: connect the Pi and run laptop-stereo-test.py. Reboot restores laptop drivers.'
