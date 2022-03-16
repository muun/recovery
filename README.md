![muun](https://muun.com/images/github-banner-v2.png)

Welcome!

You can use this Recovery Tool to transfer all funds out of your Muun account to an address 
of your choosing.

![](readme/demo.gif)

**This process requires no collaboration from Muun to work**. We wholeheartedly believe that self-custodianship
is an essential right, and we want to create a world in which people have complete and exclusive
control over their own money. Bitcoin has finally made this possible.

## Usage

Download the appropriate binary from the following table (or see [`BUILD.md`](BUILD.md) to build it yourself),
and follow the instructions below.

| System | Checksum | Link |
| --- | --- | --- |
| Linux 32-bit | `62ad190a48e56a7f54341fe89456622c233c3d9f9f1e536ae768014665b0ffad` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux32) |
| Linux 64-bit | `4d413acabbbd4d9be614edb8e492a5100daae8b8ff88beb5db9e862d1bd17b70` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux64) |
| Windows 32-bit | `82d5becec90c145cc01cce819cca7a53ebd5a16be76f5e4619be840dcc0f2b1a` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows32.exe) |
| Windows 64-bit | `a8d668a05006899dfd5e43c69871ddf40c126fb2f578865a92f9fd32f77824c1` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows64.exe) |
| MacOS 64-bit | `ef801b78b4989b0028161386c69def5dffeeea7b9d167fc94c665a06b8fc00fd` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-macos64) |

### Windows

Open the downloaded file. You'll be warned that the executable is not from a Microsoft-verified
source. Click `More info`, and then `Run anyway`.


### MacOS

Download the file to a known location (say `Downloads` in your Home directory), then open a terminal
and run:

```
cd ~/Downloads
chmod +x recovery-tool-macos64
./recovery-tool-macos64 <path to your Emergency Kit PDF>
```

#### Security Warnings

MacOS may prevent you from running the downloaded tool, depending on the active security settings. If it
does, head to System Preferences, then Security and Privacy, and give permission for the tool to run at the
bottom of the General tab.

### Linux

Download the file to a known location (say `Downloads` in your Home directory), then open a terminal
and run:

```
cd ~/Downloads
chmod +x recovery-tool-linux64
./recovery-tool-linux64 <path to your Emergency Kit PDF>
```

Use the `linux32` binary if appropriate.

### Questions?

If you have any questions, we'll be happy to answer them. Contact us at [support@muun.com](mailto:support@muun.com).


## Auditing and Veryfing

This tool is open-sourced so that auditors can inspect the code, build their own binaries and 
verify them to their benefit and everyone else's. We encourage people with the technical knowledge 
to do this.

To audit the source code, we suggest you start at `main.go` and navigate your way from there. We 
always work to improve code quality and readability with each release, so that auditing is easier 
and more effective.

To build and verify the reproducible binaries we provide, see [`BUILD.md`](BUILD.md).

## Responsible Disclosure

Send us an email to report any security related bugs or vulnerabilities at [security@muun.com](mailto:security@muun.com).

You can encrypt your email message using our public PGP key.

Public key fingerprint: `1299 28C1 E79F E011 6DA4 C80F 8DB7 FD0F 61E6 ED76`
