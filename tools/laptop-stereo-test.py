#!/usr/bin/env python3
"""Permissive iAP1 accessory emulator; capture USB PCM statistics, no playback.
Requires python-libusb1. Test only after resolving laptop USB driver faults.
Does not implement MFi cryptographic authentication or guarantee car compatibility.
"""
import array
import math
import struct
import time
import usb1


def report(payload):
    frame = bytes([0x55, len(payload)]) + payload
    frame += bytes([(-sum(frame[1:])) & 255])
    for rid, length in [(5, 8), (6, 10), (7, 14), (8, 20), (9, 63)]:
        if len(frame) <= length - 1:
            return bytes([rid, 0]) + frame.ljust(length - 1, b'\0')
    raise ValueError('Message too long')


def main():
    with usb1.USBContext() as ctx:
        dev = next((d for d in ctx.getDeviceList(skip_on_error=True)
                    if (d.getVendorID(), d.getProductID()) == (0x05ac, 0x1297)), None)
        if dev is None:
            raise RuntimeError('Pi USB 05ac:1297 not found')
        cfg = next(c for c in dev if any(s.getClass() == 3 for i in c for s in i))
        settings = [s for i in cfg for s in i]
        hid = next(s for s in settings if s.getClass() == 3)
        aud = next(s for s in settings if s.getClass() == 1 and s.getAlternateSetting() == 1)
        hid_ep = next(e.getAddress() for e in hid if e.getAddress() & 128)
        audio_ep = next(e for e in aud if (e.getAttributes() & 3) == 1)
        h = dev.open()
        claimed = []
        transfers = []
        collecting = False
        original = None
        try:
            h.setAutoDetachKernelDriver(True)
            original = h.getConfiguration()
            if original != cfg.getConfigurationValue():
                h.setConfiguration(cfg.getConfigurationValue())
            for iface in (hid.getNumber(), aud.getNumber()):
                h.claimInterface(iface)
                claimed.append(iface)

            def send(lingo, command, data=b''):
                pkt = report(bytes([lingo, command]) + data)
                h.controlWrite(0x21, 9, 0x200 | pkt[0], hid.getNumber(), pkt, timeout=1500)
                print('TX', hex(lingo), hex(command), flush=True)

            send(0, 7)  # RequestiPodName
            send(0, 0x13, struct.pack('>III', (1 << 10) | 1, 0, 1))
            ready = False
            name_seen = False
            deadline = time.monotonic() + 10
            frame = bytearray()
            while time.monotonic() < deadline and not (ready and name_seen):
                try:
                    raw = bytes(h.interruptRead(hid_ep, 64, timeout=1000))
                except usb1.USBErrorTimeout:
                    continue
                if len(raw) < 3:
                    continue
                if not (raw[1] & 1):
                    frame.clear()
                frame.extend(raw[2:])
                if raw[1] & 2:
                    continue
                if len(frame) < 4 or frame[0] != 0x55 or not frame[1]:
                    raise RuntimeError('Unexpected iAP frame')
                size = frame[1]
                if len(frame) < size + 3 or sum(frame[1:size + 3]) & 255:
                    raise RuntimeError('Invalid iAP packet length/checksum')
                p = bytes(frame[2:size + 2])
                print('RX', p.hex(), flush=True)
                if p[:2] == b'\x00\x08':
                    name_seen = True
                elif p[:2] == b'\x00\x14':
                    send(0, 0x15, b'\x01\x00')  # Legacy accessory authentication info
                elif p[:2] == b'\x0a\x02':
                    send(10, 3, struct.pack('>I', 44100))
                elif p[:2] == b'\x0a\x04':
                    if struct.unpack('>I', p[2:6])[0] != 44100:
                        raise RuntimeError('Unexpected negotiated sample rate')
                    send(10, 0, b'\x00\x04')
                    ready = True
            if not (ready and name_seen):
                raise RuntimeError('Handshake incomplete; no USB audio test performed')
            print('iAP name query and 44100 Hz audio negotiation succeeded.', flush=True)
            ep = audio_ep.getAddress()
            h.controlWrite(0x22, 1, 0x100, ep, bytes([0x44, 0xac, 0]), timeout=1500)
            h.setInterfaceAltSetting(aud.getNumber(), aud.getAlternateSetting())
            pcm = bytearray()
            errors = []
            collecting = True

            def completed(t):
                if t.getStatus() == usb1.TRANSFER_COMPLETED:
                    for status, data in t.iterISO():
                        if status == usb1.TRANSFER_COMPLETED:
                            pcm.extend(data)
                        else:
                            errors.append(status)
                    if collecting:
                        t.submit()
                elif t.getStatus() != usb1.TRANSFER_CANCELLED:
                    errors.append(t.getStatus())

            packet_size = audio_ep.getMaxPacketSize()
            for _ in range(8):
                t = h.getTransfer(iso_packets=32)
                t.setIsochronous(ep, packet_size * 32, completed, timeout=2000)
                transfers.append(t)
                t.submit()
            end = time.monotonic() + 5
            while time.monotonic() < end:
                ctx.handleEventsTimeout(tv=0.1)
            collecting = False
            for t in transfers:
                if t.isSubmitted():
                    t.cancel()
            while any(t.isSubmitted() for t in transfers):
                ctx.handleEventsTimeout(tv=0.1)
            samples = array.array('h', pcm[:len(pcm)//4*4])
            peak = max(map(abs, samples), default=0)
            rms = math.sqrt(sum(x*x for x in samples)/len(samples)) if samples else 0
            print('USB PCM bytes:', len(pcm), 'packet errors:', len(errors))
            print('Peak dBFS:', 20*math.log10(peak/32768) if peak else '-inf')
            print('RMS dBFS:', 20*math.log10(rms/32768) if rms else '-inf')
            if not pcm or errors:
                raise RuntimeError('USB audio transfer incomplete or reported errors')
            print('Non-silent USB audio verified.' if peak else 'USB stream contains silence; play phone audio and retry.')
        finally:
            collecting = False
            for t in transfers:
                if t.isSubmitted():
                    t.cancel()
            while any(t.isSubmitted() for t in transfers):
                ctx.handleEventsTimeout(tv=0.1)
            if aud.getNumber() in claimed:
                h.setInterfaceAltSetting(aud.getNumber(), 0)
            for iface in reversed(claimed):
                h.releaseInterface(iface)
            if original is not None and original != cfg.getConfigurationValue():
                h.setConfiguration(original)
            h.close()


if __name__ == '__main__':
    main()
