build-mac:
	CGO_ENABLED=1 wails build -clean

build-linux-tar:
	CGO_ENABLED=1 wails build -clean
	tar -czvf build/bin/ton-torrent.tar.gz -C build/bin .

build-linux-deb:
	CGO_ENABLED=1 wails build -clean -ldflags="-X 'main.CustomRoot=/opt/ton-torrent'"
	mkdir -p build/bin/ton-torrent/DEBIAN
	mkdir -p build/bin/ton-torrent/usr/local/bin
	mkdir -p build/bin/ton-torrent/opt/ton-torrent
	mkdir -p build/bin/ton-torrent/usr/share/applications
	cp build/bin/TON\ Torrent build/bin/ton-torrent/usr/local/bin/ton-torrent
	cp build/appicon.png build/bin/ton-torrent/opt/ton-torrent/
	cp build/linux/ton-torrent.desktop build/bin/ton-torrent/usr/share/applications/
	cp build/linux/control build/bin/ton-torrent/DEBIAN/
	cp build/linux/postinst build/bin/ton-torrent/DEBIAN/
	chmod 775 build/bin/ton-torrent/DEBIAN/postinst
	dpkg-deb --build build/bin/ton-torrent build/bin/ton-torrent.deb

build-windows:
	CGO_ENABLED=1 wails build -nsis

sign-mac:
	gon ./build/gon/config.json

mac-arm64-all:
	GOOS=darwin GOARCH=arm64 make build-mac
	sleep 1
	make sign-mac

mac-amd64-all:
	GOOS=darwin GOARCH=amd64 make build-mac
	sleep 1
	make sign-mac

windows-amd64-all:
	GOOS=windows GOARCH=amd64 make build-windows

linux-amd64-all:
	GOOS=linux GOARCH=amd64 make build-linux-deb

linux-arm64-all:
	GOOS=linux GOARCH=arm64 make build-linux-deb