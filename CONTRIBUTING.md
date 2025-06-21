# Contributing to Tarmac

Thank you for considering contributing to Tarmac. The time, skills, and perspectives you contribute to this project are valued.

## How can I contribute?

Bugs, design proposals, feature requests, and questions are all welcome and can be submitted by creating a [Github Issue](https://github.com/tarmac-project/tarmac/issues/new/choose) using one of the templates provided. Please provide as much detail as you can.

Code contributions are welcome as well! In an effort to keep this project tidy, please:

- Use `go mod` to install and lock dependencies
- Use `gofmt` to format code and tests
- Run `go vet -v ./...` to check for any inadvertent suspicious code
- Write and run unit tests when they make sense using `go test`

## Commit Message Guidelines
# Commit Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/) specifications.

## Format

Each commit message consists of a **header**, an optional **body**, and an optional **footer**.

```
<type>(<optional scope>): <subject>

<optional body>

<optional footer>
```

The **header** is mandatory and must conform to the [Commit Message Header](#commit-message-header) format.

### Commit Message Header

```
<type>(<optional scope>): <subject>
```

The `type` and `subject` fields are mandatory, the `scope` field is optional.

#### Type

Must be one of the following:

* **feat**: A new feature
* **fix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **perf**: A code change that improves performance
* **test**: Adding missing tests or correcting existing tests
* **build**: Changes that affect the build system or external dependencies
* **ci**: Changes to our CI configuration files and scripts
* **chore**: Other changes that don't modify src or test files

#### Scope

The scope specifies what area of the codebase your commit touches. For example:

* **core**: Changes to the main tarmac package
* **app**: Changes to the app package
* **callbacks**: Changes to the callbacks package
* **config**: Changes to the config package
* **sanitize**: Changes to the sanitize package
* **wasm**: Changes to the wasm package
* **sdk**: Changes to the SDK packages
* **telemetry**: Changes to the telemetry package
* **tlsconfig**: Changes to the TLS configuration
* **deps**: Changes to dependencies

#### Subject

The subject is a succinct description of the change:

* Use the imperative, present tense: "change" not "changed" nor "changes"
* Don't capitalize the first letter
* No dot (.) at the end

### Commit Message Body

The body should include the motivation for the change and contrast this with previous behavior.

### Commit Message Footer

The footer should contain any information about **Breaking Changes** and is also the place to reference GitHub issues that this commit closes.

**Breaking Changes** should start with the word `BREAKING CHANGE:` with a space or two newlines. The rest of the commit message is then used for this.

## Examples

```
feat(callbacks): add support for HTTP client timeouts

Add configurable timeout settings for HTTP client callbacks to prevent indefinite waits

Closes #123
```

```
fix(wasm): resolve memory leak in function instances

When a WASM function fails to initialize properly, the allocated resources
were not being correctly cleaned up, causing memory leaks.

BREAKING CHANGE: The WASM function interface now requires explicit Close()
```

```
docs: update installation instructions
```

```
style(app): fix comment typos
```

```
refactor(sdk): optimize parameter handling
```
