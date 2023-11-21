Thermostat proxy, gateway server & UI
=====================================
2022, (c) ekingr


TODO
----

- [ ] Add tutorial
- [ ] Manage error states (POST)
- [ ] Refresh UI every 10 (?) seconds
- [ ] Provide a way to refresh (drag-down not available in home-screen mode)
- [ ] Add home-screen icon
- [ ] Restric proxy incoming address to gateway server (192.168.xxx.xxx)

- [x] Add information on last update & API status
- [x] Manage error states (GET)
- [x] Add last refresh & API status information
- [x] Update assignment
- [x] Handle issues with transient states (eg. "Arrêt forcé" after switching ON, "Marche forcée" after switching OFF)


Requirements
------------

- Manual install in `wwwHome` nginx config:
```
upstream therm_backend {
	server 127.0.xxx.xxx:yyy;
}

server {
	location /api/therm/ {
		proxy_pass http://therm_backend;
	}
}
```
- Manual setup of authorizations in `wwwHome` `authConfig.json`:
```
        "food": {
            "key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
            "lvls": {
                "therm": ["myuser"]
            }
        },
```
Generate auth key with `utils/genAuthKey.go`
- Manually add `192.168.xxx.xxx   proxy.my.example.com` to `\etc\hosts\`


Architecture
------------

```
      ╭┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄╮
      ┆ ☺ User         ┆
  ▲   ╰┄┄┄┄▲┄┄┄┄┄┄┄┄┄┄┄╯
 WWW       │ Web UI (my.example.com/therm)
  ┆        │
  ▼   ┌────▼───────────┐
 pi1  │ therm_frontend │
  ║   └────┬───────────┘
  ║        │ REST API (my.example.com/api/therm)
  ║        │
  ║   ┌────▼───────────┐ REST API ┌───────────┐
 pi1  │ therm_gateway  ├──────────► home_auth │
  ▲   └────┬───────────┘          └───────────┘
 LAN       │ REST API (proxy.my.example.com)
  ┆        │ {cached internally}
  ▼   ┌────▼───────────┐
 pi4  │ therm_proxy    │
  ▲   └────┬───────────┘
 VPN       │ REST API (ctrl.my.example.com)
  ┆        │ {Nginx proxy}
  ▼   ┌────▼───────────┐
 pi0  │ therm_ctrl     │
      └────┬───────────┘
           │ SPI interface
           │
      ╭┄┄┄┄▼┄┄┄┄┄┄┄┄┄┄┄╮
      ┆ Controller     ┆
      ╰┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄╯
```
