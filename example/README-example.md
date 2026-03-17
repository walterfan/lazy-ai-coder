# AwesomeCLI

A cmd‑line tool to turbo‑charge batch image processing

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See **Deployment** for notes on deploying to a live system.

### Prerequisites

- Go 1.21+
- ImageMagick

### Installing

```bash
go install github.com/me/awesomecli@latest
```

### Usage Examples

```bash
awesomecli -i imgs/ -o out/ -resize 800x600
```

## Running the Tests

`go test ./...`

## Deployment

N/A

## Built With

- Go
- cobra

## Contributing

See CONTRIBUTING.md

## Versioning

Semantic Versioning via Git tags

## Authors

- You

## License

MIT License

## Acknowledgments

- Inspired by PurpleBooth’s README template
