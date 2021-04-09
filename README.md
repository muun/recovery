![muun](https://muun.com/images/github-banner-v2.png)

## Recovery Tool

Welcome!

You can use this tool to transfer all funds from your Muun wallet to an address of your choosing.

![](readme/demo.gif)

**This process requires no collaboration from Muun to work**. We wholeheartedly believe that self-custodianship
is an essential right, and we want to create a world in which people have complete and exclusive
control over their own money. Bitcoin has finally made this possible.

## Usage

To execute a recovery, you will need:

1. **Your Recovery Code**, which you wrote down during your security setup
2. **Your Emergency Kit PDF**, which you exported from the app
3. **Your destination bitcoin address**, where all your funds will be sent

Once you have that, you must:

1. Install [golang](https://golang.org/)
2. Open a terminal window
3. Run:

        git clone https://github.com/muun/recovery
        cd recovery
        ./recovery-tool <path to your Emergency Kit PDF>

The recovery process takes only a few minutes  (depending on your connection).

## Questions

If you have any questions, we'll be happy to answer them. Contact us at [contact@muun.com](mailto:contact@muun.com)

## Auditing

Begin by reading `main.go`, and follow calls to other files and modules as you see fit. We always work
to improve code quality and readability with each release, so that auditing is easier and more effective.

The low-level encryption, key handling and transaction crafting code can be found in the `libwallet`
module, and it's the same our iOS and Android applications use.


## Responsible Disclosure

Send us an email to report any security related bugs or vulnerabilities at [security@muun.com](mailto:security@muun.com).

You can encrypt your email message using our public PGP key.

Public key fingerprint: `1299 28C1 E79F E011 6DA4 C80F 8DB7 FD0F 61E6 ED76`