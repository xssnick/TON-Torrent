# TON Torrent
[![Based on TON][ton-svg]][ton] [![Join our group][join-svg]][tg]

TON Storage UI based on [tonutils-storage](https://github.com/xssnick/tonutils-storage), with optional storage-daemon support mode.

## Getting started

<img align="right"  width="500" alt="Screen Shot 2023-06-12 at 22 03 21" src="https://github.com/xssnick/TON-Torrent/assets/9332353/627b6327-910e-4b27-b1fa-9fcf2fc9bf32">

**Download**
* Windows x64
* Mac Apple Silicon
* Mac Intel
* Linux AMD64
* Linux ARM64

**Install**
* For Windows click on installer
* For Mac drag icon to Applications
* For Linux install .deb with `dpkg -i`

------
After installation, you could click **Add Torrent** and download `85d0998dcf325b6fee4f529d4dcf66fb253fc39c59687c82a0ef7fc96fed4c9f` to see how it works.

You could also create torrent from some of your folders and share bag id or meta file with any of your friends, but keep in mind that one of you should have public ip address, alternatively you could use one of the storage providers to host your files.

On the first start, this programm will try to resolve your external IP and check port availability, if ports are closen then you can download only from peers with public IP (similar to regular torrent).
You could always enable seed mode in settings and set external ip manualy, for example if check failed because of something else. 

## Switching to original storage-daemon

If you want to use C++ storage daemon and you have it running on your machine, you could switch to it in settings, by specifiying control port and storage-db path.

## Building

To build you need to install [Wails](https://wails.io/), then run:
`make build-[mac|windows|linux-[deb|tar]]`

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
