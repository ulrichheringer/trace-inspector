package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Export struct {
	ResourceSpans []ResourceSpans `json:"resourceSpans"`
}

type ResourceSpans struct {
	Resource   Resource     `json:"resource"`
	ScopeSpans []ScopeSpans `json:"scopeSpans"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type ScopeSpans struct {
	Spans []Span `json:"spans"`
}

type Attribute struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type Span struct {
	TraceID      string `json:"traceId"`
	SpanID       string `json:"spanId"`
	ParentSpanID string `json:"parentSpanId"`
	Name         string `json:"name"`
}

type Finding struct {
	TraceID string
	SpanID  string
	Service string
	Kind    string
	Detail  string
}

type indexedSpan struct {
	span    Span
	service string
}

func Inspect(export Export) []Finding {
	byTrace := make(map[string][]indexedSpan)
	for _, resource := range export.ResourceSpans {
		service := attributeString(resource.Resource.Attributes, "service.name")
		for _, scope := range resource.ScopeSpans {
			for _, span := range scope.Spans {
				if span.TraceID == "" || span.SpanID == "" {
					continue
				}
				byTrace[strings.ToLower(span.TraceID)] = append(byTrace[strings.ToLower(span.TraceID)], indexedSpan{span, service})
			}
		}
	}

	var findings []Finding
	for traceID, spans := range byTrace {
		present := make(map[string]bool, len(spans))
		rootsByService := make(map[string][]string)
		for _, item := range spans {
			present[strings.ToLower(item.span.SpanID)] = true
			if item.span.ParentSpanID == "" {
				rootsByService[item.service] = append(rootsByService[item.service], item.span.SpanID)
			}
		}
		for _, item := range spans {
			parent := strings.ToLower(item.span.ParentSpanID)
			if parent != "" && !present[parent] {
				findings = append(findings, Finding{
					TraceID: traceID, SpanID: item.span.SpanID, Service: item.service,
					Kind: "missing-parent", Detail: fmt.Sprintf("span %q references parent %q absent from this trace export", item.span.Name, item.span.ParentSpanID),
				})
			}
		}
		services := make([]string, 0, len(rootsByService))
		rootCount := 0
		for service, ids := range rootsByService {
			if service != "" {
				services = append(services, service)
				rootCount += len(ids)
			}
		}
		if len(services) > 1 {
			sort.Strings(services)
			findings = append(findings, Finding{
				TraceID: traceID, Kind: "multiple-service-roots",
				Detail: fmt.Sprintf("%d root spans start in services %s; context propagation may be missing", rootCount, strings.Join(services, ", ")),
			})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].TraceID != findings[j].TraceID {
			return findings[i].TraceID < findings[j].TraceID
		}
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		return findings[i].SpanID < findings[j].SpanID
	})
	return findings
}

func attributeString(attributes []Attribute, key string) string {
	for _, attribute := range attributes {
		if attribute.Key != key {
			continue
		}
		var value struct {
			StringValue string `json:"stringValue"`
		}
		if json.Unmarshal(attribute.Value, &value) == nil {
			return value.StringValue
		}
	}
	return ""
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: trace-inspector <otlp-traces.json>")
	}
	file, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()
	var export Export
	if err := json.NewDecoder(file).Decode(&export); err != nil {
		return fmt.Errorf("decode OTLP JSON: %w", err)
	}
	findings := Inspect(export)
	if len(findings) == 0 {
		_, err := fmt.Fprintln(stdout, "No possible context propagation gaps found.")
		return err
	}
	for _, finding := range findings {
		_, err := fmt.Fprintf(stdout, "%s trace=%s span=%s service=%s: %s\n", finding.Kind, finding.TraceID, finding.SpanID, finding.Service, finding.Detail)
		if err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintf(stderr, "%d possible issue(s) found\n", len(findings))
	return nil
}
