# (Default) build the Recovery Tool to run on this system.
build:
	mkdir -p bin
	
	echo "Building recovery tool"
	go build -mod=vendor -a -trimpath -o "bin/recovery-tool"

	echo "Success! Built to bin/recovery-tool"

# Cross-compile and checksum the Recovery Tool for a range of OS/archs.
build-checksum-all: export DOCKER_BUILDKIT=1
build-checksum-all:
	# Get vendor dependencies:
	go work vendor -v

	# Linux 32-bit:
	docker build -f Dockerfile .. -o bin \
		--build-arg os=linux \
		--build-arg arch=386 \
		--build-arg cc=i686-linux-gnu-gcc-12 \
		--build-arg out=recovery-tool-linux32

	/bin/echo -n '✓ Linux 32-bit ' && sha256sum "bin/recovery-tool-linux32"

	# Linux 64-bit:
	docker build -f Dockerfile .. -o bin \
		--build-arg os=linux \
		--build-arg arch=amd64 \
		--build-arg cc=x86_64-linux-gnu-gcc-12 \
		--build-arg out=recovery-tool-linux64

	/bin/echo -n '✓ Linux 64-bit ' && sha256sum "bin/recovery-tool-linux64"

	# Linux arm64:
	docker build -f Dockerfile .. -o bin \
		--build-arg os=linux \
		--build-arg arch=arm64 \
		--build-arg cc=aarch64-linux-gnu-gcc-12 \
		--build-arg out=recovery-tool-linuxaarch64

	/bin/echo -n '✓ Linux arm64' && sha256sum "bin/recovery-tool-linuxaarch64"

	# Windows 32-bit:
	docker build -f Dockerfile .. -o bin \
		--build-arg os=windows \
		--build-arg arch=386 \
		--build-arg cc=i686-w64-mingw32-gcc \
		--build-arg out=recovery-tool-windows32.exe

	/bin/echo -n '✓ Windows 32-bit ' && sha256sum "bin/recovery-tool-windows32.exe"
	
	# Windows 64-bit:
	docker build -f Dockerfile .. -o bin \
		--build-arg os=windows \
		--build-arg arch=amd64 \
		--build-arg cc=x86_64-w64-mingw32-gcc \
		--build-arg out=recovery-tool-windows64.exe

	/bin/echo -n '✓ Windows 64-bit ' && sha256sum "bin/recovery-tool-windows64.exe"

	# NOTE:
	# Darwin reproducible builds are disabled for now, since the inclusion of C code in the latest
	# release made building the tool inside a Linux container extremely difficult. We'll be moving the
	# process to GitHub actions, where we can build on MacOS.

	# Darwin 64-bit:
	# docker build . -o bin \
	# 	--build-arg os=darwin \
	# 	--build-arg arch=amd64 \
	# 	--build-arg out=recovery-tool-macos64

	# /bin/echo -n '✓ MacOS 64-bit ' && sha256sum "bin/recovery-tool-macos64"

.SILENT:
