// Package helmstrip removes common Helm-added labels and annotations from
// Kubernetes resource yaml.Node trees.
package helmstrip

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// helmLabels are label keys unconditionally removed from metadata.labels.
var helmLabels = map[string]bool{
	"helm.sh/chart":                  true,
	"app.kubernetes.io/managed-by":   true,
	"app.kubernetes.io/version":      true,
}

// helmAnnotations are annotation keys unconditionally removed from metadata.annotations.
var helmAnnotations = map[string]bool{
	"helm.sh/resource-policy":          true,
	"helm.sh/hook":                     true,
	"helm.sh/hook-weight":              true,
	"helm.sh/hook-delete-policy":       true,
	"meta.helm.sh/release-name":        true,
	"meta.helm.sh/release-namespace":   true,
}

// Strip removes Helm-specific labels, annotations, and other artifacts from a
// document node in place.
func Strip(doc *yaml.Node) {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return
	}

	// Top-level metadata.
	stripMetadata(root)

	// Pod template metadata: spec.template.metadata
	// Applies to Deployment, StatefulSet, DaemonSet, Job, ReplicaSet.
	if spec := mappingValue(root, "spec"); spec != nil {
		if tmpl := mappingValue(spec, "template"); tmpl != nil {
			stripMetadata(tmpl)
		}
		// CronJob: spec.jobTemplate.spec.template.metadata
		if jobTmpl := mappingValue(spec, "jobTemplate"); jobTmpl != nil {
			if jobSpec := mappingValue(jobTmpl, "spec"); jobSpec != nil {
				if tmpl := mappingValue(jobSpec, "template"); tmpl != nil {
					stripMetadata(tmpl)
				}
			}
		}
	}

	stripSourceComments(doc, root)
	UnquoteScalars(doc)
}

// stripMetadata removes Helm labels, annotations, and creationTimestamp from
// a node that contains a "metadata" key (e.g. a resource root or a pod template).
func stripMetadata(parent *yaml.Node) {
	metaNode := mappingValue(parent, "metadata")
	if metaNode == nil || metaNode.Kind != yaml.MappingNode {
		return
	}
	stripMappingKeys(metaNode, "labels", helmLabels)
	stripMappingKeys(metaNode, "annotations", helmAnnotations)
	removeMappingKey(metaNode, "creationTimestamp")
}

// stripSourceComments removes "# Source: ..." head comments that Helm adds
// to each template output document. yaml.v3 may place these on the document
// node, the root mapping node, or the first key node.
func stripSourceComments(doc, root *yaml.Node) {
	doc.HeadComment = filterSourceComment(doc.HeadComment)
	root.HeadComment = filterSourceComment(root.HeadComment)
	if len(root.Content) > 0 {
		root.Content[0].HeadComment = filterSourceComment(root.Content[0].HeadComment)
	}
}

// filterSourceComment removes lines containing "Source:" from a comment string.
// yaml.v3 stores comments with the "# " prefix intact.
func filterSourceComment(comment string) string {
	if comment == "" {
		return ""
	}
	var kept []string
	for _, line := range strings.Split(comment, "\n") {
		trimmed := strings.TrimSpace(line)
		// Match "# Source: ...", "Source: ...", with optional leading "#"
		stripped := strings.TrimLeft(trimmed, "# ")
		if strings.HasPrefix(stripped, "Source:") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// stripMappingKeys removes specific keys from a named sub-map within parent.
// If the sub-map becomes empty after removal, it is removed from parent entirely.
func stripMappingKeys(parent *yaml.Node, subKey string, keysToRemove map[string]bool) {
	subNode := mappingValue(parent, subKey)
	if subNode == nil || subNode.Kind != yaml.MappingNode {
		return
	}

	newContent := make([]*yaml.Node, 0, len(subNode.Content))
	for i := 0; i+1 < len(subNode.Content); i += 2 {
		k := subNode.Content[i]
		v := subNode.Content[i+1]
		if !keysToRemove[k.Value] {
			newContent = append(newContent, k, v)
		}
	}

	if len(newContent) == 0 {
		// Remove the subKey from parent entirely.
		removeMappingKey(parent, subKey)
	} else {
		subNode.Content = newContent
	}
}

// mappingValue returns the value node for key in a MappingNode, or nil.
func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// removeMappingKey removes a key/value pair from a MappingNode by key name.
func removeMappingKey(m *yaml.Node, key string) {
	newContent := make([]*yaml.Node, 0, len(m.Content))
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value != key {
			newContent = append(newContent, m.Content[i], m.Content[i+1])
		}
	}
	m.Content = newContent
}
