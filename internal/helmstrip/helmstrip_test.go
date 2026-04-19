package helmstrip

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStripSourceComments(t *testing.T) {
	input := `# Source: gateway-helm/templates/certgen-rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
  labels:
    helm.sh/chart: test-1.0.0
    app.kubernetes.io/managed-by: Helm
`
	var doc yaml.Node
	dec := yaml.NewDecoder(strings.NewReader(input))
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode: %v", err)
	}

	Strip(&doc)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		t.Fatalf("encode: %v", err)
	}
	enc.Close()

	out := buf.String()
	t.Logf("output:\n%s", out)

	if strings.Contains(out, "Source:") {
		t.Errorf("output still contains Source comment:\n%s", out)
	}
	if strings.Contains(out, "helm.sh/chart") {
		t.Errorf("output still contains helm.sh/chart label:\n%s", out)
	}
	if strings.Contains(out, "managed-by") {
		t.Errorf("output still contains managed-by label:\n%s", out)
	}
}
