# Building and Auditing

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

To build the tool in all its variants and verify the checksums for the provided binaries, you need to:

1. Install the [Docker](https://www.docker.com/) toolchain and start the daemon.
2. Run this command:

    ```
    make build-checksum-all
    ```

3. Verify that the printed checksums match those of the downloaded versions, using `sha256sum` 
as in the `Makefile`.

We use Docker for these builds to ensure they are reproducible.