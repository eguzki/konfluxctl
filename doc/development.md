# Development Guide

This guide describes how to build, test, and contribute to `konfluxctl`.

## Prerequisites

- **Go**: Version 1.25 or newer
- **Make**: For using the build system
- **Git**: For version control
- **golangci-lint**: (Optional) For comprehensive linting

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/eguzki/konfluxctl.git
cd konfluxctl
```

### Build the Project

```bash
# Build and install the binary to ./bin/konfluxctl
make install

# The binary will be available at:
./bin/konfluxctl version
```

To install system-wide:
```bash
sudo cp ./bin/konfluxctl /usr/local/bin/
```

## Development Workflow

### 1. Before Starting

- Ensure you're on the latest main branch:
  ```bash
  git checkout main
  git pull origin main
  ```

### 2. During Development

- **Format code frequently:**
  ```bash
  make fmt
  ```

- **Run tests to verify changes:**
  ```bash
  make test
  ```

- **Add tests for new functionality** - All new features should include appropriate test coverage

### 3. Before Committing

Run the following checks to ensure code quality:

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run all tests with coverage
make test

# (Optional) Run comprehensive linting
make run-lint
```

### 4. Commit Messages

Use clear, descriptive commit messages that explain what changed and why.

**Good examples:**
```
Add support for JSON output in image metadata command

Fix panic when image metadata is missing

Update dependencies to latest versions
```

### 5. Pull Requests

Before submitting a PR, ensure:
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No static analysis errors (`make vet`)
- [ ] No linting errors (`make run-lint`)
- [ ] Code includes appropriate test coverage
- [ ] Commit messages are clear and descriptive
- [ ] CI checks pass on GitHub

## Testing

### Test Framework

The project uses **Ginkgo v2** with **Gomega** matchers for testing.

### Running Tests

```bash
# Run all tests with coverage
make test

# Clean coverage reports
make clean-cov
```

Tests are executed with:
- `--randomize-all` - Randomize spec execution order
- `--randomize-suites` - Randomize suite execution order
- Coverage reports for `./internal/...` and `./cmd/...`

### Writing Tests

Place test files next to the code they test using the `*_test.go` pattern.

**Example using Ginkgo/Gomega:**

```go
import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("MyComponent", func() {
    It("should do something", func() {
        result := MyFunction()
        Expect(result).To(Equal(expected))
    })
})
```

**Table-driven tests:**

```go
var _ = DescribeTable("FunctionName",
    func(input string, expected bool) {
        result := MyFunction(input)
        Expect(result).To(Equal(expected))
    },
    Entry("description 1", "input1", true),
    Entry("description 2", "input2", false),
)
```

## Build System

The project uses Make for common development tasks:

| Command          | Description                                    |
| ---------------- | ---------------------------------------------- |
| `make install`   | Build and install binary to `./bin/konfluxctl` |
| `make test`      | Run all tests with coverage                    |
| `make fmt`       | Format code using `go fmt`                     |
| `make vet`       | Run `go vet` for static analysis               |
| `make run-lint`  | Run golangci-lint                              |
| `make clean-cov` | Remove coverage reports                        |

## Project Structure

```
konfluxctl/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command setup
│   ├── version.go         # Version command
│   ├── image.go           # Image command group
│   └── image/             # Image subcommands
│       └── metadata.go    # Image metadata command
├── internal/              # Internal packages (not for external use)
│   ├── utils/            # Utility functions
│   └── metadata/         # Metadata handling logic
├── doc/                   # Documentation
├── make/                  # Makefile includes
├── .github/workflows/     # CI/CD workflows
├── main.go               # Entry point
├── Makefile              # Build system
└── go.mod                # Go module definition
```

## Adding New Commands

To add a new command to `konfluxctl`:

### 1. Create the Command File

Create a new file in `cmd/` or appropriate subdirectory:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

func myNewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Brief description",
        Long:  "Longer description of what this command does",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }

    // Add flags
    cmd.Flags().StringVar(&myVar, "flag-name", "", "flag description")

    return cmd
}
```

### 2. Register the Command

Add it to the appropriate parent command (usually `cmd/root.go`):

```go
rootCmd.AddCommand(myNewCommand())
```

### 3. Add Tests

Create a corresponding test file `cmd/mycommand_test.go`:

```go
package cmd_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("MyCommand", func() {
    It("should work correctly", func() {
        // Test implementation
    })
})
```

### 4. Update Documentation

Update the README.md with the new command information.

## Code Style and Conventions

### File Headers

All Go source files must include the Apache 2.0 license header:

```go
/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
```

### Naming Conventions

- Use standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- Command files should match command names (e.g., `version.go` for version command)
- Test files follow `*_test.go` pattern
- Test suites use `*_suite_test.go` pattern

### Package Organization

- **cmd/**: Command-line interface implementation using Cobra
  - Each command should have its own file
  - Complex subcommands go in subdirectories (e.g., `cmd/image/`)
- **internal/**: Internal packages not intended for external use
  - Group related functionality into logical packages
  - Keep utility functions separate from business logic

## Troubleshooting

### Common Development Issues

#### **"Error: unable to connect to cluster" during testing**
- Some commands require an active Kubernetes connection
- Ensure your kubeconfig is properly configured
- Verify you have access to a Konflux cluster
- Check that `kubectl` can connect: `kubectl cluster-info`

#### **Tests fail with "ginkgo: command not found"**
- Install Ginkgo CLI:
  ```bash
  go install github.com/onsi/ginkgo/v2/ginkgo@latest
  ```

#### **Build fails with version information**
- The build system injects version info via ldflags
- Use `make install` instead of `go build` directly

#### **Linting errors**
- Install golangci-lint:
  ```bash
  # On macOS
  brew install golangci-lint

  # On Linux
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

### Getting Help

- Check existing issues: https://github.com/eguzki/konfluxctl/issues
- Review the [CLAUDE.md](../CLAUDE.md) for AI-assisted development guidance
- Ask questions in pull request discussions

## Continuous Integration

GitHub Actions workflows run on all pull requests:

- **testing.yaml**: Runs tests and uploads coverage to CodeCov
- **code-style.yaml**: Enforces code formatting and linting
- **release.yaml**: Handles releases and binary distribution

All CI checks must pass before a PR can be merged.

## Dependencies

Key dependencies:

- `github.com/spf13/cobra` - CLI framework
- `github.com/onsi/ginkgo/v2` - Testing framework
- `github.com/onsi/gomega` - Assertion/matcher library
- `github.com/konflux-ci/*` - Konflux API clients
- `k8s.io/*` - Kubernetes client libraries
- `sigs.k8s.io/controller-runtime` - Kubernetes controller runtime

### Managing Dependencies

```bash
# Add a new dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```

## Release Process

See [RELEASE.md](../RELEASE.md) for information about the release process.

## Additional Resources

- [Konflux Documentation](https://konflux-ci.dev/)
- [Cobra Documentation](https://cobra.dev/)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
