# first-go — Autofixer module

Autofixer module for a thesis project implemented in Go. This component is currently standalone and intended to integrate with the ADES detection system later. It provides tools and workflows for generating, validating and applying automated fixes for discovered issues.

## Overview

This repository contains a small Go codebase used to develop and test autofix logic outside the ADES detector. It includes:

- A manual runner (main.go) to exercise different fix patches and validate behavior.
- A GitHub Actions workflow to run autofix tasks as CI automation.
- Directories separating synthetic and real-world vulnerability workflows.

## Layout

- autofix/            — autofix functions and patch application logic
- vuln/               — synthetic vulnerability database and test fixtures
- security-incident/  — real-life vulnerability workflow examples and incident artifacts
- README.md           — this file
- LICENSE             — project license
- go.mod              — Go module definition

## Requirements

- Go 1.20+
- Make (optional)
- Access to any external tools listed in docs/ for running incident workflows

## Quickstart

Run locally (manual testing):

1. Clone:
    git clone <repo-url>
2. Enter project:
    cd first-go
3. Download deps:
    go mod download
4. Build:
    go build ./...
5. Run manual runner:
    go run main.go
    * Manual config of the different scenarios and fix is require for function to work as intended. See main.go for details.

## Testing

- Unit tests:
    Example: 
        go test -v -run ^TestADES102Fix$

## Versioning & Contributing

- Use git tags for releases. Follow semantic versioning where appropriate.
- Open issues for bugs and feature requests. Provide tests for new functionality and describe expected behavior for fix patches.

## License & Contact

- Add a LICENSE file at the repository root (MIT, Apache 2.0, or institutional license).
- Include AUTHORS or CONTRIBUTORS for contact details and advisor information.

Customize this README to reflect integration steps with the ADES system and the exact workflow your thesis requires.