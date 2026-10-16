# DietPi Bluetooth iPod adapter

This version bundles the modified iPod protocol implementation. It reads the connected phone's BlueZ media player directly, without writing metadata to disk.

## Added capabilities

- Live title, artist, album, genre, duration, playback position and status when supplied by the phone.
- Head-unit next/previous track, play/pause toggle, explicit play, pause and stop.
- Fast-forward/rewind and shuffle/repeat requests when supported by the phone's player.
- Subscribed track/position notifications and simple-remote button hold suppression.
- Writable DietPi setup, USB serial maintenance switching and optional RAM-backed logging/zram.

The adapter exposes the current song as one virtual track. Full phone-library browsing and album artwork are not implemented. Protocol and race tests pass; physical head-unit compatibility still needs testing in a car. The laptop emulator is experimental and has not completed a hardware handshake test.

## DietPi installation

Use a 64-bit DietPi image on Pi Zero 2 W. Configure Wi-Fi and account credentials with DietPi first. Keep SSH available when changing USB modes.

```sh
sudo git clone https://github.com/Modsplay/bt2iap /opt/bt2iap
cd /opt/bt2iap
sudo apt update
sudo apt install -y build-essential linux-headers-$(uname -r) golang git bluez bluez-tools bluez-alsa-utils libasound2-plugin-bluez alsa-utils python3-dbus python3-gi
cd vendor/ipod-gadget/gadget
sudo make -j2
cd ../../ipod
sudo go build -o ipod ./cmd/ipod
cd /opt/bt2iap
sudo bash configure-dietpi.sh
```

Enable `dtoverlay=dwc2,dr_mode=peripheral` in the active boot config, then reboot if the overlay was not previously active. With SSH connected, run `sudo bt2iap-usb-mode ipod`. Pair your phone with **bt2iap**. The pairing agent automatically accepts and trusts devices while discoverable.

Optional write reduction: `sudo bash scripts/reduce-writes-dietpi.sh`. Logs disappear at reboot; the root filesystem remains writable.

## Diagnostics and development

```sh
sudo /opt/bt2iap/vendor/ipod/ipod media
sudo /opt/bt2iap/vendor/ipod/ipod control Next
cd /opt/bt2iap/vendor/ipod
go test -race ./...
```

Other control actions include Previous, Play, Pause, Stop, Toggle, FastForward and Rewind. Seek and playback-setting support depend on the phone/app. See the deployment notes for USB console mode and laptop test preparation.
