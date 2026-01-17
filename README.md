# Negotiation

[![tag](https://img.shields.io/github/tag/talav/negotiation.svg)](https://github.com/talav/negotiation/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/talav/negotiation.svg)](https://pkg.go.dev/github.com/talav/negotiation)
[![Go Report Card](https://goreportcard.com/badge/github.com/talav/negotiation)](https://goreportcard.com/report/github.com/talav/negotiation)
[![CI](https://github.com/talav/negotiation/actions/workflows/negotiation-ci.yml/badge.svg)](https://github.com/talav/negotiation/actions)
[![codecov](https://codecov.io/gh/talav/negotiation/graph/badge.svg)](https://codecov.io/gh/talav/negotiation)
[![License](https://img.shields.io/github/license/talav/negotiation)](./LICENSE)

Go library for HTTP content negotiation based on RFC 7231 with support for media types, languages, charsets, and encodings. Provides comprehensive content negotiation tools for building robust HTTP services.

**Requires Go 1.25 or later.**

## Features

- **Media Type Negotiation** - Negotiate based on `Accept` headers
- **Language Negotiation** - Negotiate based on `Accept-Language` headers
- **Charset Negotiation** - Negotiate based on `Accept-Charset` headers
- **Encoding Negotiation** - Negotiate based on `Accept-Encoding` headers
- **RFC 7231 Compliant** - Follows HTTP content negotiation standards
- **Quality Value Support** - Handles q-values for preference ordering
- **Wildcard Support** - Supports wildcard matching (`*/*`, `text/*`, etc.)
- **Parameter Matching** - Matches media type parameters (e.g., `charset=UTF-8`)
- **Plus-Segment Matching** - Supports media types with plus segments (e.g., `application/vnd.api+json`)

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/talav/negotiation"
)

func main() {
    // Create a media type negotiator
    negotiator := negotiation.NewMediaNegotiator()

    // Negotiate based on Accept header
    acceptHeader := "text/html, application/json;q=0.9, */*;q=0.8"
    priorities := []string{"application/json", "text/html"}

    best, err := negotiator.GetBest(acceptHeader, priorities, false)
    if err != nil {
        panic(err)
    }

    if best != nil {
        fmt.Printf("Best match: %s\n", best.Type)
        // Output: Best match: text/html
    }
}
```

## Installation

```bash
go get github.com/talav/negotiation
```

## Usage

### Media Type Negotiation

```go
package main

import (
    "fmt"
    "github.com/talav/negotiation"
)

func main() {
    negotiator := negotiation.NewMediaNegotiator()
    
    acceptHeader := "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8"
    priorities := []string{"application/json", "application/xml", "text/html"}
    
    best, err := negotiator.GetBest(acceptHeader, priorities, false)
    if err != nil {
        panic(err)
    }
    
    if best != nil {
        fmt.Printf("Best match: %s\n", best.Type)
        // Output: Best match: text/html
    }
}
```

### Language Negotiation

```go
negotiator := negotiation.NewLanguageNegotiator()

acceptLanguageHeader := "en; q=0.1, fr; q=0.4, fu; q=0.9, de; q=0.2"
priorities := []string{"en", "fu", "de"}

best, err := negotiator.GetBest(acceptLanguageHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best language: %s\n", best.Type)
    // Output: Best language: fu
    fmt.Printf("Quality: %f\n", best.Quality)
    // Output: Quality: 0.900000
}
```

### Charset Negotiation

```go
negotiator := negotiation.NewCharsetNegotiator()

acceptCharsetHeader := "ISO-8859-1, UTF-8; q=0.9"
priorities := []string{"iso-8859-1;q=0.3", "utf-8;q=0.9", "utf-16;q=1.0"}

best, err := negotiator.GetBest(acceptCharsetHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best charset: %s\n", best.Type)
    // Output: Best charset: utf-8
}
```

### Encoding Negotiation

```go
negotiator := negotiation.NewEncodingNegotiator()

acceptEncodingHeader := "gzip;q=1.0, identity; q=0.5, *;q=0"
priorities := []string{"identity", "gzip"}

best, err := negotiator.GetBest(acceptEncodingHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best encoding: %s\n", best.Type)
    // Output: Best encoding: identity
}
```

### Getting Ordered Elements

You can also get all accept header elements ordered by quality:

```go
negotiator := negotiation.NewMediaNegotiator()

elements, err := negotiator.GetOrderedElements("text/html;q=0.3, text/html;q=0.7")
if err != nil {
    panic(err)
}

for _, elem := range elements {
    fmt.Printf("%s (q=%f)\n", elem.Value, elem.Quality)
}
// Output:
// text/html;q=0.7 (q=0.700000)
// text/html;q=0.3 (q=0.300000)
```

## Error Handling

The package defines several error types:

- `ErrInvalidArgument` - Invalid argument provided
- `ErrInvalidHeader` - Header cannot be parsed
- `ErrInvalidMediaType` - Invalid media type format
- `ErrInvalidLanguage` - Invalid language tag format

## Limitations and Best Practices

### Quality Value Handling

⚠️ **Important:** Quality values (q-values) are clamped to the range [0.0, 1.0]:

```go
// These are equivalent:
"application/json;q=1.5"  // Treated as q=1.0
"application/json;q=-0.5" // Treated as q=0.0
```

### Header Parsing

- Headers are parsed case-insensitively for media types and charsets
- Language tags are normalized to lowercase
- Parameters are sorted alphabetically for consistent matching
- Malformed headers return `ErrInvalidHeader`

## API Stability

This library follows semantic versioning. The public API is stable for v1.x:

**Stable APIs:**
- `NewMediaNegotiator()`, `NewLanguageNegotiator()`, `NewCharsetNegotiator()`, `NewEncodingNegotiator()`
- `Negotiator.Negotiate(header, priorities, strict)`, `Negotiator.GetOrderedElements(header)`
- `Header` struct and all exported fields
- All exported error types: `ErrInvalidArgument`, `ErrInvalidHeader`, `ErrInvalidMediaType`, `ErrInvalidLanguage`

## Development Commands

```bash
# Run tests
go test -v ./...

# Run with race detector
go test -race ./...

# Run linter
golangci-lint run

# Run tests with coverage and generate report
go test -v -coverpkg=./... -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks (when implemented)
go test -bench=. -benchmem

# Generate coverage report for CI
go test -coverprofile=coverage.out -covermode=atomic ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Developed by [Talav](https://github.com/talav).

**Questions?** Open an issue or discussion on GitHub.

**Found a bug?** Please report it with a minimal reproduction case.