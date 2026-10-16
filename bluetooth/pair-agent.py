#!/usr/bin/python3
"""Headless pairing agent for the bt2iap audio receiver."""
import dbus
import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GLib
DBusGMainLoop(set_as_default=True)
bus=dbus.SystemBus()
def trust(device):
    dbus.Interface(bus.get_object('org.bluez',device),'org.freedesktop.DBus.Properties').Set('org.bluez.Device1','Trusted',dbus.Boolean(True))
class Agent(dbus.service.Object):
    @dbus.service.method('org.bluez.Agent1',in_signature='',out_signature='')
    def Release(self): loop.quit()
    @dbus.service.method('org.bluez.Agent1',in_signature='o',out_signature='s')
    def RequestPinCode(self,device): trust(device); return '0000'
    @dbus.service.method('org.bluez.Agent1',in_signature='o',out_signature='u')
    def RequestPasskey(self,device): trust(device); return dbus.UInt32(0)
    @dbus.service.method('org.bluez.Agent1',in_signature='ou',out_signature='')
    def RequestConfirmation(self,device,passkey): trust(device)
    @dbus.service.method('org.bluez.Agent1',in_signature='o',out_signature='')
    def RequestAuthorization(self,device): trust(device)
    @dbus.service.method('org.bluez.Agent1',in_signature='os',out_signature='')
    def AuthorizeService(self,device,uuid): trust(device)
    @dbus.service.method('org.bluez.Agent1',in_signature='',out_signature='')
    def Cancel(self): pass
agent=Agent(bus,'/bt2iap/agent')
manager=dbus.Interface(bus.get_object('org.bluez','/org/bluez'),'org.bluez.AgentManager1')
manager.RegisterAgent('/bt2iap/agent','NoInputNoOutput')
manager.RequestDefaultAgent('/bt2iap/agent')
objects=dbus.Interface(bus.get_object('org.bluez','/'),'org.freedesktop.DBus.ObjectManager').GetManagedObjects()
for path,interfaces in objects.items():
    if 'org.bluez.Adapter1' not in interfaces: continue
    props=dbus.Interface(bus.get_object('org.bluez',path),'org.freedesktop.DBus.Properties')
    for key,val in [('Powered',dbus.Boolean(True)),('Alias','bt2iap'),('DiscoverableTimeout',dbus.UInt32(0)),('PairableTimeout',dbus.UInt32(0)),('Pairable',dbus.Boolean(True)),('Discoverable',dbus.Boolean(True))]:
        props.Set('org.bluez.Adapter1',key,val)
print('bt2iap pairing agent ready',flush=True)
loop=GLib.MainLoop();loop.run()
