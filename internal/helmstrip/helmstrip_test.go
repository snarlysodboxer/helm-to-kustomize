package helmstrip

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func decode(t *testing.T, input string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.NewDecoder(strings.NewReader(input)).Decode(&doc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return &doc
}

func encode(t *testing.T, doc *yaml.Node) string {
	t.Helper()
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		t.Fatalf("encode: %v", err)
	}
	enc.Close()
	return buf.String()
}

func TestStripHelmLabels(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  labels:
    app: myapp
    helm.sh/chart: myapp-1.0.0
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/name: myapp
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	for _, label := range []string{"helm.sh/chart", "app.kubernetes.io/managed-by", "app.kubernetes.io/version"} {
		if strings.Contains(out, label) {
			t.Errorf("label %q should have been removed:\n%s", label, out)
		}
	}
	for _, label := range []string{"app: myapp", "app.kubernetes.io/name: myapp"} {
		if !strings.Contains(out, label) {
			t.Errorf("label %q should have been kept:\n%s", label, out)
		}
	}
}

func TestStripHelmLabels_AllRemoved(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  labels:
    helm.sh/chart: myapp-1.0.0
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/version: "1.0.0"
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "labels") {
		t.Errorf("labels key should have been removed entirely when empty:\n%s", out)
	}
}

func TestStripHelmAnnotations(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  annotations:
    helm.sh/resource-policy: keep
    meta.helm.sh/release-name: my-release
    meta.helm.sh/release-namespace: default
    custom-annotation: keep-me
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	for _, ann := range []string{"helm.sh/resource-policy", "meta.helm.sh/release-name", "meta.helm.sh/release-namespace"} {
		if strings.Contains(out, ann) {
			t.Errorf("annotation %q should have been removed:\n%s", ann, out)
		}
	}
	if !strings.Contains(out, "custom-annotation: keep-me") {
		t.Errorf("custom-annotation should have been kept:\n%s", out)
	}
}

func TestStripHelmHookAnnotations(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: certgen
  annotations:
    "helm.sh/hook": pre-install, pre-upgrade
    "helm.sh/hook-weight": "-1"
    "helm.sh/hook-delete-policy": before-hook-creation
    custom-annotation: keep-me
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	for _, ann := range []string{"helm.sh/hook", "helm.sh/hook-weight", "helm.sh/hook-delete-policy"} {
		if strings.Contains(out, ann) {
			t.Errorf("annotation %q should have been removed:\n%s", ann, out)
		}
	}
	if !strings.Contains(out, "custom-annotation: keep-me") {
		t.Errorf("custom-annotation should have been kept:\n%s", out)
	}
}

func TestStripHelmAnnotations_AllRemoved(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "-1"
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "annotations") {
		t.Errorf("annotations key should have been removed entirely when empty:\n%s", out)
	}
}

func TestStripSourceComment(t *testing.T) {
	input := `# Source: gateway-helm/templates/certgen-rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "Source:") {
		t.Errorf("Source comment should have been removed:\n%s", out)
	}
}

func TestStripNoMetadata(t *testing.T) {
	input := `apiVersion: v1
kind: Namespace
`
	doc := decode(t, input)
	// Should not panic.
	Strip(doc)
}
