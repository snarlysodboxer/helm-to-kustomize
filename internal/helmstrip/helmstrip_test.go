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

func TestStripCreationTimestamp_Null(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  creationTimestamp: null
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "creationTimestamp") {
		t.Errorf("creationTimestamp: null should have been removed:\n%s", out)
	}
}

func TestStripCreationTimestamp_WithValue(t *testing.T) {
	input := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  creationTimestamp: "2024-01-01T00:00:00Z"
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "creationTimestamp") {
		t.Errorf("creationTimestamp should have been removed regardless of value:\n%s", out)
	}
}

func TestStripDeploymentPodTemplateLabels(t *testing.T) {
	input := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
    helm.sh/chart: myapp-1.0.0
spec:
  template:
    metadata:
      labels:
        app: myapp
        helm.sh/chart: myapp-1.0.0
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/version: "1.0.0"
      annotations:
        helm.sh/hook: pre-install
    spec:
      containers:
      - name: myapp
        image: myapp:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	// Check pod template labels are stripped
	// Split at "spec:" to isolate the template section
	parts := strings.SplitN(out, "spec:", 2)
	if len(parts) < 2 {
		t.Fatalf("expected spec: in output:\n%s", out)
	}
	templateSection := parts[1]

	for _, label := range []string{"helm.sh/chart", "app.kubernetes.io/managed-by", "app.kubernetes.io/version"} {
		if strings.Contains(templateSection, label) {
			t.Errorf("pod template label %q should have been removed:\n%s", label, out)
		}
	}
	if !strings.Contains(templateSection, "app: myapp") {
		t.Errorf("pod template label 'app: myapp' should have been kept:\n%s", out)
	}
	if strings.Contains(templateSection, "helm.sh/hook") {
		t.Errorf("pod template annotation 'helm.sh/hook' should have been removed:\n%s", out)
	}
}

func TestStripStatefulSetPodTemplateLabels(t *testing.T) {
	input := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mydb
  labels:
    app: mydb
    helm.sh/chart: mydb-1.0.0
spec:
  template:
    metadata:
      labels:
        app: mydb
        helm.sh/chart: mydb-1.0.0
        app.kubernetes.io/managed-by: Helm
    spec:
      containers:
      - name: mydb
        image: mydb:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	parts := strings.SplitN(out, "spec:", 2)
	if len(parts) < 2 {
		t.Fatalf("expected spec: in output:\n%s", out)
	}
	templateSection := parts[1]

	for _, label := range []string{"helm.sh/chart", "app.kubernetes.io/managed-by"} {
		if strings.Contains(templateSection, label) {
			t.Errorf("pod template label %q should have been removed:\n%s", label, out)
		}
	}
	if !strings.Contains(templateSection, "app: mydb") {
		t.Errorf("pod template label 'app: mydb' should have been kept:\n%s", out)
	}
}

func TestStripDaemonSetPodTemplateLabels(t *testing.T) {
	input := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: agent
spec:
  template:
    metadata:
      labels:
        app: agent
        helm.sh/chart: agent-1.0.0
    spec:
      containers:
      - name: agent
        image: agent:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	parts := strings.SplitN(out, "spec:", 2)
	if len(parts) < 2 {
		t.Fatalf("expected spec: in output:\n%s", out)
	}
	if strings.Contains(parts[1], "helm.sh/chart") {
		t.Errorf("pod template label 'helm.sh/chart' should have been removed:\n%s", out)
	}
}

func TestStripJobPodTemplateLabels(t *testing.T) {
	input := `apiVersion: batch/v1
kind: Job
metadata:
  name: migrate
spec:
  template:
    metadata:
      labels:
        app: migrate
        helm.sh/chart: migrate-1.0.0
    spec:
      containers:
      - name: migrate
        image: migrate:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	parts := strings.SplitN(out, "spec:", 2)
	if len(parts) < 2 {
		t.Fatalf("expected spec: in output:\n%s", out)
	}
	if strings.Contains(parts[1], "helm.sh/chart") {
		t.Errorf("pod template label 'helm.sh/chart' should have been removed:\n%s", out)
	}
}

func TestStripCronJobPodTemplateLabels(t *testing.T) {
	input := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup
spec:
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: backup
            helm.sh/chart: backup-1.0.0
            app.kubernetes.io/managed-by: Helm
        spec:
          containers:
          - name: backup
            image: backup:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	if strings.Contains(out, "helm.sh/chart") {
		t.Errorf("CronJob pod template label 'helm.sh/chart' should have been removed:\n%s", out)
	}
	if strings.Contains(out, "app.kubernetes.io/managed-by") {
		t.Errorf("CronJob pod template label 'managed-by' should have been removed:\n%s", out)
	}
	if !strings.Contains(out, "app: backup") {
		t.Errorf("CronJob pod template label 'app: backup' should have been kept:\n%s", out)
	}
}

func TestStripPodTemplateCreationTimestamp(t *testing.T) {
	input := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest
`
	doc := decode(t, input)
	Strip(doc)
	out := encode(t, doc)

	parts := strings.SplitN(out, "spec:", 2)
	if len(parts) < 2 {
		t.Fatalf("expected spec: in output:\n%s", out)
	}
	if strings.Contains(parts[1], "creationTimestamp") {
		t.Errorf("pod template creationTimestamp should have been removed:\n%s", out)
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
