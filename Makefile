git-modules:
	git submodule update --remote --init --recursive

compile-storage:
	mkdir -p ton-build
	cd ton-build; CC=clang CXX=clang++ cmake -DCMAKE_BUILD_TYPE=Release ../ton; make storage-daemon -j8

build-mac:
	wails build -clean
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/TON\ Torrent.app/Contents/MacOS/storage-daemon
	wget https://ton.org/global.config.json -P build/bin/TON\ Torrent.app/Contents/MacOS/

build-linux-tar:
	wails build -clean
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/storage-daemon
	wget https://ton.org/global.config.json -P build/bin/
	tar -czvf build/bin/ton-torrent.tar.gz -C build/bin .

build-linux-deb:
	wails build -clean
	wget https://ton.org/global.config.json -P build/bin/
	mkdir -p build/bin/ton-torrent/DEBIAN
	mkdir -p build/bin/ton-torrent/usr/local/bin
	mkdir -p build/bin/ton-torrent/var/lib/ton-torrent
	mkdir -p build/bin/ton-torrent/usr/share/applications
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/ton-torrent/usr/local/bin/
	cp build/bin/TON\ Torrent build/bin/ton-torrent/usr/local/bin/ton-torrent
	cp build/appicon.png build/bin/ton-torrent/var/lib/ton-torrent/
	cp build/linux/TON\ Torrent.desktop build/bin/ton-torrent/usr/share/applications/
	cp build/linux/control build/bin/ton-torrent/DEBIAN/
	dpkg-deb --build build/bin/ton-torrent build/bin/ton-torrent.deb

build-windows:
	curl --create-dirs -O --output-dir build https://ton.org/global.config.json
	wails build -nsis