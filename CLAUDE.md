# CLAUDE.md - Development Guide for AI Assistants

This document provides guidance for Claude (or other AI assistants) when working with the `konfluxctl` codebase.

## Project Overview

**konfluxctl** is a CLI tool for [Konflux](https://konflux-ci.dev/) that provides extra capabilities for working with Konflux clusters and resources.

- **Language**: Go 1.25.x
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Testing Framework**: [Ginkgo v2](https://onsi.github.io/ginkgo/) + [Gomega](https://onsi.github.io/gomega/)
- **License**: Apache 2.0

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
├── make/                  # Makefile includes
├── .github/workflows/     # CI/CD workflows
├── main.go               # Entry point
├── Makefile              # Build system
└── go.mod                # Go module definition
```

## Build System

The project uses a Makefile for common development tasks:

- `make install` - Build and install the binary to `./bin/konfluxctl` with version info
- `make test` - Run all tests with coverage (uses Ginkgo)
- `make fmt` - Format code using `go fmt`
- `make vet` - Run `go vet` for static analysis
- `make run-lint` - Run golangci-lint
- `make clean-cov` - Remove coverage reports

**Always run `make fmt` and `make vet` before committing code.**

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

### Package Organization

- **cmd/**: Command-line interface implementation using Cobra
  - Each command should have its own file
  - Complex subcommands go in subdirectories (e.g., `cmd/image/`)
- **internal/**: Internal packages not intended for external use
  - Group related functionality into logical packages
  - Keep utility functions separate from business logic

### Naming Conventions

- Use standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- Command files should match command names (e.g., `version.go` for version command)
- Test files follow `*_test.go` pattern
- Test suites use `*_suite_test.go` pattern

### Code Quality

- Run `go fmt` on all code (enforced by CI)
- Run `go vet` to catch common mistakes
- Use `golangci-lint` for comprehensive linting
- Maintain test coverage (coverage reports in `./coverage/`)

## Testing

### Test Framework

The project uses **Ginkgo v2** with **Gomega** matchers:

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

### Table-Driven Tests

Use `DescribeTable` for testing multiple scenarios:

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

### Running Tests

```bash
make test  # Runs all tests with randomization and coverage
```

Tests are run with:
- `--randomize-all` - Randomize spec execution order
- `--randomize-suites` - Randomize suite execution order
- Coverage for `./internal/...` and `./cmd/...`

### Test Organization

- Place test files next to the code they test
- Use `suite_test.go` files to set up Ginkgo test suites
- Group related tests in `Describe` blocks
- Use `Context` for different scenarios

## Adding New Commands

To add a new command:

1. Create a new file in `cmd/` or appropriate subdirectory
2. Define the command using Cobra:

```go
func myNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mycommand",
		Short: "Brief description",
		Long:  "Longer description",
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

3. Register the command in the appropriate parent (usually `cmd/root.go`):

```go
rootCmd.AddCommand(myNewCommand())
```

4. Add tests for the new command in `cmd/*_test.go`

## Logging

The project uses Go's standard `log/slog` package:

- Log level is controlled by the `-v/--verbose` flag
- Default level: `slog.LevelInfo`
- Verbose mode: `slog.LevelDebug`
- Set up in `cmd/root.go` PersistentPreRun

## Dependencies

Key dependencies:

- `github.com/spf13/cobra` - CLI framework
- `github.com/onsi/ginkgo/v2` - Testing framework
- `github.com/onsi/gomega` - Assertion/matcher library
- `github.com/konflux-ci/*` - Konflux API clients
- `k8s.io/*` - Kubernetes client libraries
- `sigs.k8s.io/controller-runtime` - Kubernetes controller runtime

Use `go mod tidy` to maintain clean dependencies.

## CI/CD

GitHub Actions workflows:

- **testing.yaml** - Runs tests and uploads coverage to CodeCov
- **code-style.yaml** - Enforces code formatting and linting
- **release.yaml** - Handles releases and binary distribution

All PRs must pass CI checks before merging.

## Development Workflow

1. **Before starting**: Make sure you're on latest main branch
2. **During development**:
   - Run `make fmt` frequently
   - Run `make test` to verify tests pass
   - Add tests for new functionality
3. **Before committing**:
   - Run `make fmt` and `make vet`
   - Run `make test` to ensure all tests pass
   - Consider running `make run-lint` for comprehensive checks
4. **Commit messages**: Use clear, descriptive commit messages
5. **PRs**: Ensure CI passes before requesting review

## Common Patterns

### Context Usage

Commands receive a context via `cmd.Context()` which is set up in the root command's `PersistentPreRun`.

### Error Handling

- Use `RunE` instead of `Run` in Cobra commands to return errors
- Return errors to the caller; the root command handles exit codes
- `SilenceUsage` is set to `true` to avoid showing usage on errors

### Kubernetes Client Access

The tool interacts with Konflux clusters (which are Kubernetes clusters). When working with Kubernetes resources:

- Use the controller-runtime client libraries
- Assume a kubeconfig session is available
- Commands may require active cluster connectivity

## Version Information

Version and Git SHA are injected at build time via ldflags:

```bash
LDFLAGS="-X 'github.com/eguzki/konfluxctl/cmd.gitSHA=$$GIT_HASH'
         -X 'github.com/eguzki/konfluxctl/cmd.version=$(VERSION)'"
```

See `Makefile` for the complete build process.

## Tips for AI Assistants

1. **Always include Apache 2.0 headers** in new Go files
2. **Use Ginkgo/Gomega** for tests, not the standard testing package
3. **Follow the existing command structure** when adding new commands
4. **Run `make fmt` and `make vet`** before suggesting code is complete
5. **Add tests** for any new functionality
6. **Check existing patterns** in similar commands before implementing new ones
7. **Use table-driven tests** when testing multiple scenarios
8. **Keep internal packages in `internal/`** to prevent external imports
9. **Use descriptive variable names** and add comments for complex logic
10. **Consider Kubernetes API conventions** when working with cluster resources

## Resources

- [Konflux Documentation](https://konflux-ci.dev/)
- [Cobra Documentation](https://cobra.dev/)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Development Guide](doc/development.md) (if exists)
- [Release Process](RELEASE.md)
