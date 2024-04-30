# nvmrc-gen

`nvmrc-gen` is a CLI utility that generates a `.nvmrc` file in a Node project for use with `nvm`. Its primary purpose is for use in CI/CD pipelines, but it can also be useful for developers generating `.nvmrc` files for their own use.


## Usage

Add `nvmrc-gen` to your `PATH` or copy it directly where it is intended to be ran, then run it there. It uses the working directory of the shell it is ran from.

<img width="976" alt="image" src="https://github.com/MBDesu/nvmrc-gen/assets/39097222/d1cfc51d-4436-4878-b8e9-e776bd152e9b">


### Flags

`-c`: CI mode. Don't prompt for writing of files.
`-s`: Silent mode. Output no logs.


## Building

Clone this repository and run `go build -ldflags="-w -s" -gcflags=all=-l -o /path/to/output/nvmrc-gen` to build for your architecture.

To build for other architectures, run `env GOOS=<OS> GOARCH=<arch> go build -ldflags="-w -s" -gcflags=all=-l -o /path/to/output/nvmrc-gen`.

The following builds are available in each release:

|    Target      | `amd64` | `arm` | `arm64` |
| -------------: | :-----: | :---: | :-----: |
| macOS (darwin) |   ✅    |   ❌   |    ✅   |
| FreeBSD        |   ✅    |   ✅   |    ✅   |
| Linux          |   ✅    |   ✅   |    ✅   |
| NetBSD         |   ✅    |   ✅   |    ✅   |
| OpenBSD        |   ✅    |   ✅   |    ✅   |
| Windows        |   ✅    |   ✅   |    ✅   |

The .zip file for each OS and arch is formatted as `nvmrc-gen-<OS>-<arch>-<version>.zip`. For example, the `amd64` for macOS build of `nvmrc-gen` v0.1.0 is named `nvmrc-gen-darwin-amd64-0.1.0.zip`.


## Contributing

Feel free to contribute in any way you like. This was my first project in Go, and it shows.


### TODO

- [ ] Better error handling
