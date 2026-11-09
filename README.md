# nsgo

A Go command-line tool to run ns-3 simulation files and retrieve the results.

## Features

- Runs ns-3 files from the command line
- Retrieves and reports the results of each run
- Built-in test mode (`-test=<path/to/test-conf-file>`) for batch testing
- Clear status output (info, success, and error messages)

## Requirements

- [Go](https://go.dev/dl/) `1.27.3` or newer
- [ns-3](https://www.nsnam.org/) installed and working on your machine

## Installation

```bash
git clone https://github.com/isik13exe/nsgo.git
cd nsgo
go mod tidy
go build -o nsgo .
```

Optionally, move the binary somewhere on your `PATH`:

```bash
sudo mv nsgo /usr/local/bin/
```

## Usage

```bash
nsgo <flags> <path/to/file.cc>
```

Examples:

```bash
# Run a simulation with flags
nsgo delay=<delay> bandwidth=<bandwidth> npacket=<n> pcapdir=<dir> packetsize=<size> simulation.cc

# Run a simulation with conf file
nsgo -conf=<path/to/conf> simulation.cc

# Run in batch test mode
nsgo -test=<path/to/test-conf-file> simulation.cc

# Show the help guide
nsgo -help
```
## Note

`<file>.cc` files must be put in the scratch directory of ns-3.
