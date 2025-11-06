# KONFLUXCTL
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

`konfluxctl` is a CLI tool for [Konflux](https://konflux-ci.dev/) to provide extra capabilities.

## Installing

`konfluxctl` can be installed either by downloading pre-compiled binaries or by compiling from source. For most users, downloading the binary is the easiest and recommended method.

### Installing Pre-compiled Binaries

1. Download the latest binary for your platform from the [`konfluxctl` Releases](https://github.com/eguzki/konfluxctl/releases) page.
2. Unpack the binary.
3. Move it to a directory in your `$PATH` so that it can be executed from anywhere.

### Compiling from Source

If you prefer to compile from source or are contributing to the project, you can install `konfluxctl` using  `make install`. This method requires Golang 1.25 or newer.

It is possible to use the make target `install` to compile from source. From root of the repository, run 

```bash
make install
```

This will compile `konfluxctl` and install it in the `bin` directory at root of directory. It will also ensure the correct version of the binary is displayed . It can be ran using `./bin/konfluxctl` .  

## Usage

 Below is a high-level overview of its commands, along with links to detailed documentation for more complex commands.

### General Syntax

```bash
konfluxctl [command] [subcommand] [flags]
```

### Commands Overview

| Command      | Description                                                |
| ------------ | ---------------------------------------------------------- |
| `image`      | Docker/OCI image related subcommands                       |
| `help`       | Help about any command                                     |
| `completion` | Generate autocompletion script for the  specific shell     |
| `version`    | Print the version number of `konfluxctl`                   |

#### `image`

| Command      | Description                                                |
| ------------ | ---------------------------------------------------------- |
| `metadata`      | Inspect Konflux metadata of the image                   |

### Flags

| Flag               | Description           |
| ------------------ | --------------------- |
| `-h`, `--help`     | Help for `konfluxctl`  |
| `-v`, `--verbose`  | Enable verbose output |

### Commands Detail

#### `completion`

Generate an autocompletion script for the specified shell.

| Subcommand   | Description                                 |
| ------------ | ------------------------------------------- |
| `bash`       | Generate script for Bash                    |
| `fish`       | Generate script for Fish                    |
| `powershell` | Generate script for PowerShell              |
| `zsh`        | Generate script for Zsh                     |

#### `metadata`

### Usage

```shell
$ konfluxctl image metadata -h
Returns Docker/OCI image related konflux metadata

Usage:
  konfluxctl image metadata [flags]

Flags:
  -h, --help                   help for metadata
      --image string           Docker/OCI image URL (required)
  -o, --output-format string   Output format: 'yaml' or 'json'.

Global Flags:
  -v, --verbose   verbose output
```

>NOTE: requires kube session open to konflux cluster

#### `version`

Print the version number of `konfluxctl`.

No additional flags or subcommands.

## Using with GitHub Actions

```yaml
- name: Install konfluxctl
  uses: jaxxstorm/action-install-gh-release@v1.10.0
  with: # Grab the latest version
    repo: eguzki/konfluxctl
```

## Contributing
The [Development guide](doc/development.md) describes how to build the konfluxctl CLI and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
