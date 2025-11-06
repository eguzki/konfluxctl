# konfluxctl

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/eguzki/konfluxctl)](https://goreportcard.com/report/github.com/eguzki/konfluxctl)
[![codecov](https://codecov.io/gh/eguzki/konfluxctl/branch/main/graph/badge.svg)](https://codecov.io/gh/eguzki/konfluxctl)

A powerful CLI tool for [Konflux](https://konflux-ci.dev/) that extends and enhances your Konflux workflow with additional capabilities for managing images, inspecting metadata, and more.

## Features

- **Image Metadata Inspection**: Extract and inspect Konflux metadata from Docker/OCI images

## Installation

### Prerequisites

- For building from source: Go 1.25 or newer

**Note:** Some commands (like `image metadata`) require an active kubeconfig session connected to a Konflux cluster.

#### Option 1: Pre-compiled Binaries (Recommended)

1. Download the latest binary for your platform from the [Releases](https://github.com/eguzki/konfluxctl/releases) page
2. Extract the archive:
   ```bash
   tar -xzf konfluxctl_<version>_<os>_<arch>.tar.gz
   ```
3. Move the binary to a directory in your `$PATH`:
   ```bash
   sudo mv konfluxctl /usr/local/bin/
   ```
4. Verify the installation:
   ```bash
   konfluxctl version
   ```

#### Option 2: Build from Source

If you prefer to compile from source or are contributing to the project:

```bash
# Clone the repository
git clone https://github.com/eguzki/konfluxctl.git
cd konfluxctl

# Build and install
make install

# The binary will be available at ./bin/konfluxctl
./bin/konfluxctl version
```

To install system-wide:
```bash
sudo cp ./bin/konfluxctl /usr/local/bin/
```

## Quick Start

1. Ensure you have a valid kubeconfig pointing to your Konflux cluster
2. Inspect metadata from a Konflux-built image:
   ```bash
   konfluxctl image metadata --image quay.io/your-org/your-image@sha256:abcd1234...
   ```
3. Get output in JSON format:
   ```bash
   konfluxctl image metadata --image quay.io/your-org/your-image@sha256:abcd1234... -o json
   ```  

## Usage

### General Syntax

```bash
konfluxctl [command] [subcommand] [flags]
```

### Global Flags

| Flag              | Short | Description                    |
| ----------------- | ----- | ------------------------------ |
| `--help`          | `-h`  | Display help for any command   |
| `--verbose`       | `-v`  | Enable verbose/debug output    |

### Available Commands

| Command      | Description                                         |
| ------------ | --------------------------------------------------- |
| `image`      | Docker/OCI image related operations                 |
| `version`    | Print the version number of konfluxctl              |
| `completion` | Generate shell autocompletion scripts               |
| `help`       | Display help information for any command            |

---

### Command Details

#### `version`

Display the current version and build information.

```bash
konfluxctl version
```

**Output:**
```
konfluxctl version: v0.1.0
Git SHA: abc1234
```

#### `image`

Manage and inspect Docker/OCI images in Konflux.

**Subcommands:**

##### `image metadata`

Retrieve and display Konflux metadata from a Docker/OCI image.

**Usage:**
```bash
konfluxctl image metadata --image <image-url> [flags]
```

**Flags:**
| Flag              | Description                                  | Required |
| ----------------- | -------------------------------------------- | -------- |
| `--image`         | Docker/OCI image URL                         | Yes      |
| `-o`, `--output-format` | Output format: `yaml` or `json`        | No       |

**Note:** Requires an active kubeconfig session connected to a Konflux cluster.

**Examples:**
```bash
# Display metadata in YAML format (default)
konfluxctl image metadata --image quay.io/konflux-ci/my-app@sha256:a1b2c3d4e5f67890...

# Display metadata in JSON format
konfluxctl image metadata --image quay.io/konflux-ci/my-app@sha256:a1b2c3d4e5f67890... -o json

# Use verbose mode for debugging
konfluxctl image metadata --image quay.io/konflux-ci/my-app@sha256:a1b2c3d4e5f67890... --verbose
```

#### `completion`

Generate shell autocompletion scripts to enhance your CLI experience.

**Supported Shells:**

| Shell        | Command                              |
| ------------ | ------------------------------------ |
| Bash         | `konfluxctl completion bash`         |
| Zsh          | `konfluxctl completion zsh`          |
| Fish         | `konfluxctl completion fish`         |
| PowerShell   | `konfluxctl completion powershell`   |

**Setup Instructions:**

**Bash:**
```bash
# Linux
konfluxctl completion bash | sudo tee /etc/bash_completion.d/konfluxctl

# macOS
konfluxctl completion bash > /usr/local/etc/bash_completion.d/konfluxctl
```

**Zsh:**
```bash
konfluxctl completion zsh > "${fpath[1]}/_konfluxctl"
```

**Fish:**
```bash
konfluxctl completion fish > ~/.config/fish/completions/konfluxctl.fish
```

**PowerShell:**
```powershell
konfluxctl completion powershell | Out-String | Invoke-Expression
```

---

## Examples

### Inspecting Image Metadata

Get comprehensive metadata about a Konflux-built image:

```bash
$ konfluxctl image metadata --image quay.io/my-org/my-app@sha256:f1e2d3c4b5a67890abcdef1234567890abcdef1234567890abcdef1234567890
```

**Sample Output (YAML):**
```yaml
buildPipelineName: docker-build
componentName: my-app
applicationName: my-application
pipelineRunName: my-app-build-12345
...
```

### Using Different Output Formats

```bash
# YAML format (default)
konfluxctl image metadata --image quay.io/my-org/my-app@sha256:f1e2d3c4b5a67890abcdef1234567890abcdef1234567890abcdef1234567890

# JSON format for programmatic processing
konfluxctl image metadata --image quay.io/my-org/my-app@sha256:f1e2d3c4b5a67890abcdef1234567890abcdef1234567890abcdef1234567890 -o json | jq .

# JSON format piped to a file
konfluxctl image metadata --image quay.io/my-org/my-app@sha256:f1e2d3c4b5a67890abcdef1234567890abcdef1234567890abcdef1234567890 -o json > metadata.json
```

### Debugging with Verbose Mode

Enable verbose logging to troubleshoot issues:

```bash
konfluxctl image metadata --image quay.io/my-org/my-app@sha256:f1e2d3c4b5a67890abcdef1234567890abcdef1234567890abcdef1234567890 --verbose
```

## GitHub Actions Integration

Integrate `konfluxctl` into your GitHub Actions workflows:

```yaml
name: Inspect Image Metadata

on:
  push:
    branches: [main]

jobs:
  inspect:
    runs-on: ubuntu-latest
    steps:
      - name: Install konfluxctl
        uses: jaxxstorm/action-install-gh-release@v1.10.0
        with:
          repo: eguzki/konfluxctl

      - name: Configure kubeconfig
        run: |
          # Set up your Konflux cluster credentials
          echo "${{ secrets.KUBECONFIG }}" > kubeconfig.yaml
          export KUBECONFIG=kubeconfig.yaml

      - name: Inspect image metadata
        run: |
          konfluxctl image metadata --image quay.io/my-org/my-app:${{ github.sha }} -o json
```

## Contributing

We welcome contributions! Please see the [Development Guide](doc/development.md) for comprehensive information on:

- Building and testing the project
- Development workflow and best practices
- Adding new commands
- Code style and conventions
- Troubleshooting common issues

For quick reference:
```bash
# Clone and build
git clone https://github.com/eguzki/konfluxctl.git
cd konfluxctl
make install

# Before submitting a PR
make fmt && make vet && make test
```

## License

This software is licensed under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).

See the [LICENSE](LICENSE) and [NOTICE](NOTICE) files for details.

---

**Links:**
- [Konflux Documentation](https://konflux-ci.dev/)
- [Issue Tracker](https://github.com/eguzki/konfluxctl/issues)
- [Releases](https://github.com/eguzki/konfluxctl/releases)
