![muun](https://muun.com/images/github-banner-v2.png)

## About

You can use this tool to swipe all the funds in your muun account to an address of your choosing.

To do this you will need:
* The recovery code, that you set up when you created your muun account
* The two encrypted private keys that you exported from your muun wallet
* A destination Bitcoin address where all your funds will be sent

## Setup

1. Clone this repository
2. Install [golang](https://golang.org/)
3. Run the tool with the following line: `go run -mod=vendor .`

## Questions

If you have any questions, contact us at contact@muun.com

## Auditing

* Most of the key handling and transaction crafting operations happens in the **libwallet** module.
* All the blockchain scan code is in the **neutrino** module.

## Responsible Disclosure

Send us an email to report any security related bugs or vulnerabilities at [security@muun.com](mailto:security@muun.com).

You can encrypt your email message using our public PGP key.

Public key fingerprint: `1299 28C1 E79F E011 6DA4 C80F 8DB7 FD0F 61E6 ED76`
