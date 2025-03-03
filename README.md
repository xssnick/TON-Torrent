# TON Torrent
[![Based on TON][ton-svg]][ton] [![Join our group][join-svg]][tg]

TON Storage UI based on [tonutils-storage](https://github.com/xssnick/tonutils-storage), with storage-daemon support mode.

## Getting started
<img align="right"  width="500" alt="Preview" src="https://github.com/xssnick/TON-Torrent/assets/9332353/c8065c0e-b7f2-4b6f-bcf0-37d0180d1dbe">

**Download**
* [Windows x64](https://github.com/xssnick/TON-Torrent/releases/latest/download/ton-torrent-windows-x64-installer.exe)
* [Mac Apple Silicon](https://github.com/xssnick/TON-Torrent/releases/latest/download/ton-torrent-mac-apple-silicon.dmg)
* [Mac Intel](https://github.com/xssnick/TON-Torrent/releases/latest/download/ton-torrent-mac-intel.dmg)
* [Linux AMD64](https://github.com/xssnick/TON-Torrent/releases/latest/download/ton-torrent-linux-amd64.deb)
* [Linux ARM64](https://github.com/xssnick/TON-Torrent/releases/latest/download/ton-torrent-linux-arm64.deb)

**Install**
* For Windows click on installer
* For Mac drag icon to Applications
* For Linux install .deb with `dpkg -i`

------
After installation, you could click **Add Torrent** and download `85d0998dcf325b6fee4f529d4dcf66fb253fc39c59687c82a0ef7fc96fed4c9f` to see how it works.

You could also create torrent from some of your folders and share bag id or meta file with your friends, but keep in mind, that at least one of you should have public ip address. Otherwise, you could use one of the storage providers to host your files.

At the first start, this program will try to resolve your external IP and check port availability. If ports are closen, then you can download only from peers with public IP (similar to regular torrent).
You could always enable seed mode in settings and set external ip manually, for example, if check failed because of something else. 

## Switching to original storage-daemon

If you want to use C++ storage daemon and you have it running on your machine, you could switch to it in settings by specifying control port and storage-db path.

## Building

To build, you need to install [Wails](https://wails.io/), then run:
`make build-[mac|windows|linux-[deb|tar]]`

## Tunnel usage

Tunnels adds ability to rent address+port from another node, for example to can share your bags even to peers with non-public IP.
This is useful when your provider not gives you public ip address, but you want to host your ton-site or bags from local computer.
This can also speed up download speed becuase you can connect to more peers.

1. Create empty file with name `tunnel-config.json`
2. At Settings select this file as tunnel config.
3. Restart TON Torrent, on next start it will generate basic config in this file.
4. Fill config with desired tunnel route and start TON Torrent again.
5. You will be connected to peers through tunnel.

In case of fail, or if you want to disable tunnel - remove tunnel-config file, restart app and clear config in Settings. 

### Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

<!-- Badges -->
[ton-svg]: https://img.shields.io/badge/Based%20on-TON-blue
[join-svg]: https://img.shields.io/badge/Join%20-Telegram-blue
[ton]: https://ton.org
[tg]: https://t.me/tonrh
