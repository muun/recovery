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
| Linux 32-bit | `4d93d4815e865b21e31e1060fcc5311b380a4ba2d0d53a6fb952850927019d8b` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux32) |
| Linux 64-bit | `cc9ac14315a0d8b9755b274e1b31f7fb2e16c25100af83f2f54af9be8e0af901` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux64) |
| Windows 32-bit | `081955009638c596af7193daaf8aae885dea12110f7620f3559bdf611224e393` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows32.exe) |
| Windows 64-bit | `4abbae9e94855e5cbf2ec44896671f26188128b3b0fdc8954cc01ba6dc533754` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows64.exe) |
| MacOS 64-bit | `af245b78175d74f315af339eebfe75d495d70cd80b038fd5f00f05d7247651b9` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-macos64) |

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

If you attempt to open the file directly, MacOS will block you from using it.

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
