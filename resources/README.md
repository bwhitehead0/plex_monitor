# Installing as a service

The systemd unit file `plex_monitor.service` should suffice as a functional service for most use cases.

Simply copy the file to `/etc/systemd/system/` and, after, updating the binary and config file locations for `ExecStart`, start the service.

Service logs will be written to `/var/log/plex_monitor.log`.

Manage the service with typical `systemctl` commands.

```
bwhitehead@log01:~$ sudo systemctl status plex_monitor.service 
● plex_monitor.service - Plex Monitor
     Loaded: loaded (/etc/systemd/system/plex_monitor.service; enabled; vendor preset: enabled)
     Active: active (running) since Wed 2023-10-04 15:26:49 UTC; 2s ago
       Docs: https://github.com/bwhitehead0/plex_monitor
   Main PID: 24271 (plex_monitor)
      Tasks: 8 (limit: 9347)
     Memory: 1.3M
        CPU: 9ms
     CGroup: /system.slice/plex_monitor.service
             └─24271 /usr/local/bin/plex_monitor --config.file=/etc/plex_monitor.yaml

Oct 04 15:26:49 log01.fw.home systemd[1]: Started Plex Monitor.
```