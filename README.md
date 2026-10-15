# trace-inspector

Inspect an OpenTelemetry OTLP JSON trace export for spans whose parent is missing and traces with roots in multiple services.

## Install

Requires Go 1.23 or newer. Build with `go install github.com/ulrichheringer/trace-inspector@latest`.

## Use

Run `trace-inspector traces.json` with an OTLP JSON export.

## Example

Run `go run . examples/simulated-traces.json` for a simulated export with a possible propagation gap.
