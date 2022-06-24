# Run DSUL daemon with systemd and udev

Using _systemd_ and _udev_ together let's us automatically start the daemon when the USB device is plugged in.

1. Edit `99-dsul.rules` so that the vendor and product id's match the device you are using.
0. Place the file in `/etc/udev/rules.d/` and run `udevadm control --reload-rules`.
0. Edit `dsul.service` so it uses the appropriate ExecStart format for your setup.
0. Place the file in `/etc/systemd/system/dsul.service` and run `systemctl daemon-reload`.


## How it works

When the USB device is plugged in, it's detected by _udev_ and our rule is matched
against the id's and the tty device created gets a symlink (`/dev/dsul`) and we
also tag it with `systemd` to make sure _systemd_ knows to handle it. We also tell
_systemd_ which service is wanted by our newly created device. Once the device is
available to _systemd_ the service is started. If the USB device is removed, _udev_
will remove all devices and once the _systemd_ device no longer exists, _systemd_
then stops the service.
