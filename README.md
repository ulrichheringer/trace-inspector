# trace-inspector

Explore OpenTelemetry traces and spot possible context propagation gaps. The CLI prints spans as a service tree and flags missing parents or roots in multiple services.

## Install

With Go 1.23 or newer, install it with `go install github.com/ulrichheringer/trace-inspector@latest`.

## Use

Run `trace-inspector traces.json` with an OTLP JSON trace export.

## Example

The example data is simulated. Run `go run . examples/simulated-traces.json` to see a possible propagation gap.
