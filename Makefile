git-modules:
	git submodule update --remote --init --recursive

compile-storage:
	mkdir -p ton-build
	cd ton-build; CC=clang CXX=clang++ cmake -DCMAKE_BUILD_TYPE=Release ../ton; make storage-daemon -j8

download-storage-mac-arm:
	curl --create-dirs -O --output-dir ton-build/storage/storage-daemon/ https://cicd.neodix.io/view/neodiX/job/TON_macOS_aarch64_arm64_neodix/17/artifact/artifacts/storage-daemon

download-storage-mac-amd:
	curl --create-dirs -O --output-dir ton-build/storage/storage-daemon/ https://cicd.neodix.io/job/TON_macOS_10.15_x86-64_master/lastSuccessfulBuild/artifact/artifacts/storage-daemon

download-storage-windows-amd:
	curl --create-dirs -O --output-dir build/windows/ https://aka.ms/vs/17/release/vc_redist.x64.exe
	curl --create-dirs -O --output-dir ton-build/storage/storage-daemon/ https://cicd.neodix.io/job/TON_Windows_x86-64_master/lastSuccessfulBuild/artifact/artifacts/storage-daemon.exe

download-storage-linux-amd:
	curl --create-dirs -O --output-dir ton-build/storage/storage-daemon/ https://cicd.neodix.io/job/TON_Linux_x86-64/lastSuccessfulBuild/artifact/artifacts/storage-daemon

download-storage-linux-arm:
	curl --create-dirs -O --output-dir ton-build/storage/storage-daemon/ https://cicd.neodix.io/job/TON_Linux_arm64/lastSuccessfulBuild/artifact/artifacts/storage-daemon

build-mac:
	CGO_ENABLED=1 wails build -clean
	mkdir -p build/bin/TON\ Torrent.app/Contents/Resources/Storage.app/Contents/MacOS
	mkdir -p build/bin/TON\ Torrent.app/Contents/Resources/Storage.app/Contents/Resources
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/TON\ Torrent.app/Contents/Resources/Storage.app/Contents/MacOS/
	cp build/darwin/daemon/Info.plist build/bin/TON\ Torrent.app/Contents/Resources/Storage.app/Contents/
	cp build/bin/TON\ Torrent.app/Contents/Resources/iconfile.icns build/bin/TON\ Torrent.app/Contents/Resources/Storage.app/Contents/Resources/

build-linux-tar:
	CGO_ENABLED=1 wails build -clean
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/storage-daemon
	tar -czvf build/bin/ton-torrent.tar.gz -C build/bin .

build-linux-deb:
	CGO_ENABLED=1 wails build -clean -ldflags="-X 'main.CustomRoot=/opt/ton-torrent'"
	mkdir -p build/bin/ton-torrent/DEBIAN
	mkdir -p build/bin/ton-torrent/usr/local/bin
	mkdir -p build/bin/ton-torrent/opt/ton-torrent
	mkdir -p build/bin/ton-torrent/usr/share/applications
	cp ton-build/storage/storage-daemon/storage-daemon build/bin/ton-torrent/opt/ton-torrent/
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
	gon ./build/gon/config-storage.json
	gon ./build/gon/config.json

mac-arm64-all:
	make download-storage-mac-arm
	GOOS=darwin GOARCH=arm64 make build-mac
	sleep 1
	make sign-mac

mac-amd64-all:
	make download-storage-mac-amd
	GOOS=darwin GOARCH=amd64 make build-mac
	sleep 1
	make sign-mac

windows-amd64-all:
	make download-storage-windows-amd
	GOOS=windows GOARCH=amd64 make build-windows

linux-amd64-all:
	make download-storage-linux-amd
	GOOS=linux GOARCH=amd64 make build-linux-deb

linux-arm64-all:
	make download-storage-linux-arm
	GOOS=linux GOARCH=arm64 make build-linux-deb