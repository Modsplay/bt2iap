# DietPi installation on this Pi

Installed 2026-10-16 on Pi Zero 2 W, DietPi 10.7.2 / Debian 13 arm64.
Kernel: 6.18.50+rpt-rpi-v8. bt2iap source: 6369c96a8c88d99527976c88c0b16c4694c8de24.

## Use

Connect the Pi's USB data port to the stereo's iPod USB input. Pair the phone
with Bluetooth **bt2iap**, select the stereo's iPod source, and play music.
Login over Wi-Fi: `ssh ipod@<pi-address>` on your network. The DHCP
address may change on another network. Use the password you supplied.

Root and boot filesystems remain writable. No overlay/read-only filesystem
configuration was installed.

## USB maintenance mode

The iPod gadget and USB serial gadget share the controller and cannot run
simultaneously. From an SSH session:

- `sudo bt2iap-usb-mode serial` restores USB serial and persists that mode.
- `sudo bt2iap-usb-mode ipod` restores the adapter and persists that mode.

After switching to serial, connect from the PC with `sudo screen /dev/ttyACM0 115200`.
Switching back to iPod disconnects serial; use SSH for that command.

## Installation changes

- Sources in `/opt/bt2iap`, upstream dependencies in `vendor/`.
- Installed matching kernel headers, compiler tools, BlueZ, bluez-alsa-utils,
  libasound2-plugin-bluez, alsa-utils, Go, Git, python3-dbus and python3-gi.
- Built all three ipod-gadget modules against the running kernel.
- Built the Go control program for linux/arm64 and verified it runs.
- Installed repository audio routing and service definitions.
- `configure-dietpi.sh` applies DietPi configuration; it expects dependencies
  and compiled binaries/modules to exist. The original bootstrap.sh was not run.
- A Python D-Bus pairing agent replaces the interactive bt-agent wrapper via
  a systemd drop-in. It accepts pairing and trusts connected devices.
- Stock bluealsa-aplay.service is disabled to avoid duplicate audio consumers.
- Wi-Fi credentials and account passwords were preserved.
- Configuration backups: `/root/bt2iap-before-setup`.

## Checks performed

After reboot: ipod-gadget, ipod-session, bluealsa, pair-agent, audio-bridge,
and audio-loopback are active; no failed systemd units. ALSA shows Loopback
and iPodUSB. The PC detects USB 05ac:1297. Bluetooth is discoverable and
pairable as bt2iap. Root mount reports `ext4 rw,noatime,lazytime`.

Phone pairing and incoming Bluetooth audio have been verified. Actual stereo playback has not been verified. Some stereos
reject particular product IDs or require unsupported iPod authentication.
Default product ID is 0x1297; alternatives can be set in /etc/default/bt2iap.

## Maintenance

After a kernel update, rebuild the modules for the new kernel before using
adapter mode. Install that kernel's matching headers, then:

    cd /opt/bt2iap/vendor/ipod-gadget/gadget
    sudo make clean
    sudo make -j2

Diagnose with `systemctl --failed`, `journalctl -u ipod-session`,
`journalctl -u pair-agent`, and `journalctl -u audio-loopback`.

## Reduced SD writes (2026-10-16)

Replaced /var/swap (1.5 GiB SD-backed swap) with DietPi-managed 231 MiB
zstd zram swap using `/boot/dietpi/func/dietpi-set_swapfile 1 zram`.
DietPi resized the existing /tmp RAM mount limit to 346 MiB. /var/log remains
its existing 50 MiB RAM mount. Explicit journald policy lives in
`/etc/systemd/journald.conf.d/zz-bt2iap-ram.conf`: volatile storage, 8 MiB cap,
2 MiB files, 32 MiB free-space reserve, info-level storage, no syslog forwarding.
Logs are temporary and disappear on reboot; pairing and settings still persist.
No read-only root or overlay filesystem was enabled. Update checks and the
filesystem journal remain enabled. This reduces writes, not all sudden-power-loss risk.

Verified after reboot: only zram swap active, SD swap file absent, all six
adapter services active, no failed units, root still writable ext4.
Backups of prior fstab and DietPi settings are in
`/root/bt2iap-before-write-reduction`.
To restore the former swap arrangement, run
`sudo /boot/dietpi/func/dietpi-set_swapfile 1586 /var/swap`.
To revert journal policy, remove the named zz-bt2iap-ram.conf drop-in and restart
systemd-journald.

## Laptop stereo emulation test (prepared, not yet hardware-tested)

`tools/laptop-stereo-test.py` queries the iAP name, identifies an accessory,
negotiates 44100 Hz audio, and measures USB isochronous PCM for five seconds.
It does not play audio or save a recording. Dependencies: python-libusb1.
Syntax and packet checksum checks passed; hardware execution is pending.

Laptop kernel 7.0.0-31-generic logged a snd_usb_audio probe NULL-pointer fault
while usbmuxd selected the Pi configuration at 10:57 on Oct 16. USB opens are
blocked in the kernel. Reboot the laptop with the Pi unplugged before testing.
Then run `sudo bash tools/prepare-laptop-test.sh` BEFORE reconnecting the Pi.
This masks usbmuxd at runtime and loads snd_usb_audio with enable=0 so the
userspace libusb test owns the interfaces. It temporarily disables USB audio
on the laptop; a reboot restores its normal behavior. No Pi settings change.
