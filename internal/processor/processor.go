// Package processor orchestrates splitting a multi-document Helm template
// output YAML file into individual kustomize-ready resource files.
package processor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/snarlysodboxer/helm-to-kustomize/internal/helmstrip"
)

// Run reads inputFile, splits it into individual resource files under outputDir,
// removes Helm labels/annotations, and writes a kustomization.yaml.
func Run(inputFile, outputDir string) error {
	f, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	dec := yaml.NewDecoder(f)

	// nameCounts tracks how many times we've seen each kind.name combo to
	// handle collisions by appending a numeric suffix.
	nameCounts := map[string]int{}
	var resources []string

	for {
		var doc yaml.Node
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode YAML: %w", err)
		}

		// Skip empty documents.
		if doc.Kind == 0 || len(doc.Content) == 0 {
			continue
		}

		kind, name, err := extractKindName(&doc)
		if err != nil {
			// Skip documents that don't look like Kubernetes resources
			// (e.g. trailing --- separators or non-resource documents).
			continue
		}

		helmstrip.Strip(&doc)

		data, err := marshalDoc(&doc)
		if err != nil {
			return fmt.Errorf("marshal YAML for %s.%s: %w", kind, name, err)
		}

		filename := buildFilename(kind, name, nameCounts)
		resources = append(resources, filename)

		outPath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		fmt.Printf("wrote %s\n", outPath)
	}

	if len(resources) == 0 {
		return fmt.Errorf("no resources found in %s", inputFile)
	}

	sort.Strings(resources)

	if err := writeKustomization(outputDir, resources); err != nil {
		return fmt.Errorf("write kustomization.yaml: %w", err)
	}

	fmt.Printf("wrote %s\n", filepath.Join(outputDir, "kustomization.yaml"))
	return nil
}

// extractKindName returns the camelCase kind and lowercase metadata.name from a document node.
func extractKindName(doc *yaml.Node) (kind, name string, err error) {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return "", "", fmt.Errorf("not a document node")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return "", "", fmt.Errorf("root is not a mapping node")
	}

	kind = lowerFirst(mappingValue(root, "kind"))
	if kind == "" {
		return "", "", fmt.Errorf("missing 'kind' field")
	}

	metaNode := mappingNode(root, "metadata")
	if metaNode == nil {
		return "", "", fmt.Errorf("missing 'metadata' field")
	}
	name = strings.ReplaceAll(strings.ToLower(mappingValue(metaNode, "name")), ":", "_")
	if name == "" {
		return "", "", fmt.Errorf("missing 'metadata.name' field")
	}

	return kind, name, nil
}

// buildFilename generates a unique filename for kind.name, appending a counter
// suffix if the combination has been seen before.
func buildFilename(kind, name string, counts map[string]int) string {
	base := kind + "." + name
	counts[base]++
	if counts[base] == 1 {
		return base + ".yaml"
	}
	return fmt.Sprintf("%s.%d.yaml", base, counts[base])
}

// marshalDoc encodes a yaml.Node document to bytes.
func marshalDoc(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeKustomization writes a kustomization.yaml listing the given resources.
func writeKustomization(outputDir string, resources []string) error {
	resourceNodes := make([]*yaml.Node, 0, len(resources))
	for _, r := range resources {
		resourceNodes = append(resourceNodes, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: r,
			Tag:   "!!str",
		})
	}

	resourceSeq := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: resourceNodes,
	}

	doc := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					scalar("apiVersion"), scalar("kustomize.config.k8s.io/v1beta1"),
					scalar("kind"), scalar("Kustomization"),
					scalar("resources"), resourceSeq,
				},
			},
		},
	}

	data, err := marshalDoc(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, "kustomization.yaml"), data, 0o644)
}

// mappingValue returns the scalar value for key in a MappingNode, or "".
func mappingValue(m *yaml.Node, key string) string {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1].Value
		}
	}
	return ""
}

// mappingNode returns the value node for key in a MappingNode, or nil.
func mappingNode(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// lowerFirst returns s with only the first character lowercased (camelCase kind).
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// scalar creates a simple string scalar node.
func scalar(val string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: val,
		Tag:   "!!str",
	}
}
