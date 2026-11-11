# Project Instructions

## Project
TradeScope — учебный проект по исследованию алгоритмических торговых стратегий.

## Main instructions
- Follow master_prompt.md as the primary implementation specification.
- Read PROJECT_CONTEXT.md if it exists.
- Preserve existing project files and documentation.
- Do not invent experimental results, data sources or bibliography.
- Keep the architecture simple and idiomatic.
- Use Go 1.27 as the target version.
- Write code that a third-year CS student can understand and explain.

## Development rules
- Run gofmt on modified Go files.
- Write unit tests for mathematical algorithms.
- Handle errors explicitly.
- Avoid unnecessary dependencies and abstractions.
- Keep documentation consistent with actual implementation.
- Compile the Typst report after making changes.

## Before finishing
- Run go test ./...
- Run go vet ./...
- Run go build ./...
- Verify Typst compilation.
- Update README.md.