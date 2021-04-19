![muun](https://muun.com/images/github-banner-v2.png)

## Recovery Tool

Welcome!

You can use this tool to transfer all funds out of your Muun account to an address of your choosing.

![](readme/demo.gif)

**This process requires no collaboration from Muun to work**. We wholeheartedly believe that self-custodianship
is an essential right, and we want to create a world in which people have complete and exclusive
control over their own money. Bitcoin has finally made this possible.

## Usage

Download the appropriate binary in the following table, according to your operating system and
architecture.

| System | Checksum | Link |
| --- | --- | --- |
| Linux 32-bit | `65c0e27bcff10210f5637a8b9f95ffd8c932d258c21d23d5d9da40ba091864a3` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux32) |
| Linux 64-bit | `596c819d22501e267385325dd2bba7e5260f711eb3d210c468a606699c8d8369` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-linux64) |
| Windows 32-bit | `897ff4db5ccc7f5b37c9c479f018b5ba4a98a243137f186fbf4b96138eff6adc` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows32.exe) |
| Windows 64-bit | `c03e981119c18270d517d74691283fd3e4d57460d1bf02189c7552b8daa06625` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-windows64.exe) |
| MacOS 64-bit | `c5b5d0f65f6b0a1a98bcbf405b50a691b33c347b06b02af98d3350bddb9353f3` | [Download](https://raw.githubusercontent.com/muun/recovery/master/bin/recovery-tool-macos64) |

Once you have that, run:

```
./recovery-tool <path to your Emergency Kit PDF>
```

The process takes only a few minutes (depending on your connection).

If you have any questions, we'll be happy to answer them. Contact us at support@muun.com

## Auditing

This tool is open-sourced so that auditors can dive into the code, and verify it to their benefit
and everyone else's. We encourage people with the technical knowledge to do this.

To build the tool locally and run it, you must:

1. Install the [Go](https://golang.org/) toolchain.
2. Clone the repository:

    ```
    git clone https://github.com/muun/recovery
    cd recovery
    ```
      
3. Run the tool with:

    ```
    go run -mod=vendor . -- <path to your Emergency Kit PDF>
    ```

To build the tool in all its variants and verify the checksums for the above binaries, you need to:

1. Install the [Docker](https://www.docker.com/) toolchain and start the daemon.
2. Run this command:

    ```
    make build-checksum-all
    ```

3. Verify that the printed checksums match those of the downloaded versions, using `sha256sum` 
as in the `Makefile`.

We use Docker for these builds to ensure they are reproducible.


## Questions

If you have any questions, we'll be happy to answer them. Contact us at contact@muun.com


## Responsible Disclosure

Send us an email to report any security related bugs or vulnerabilities at [security@muun.com](mailto:security@muun.com).

You can encrypt your email message using our public PGP key.

Public key fingerprint: `1299 28C1 E79F E011 6DA4 C80F 8DB7 FD0F 61E6 ED76`
