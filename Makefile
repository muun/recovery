# (Default) build the Recovery Tool to run on this system.
build:
	mkdir -p bin
	
	echo "Building recovery tool"
	go build -a -trimpath -o "bin/recovery-tool"

	echo "Success! Built to bin/recovery-tool"

# Cross-compile and checksum the Recovery Tool for a range of OS/archs.
build-checksum-all: export DOCKER_BUILDKIT=1
build-checksum-all:
	# Get vendor dependencies:
	go mod vendor -v

	# Linux 32-bit:
	docker build . -o bin --build-arg os=linux --build-arg arch=386 --build-arg out=recovery-tool-linux32
	/bin/echo -n '✓ Linux 32-bit ' && sha256sum "bin/recovery-tool-linux32"

	# Linux 64-bit:
	docker build . -o bin --build-arg os=linux --build-arg arch=amd64 --build-arg out=recovery-tool-linux64
	/bin/echo -n '✓ Linux 64-bit ' && sha256sum "bin/recovery-tool-linux64"
	
	# Windows 32-bit:
	docker build . -o bin --build-arg os=windows --build-arg arch=386 --build-arg out=recovery-tool-windows32.exe
	/bin/echo -n '✓ Windows 32-bit ' && sha256sum "bin/recovery-tool-windows32.exe"
	
	# Windows 64-bit:
	docker build . -o bin --build-arg os=windows --build-arg arch=amd64 --build-arg out=recovery-tool-windows64.exe
	/bin/echo -n '✓ Windows 64-bit ' && sha256sum "bin/recovery-tool-windows64.exe"

	# Darwin 64-bit:
	docker build . -o bin --build-arg os=darwin --build-arg arch=amd64 --build-arg out=recovery-tool-macos64
	/bin/echo -n '✓ MacOS 64-bit ' && sha256sum "bin/recovery-tool-macos64"

.SILENT:
