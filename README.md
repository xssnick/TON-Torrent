# TON Torrent
[![Based on TON][ton-svg]][ton] [![Join our group][join-svg]][tg]

TON Storage UI based on [tonutils-storage](https://github.com/xssnick/tonutils-storage), and with optional storage-daemon support mode.

## Quick start

Download:
* Windows x64
* Mac Apple Silicon
* Mac Intel
* Linux AMD64
* Linux ARM64

After installation, you could click Add Torrent and download `85d0998dcf325b6fee4f529d4dcf66fb253fc39c59687c82a0ef7fc96fed4c9f` to see how it works.

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