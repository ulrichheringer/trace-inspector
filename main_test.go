package main

import "testing"

func TestInspectFindsMissingParentAndMultipleRoots(t *testing.T) {
	export := Export{ResourceSpans: []ResourceSpans{
		{Resource: Resource{Attributes: []Attribute{{Key: "service.name", Value: []byte(`{"stringValue":"checkout"}`)}}}, ScopeSpans: []ScopeSpans{{Spans: []Span{{TraceID: "abc", SpanID: "1", Name: "request"}, {TraceID: "abc", SpanID: "2", ParentSpanID: "missing", Name: "call"}}}}},
		{Resource: Resource{Attributes: []Attribute{{Key: "service.name", Value: []byte(`{"stringValue":"payments"}`)}}}, ScopeSpans: []ScopeSpans{{Spans: []Span{{TraceID: "abc", SpanID: "3", Name: "request"}}}}},
	}}
	findings := Inspect(export)
	if len(findings) != 2 {
		t.Fatalf("got %d findings: %#v", len(findings), findings)
	}
	if findings[0].Kind != "missing-parent" && findings[1].Kind != "missing-parent" {
		t.Fatalf("missing parent finding: %#v", findings)
	}
	if findings[0].Kind != "multiple-service-roots" && findings[1].Kind != "multiple-service-roots" {
		t.Fatalf("multiple roots finding: %#v", findings)
	}
}

func TestRenderBuildsServiceSpanTree(t *testing.T) {
	export := Export{ResourceSpans: []ResourceSpans{
		{Resource: Resource{Attributes: []Attribute{{Key: "service.name", Value: []byte(`{"stringValue":"checkout"}`)}}}, ScopeSpans: []ScopeSpans{{Spans: []Span{{TraceID: "abc", SpanID: "01", Name: "request"}}}}},
		{Resource: Resource{Attributes: []Attribute{{Key: "service.name", Value: []byte(`{"stringValue":"payments"}`)}}}, ScopeSpans: []ScopeSpans{{Spans: []Span{{TraceID: "abc", SpanID: "02", ParentSpanID: "01", Name: "charge"}}}}},
	}}
	lines := Render(export)
	if len(lines) != 3 || lines[0] != "Trace abc" || lines[2] != "  [payments] charge (span 02)" {
		t.Fatalf("unexpected trace tree: %#v", lines)
	}
}

func TestInspectIgnoresSingleServiceRoot(t *testing.T) {
	export := Export{ResourceSpans: []ResourceSpans{{Resource: Resource{Attributes: []Attribute{{Key: "service.name", Value: []byte(`{"stringValue":"api"}`)}}}, ScopeSpans: []ScopeSpans{{Spans: []Span{{TraceID: "abc", SpanID: "1"}, {TraceID: "abc", SpanID: "2", ParentSpanID: "1"}}}}}}}
	if findings := Inspect(export); len(findings) != 0 {
		t.Fatalf("unexpected findings: %#v", findings)
	}
}
