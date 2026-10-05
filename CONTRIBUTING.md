# Contributing to NACRE

## Getting Started

1. Clone the repository
2. Install prerequisites: Go 1.26+, Rust toolchain, C compiler, [Task](https://taskfile.dev)
3. Build: `task build:blockchain`
4. Test: `task test:go`

## Submitting Changes

1. Create a branch from `main`
2. Keep PRs focused: one fix or feature per PR
3. Run `task fmt tidy` before pushing
4. All CI checks must pass
5. Bug fixes should include a test that reproduces the issue

## Commit Messages

Use the format: `type(scope): description`

```
fix(node): correct block validation for edge case
feat(node): add a mining RPC
docs: update build instructions
```

## Sign-Off

By submitting a PR you certify that your contribution is your own work
and you have the right to submit it under the project's ISC License.

## Security

Do **not** open public issues for security vulnerabilities.
See [SECURITY.md](SECURITY.md).
