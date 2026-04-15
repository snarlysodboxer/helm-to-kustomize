# helm-to-kustomize

An opinionated tool that converts helm template output into kustomize-ready YAML files.

## What it does

- Splits a multi-document Helm template output into one file per Kubernetes resource
- Names each file `<kind>.<metadata.name>.yaml` (lowercase)
- Removes common Helm-added labels and annotations
- Writes a `kustomization.yaml` listing all generated resources

### Removed labels

- `helm.sh/chart`
- `app.kubernetes.io/managed-by`
- `app.kubernetes.io/version`

### Removed annotations

- `helm.sh/resource-policy`
- `meta.helm.sh/release-name`
- `meta.helm.sh/release-namespace`

If removing all Helm labels or annotations leaves an empty `labels` or `annotations` map, the key is dropped entirely.

## Usage

```sh
helm template <release> <chart> [flags] > chart.yaml
helm-to-kustomize --input-file chart.yaml --output-dir ./k8s/base
```

Both flags are required.

## YAML formatting

The output YAML is valid but unformatted beyond `yaml.v3`'s defaults (2-space indent). Run [`yamlfmt`](https://github.com/google/yamlfmt) over the output directory to apply consistent formatting:

```sh
yamlfmt ./k8s/base/*.yaml
```

A `.yamlfmt` config is included in this repo with the project's preferred style (`indentless_arrays`, `include_document_start`, etc.).

## Installation

### Nix flake

```sh
nix run github:snarlysodboxer/helm-to-kustomize -- --input-file chart.yaml --output-dir ./k8s/base
```

Or install into your profile:

```sh
nix profile install github:snarlysodboxer/helm-to-kustomize
```

### Go install

```sh
go install github.com/snarlysodboxer/helm-to-kustomize@latest
```

## Development

With Nix:

```sh
nix develop
go run . --input-file istio-ambient.yaml --output-dir /tmp/out
yamlfmt /tmp/out/*.yaml
```

Without Nix, you need Go 1.25+:

```sh
go build -o helm-to-kustomize .
```

## Example

```sh
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm template istio-base istio/base --namespace istio-system > istio-base.yaml
helm-to-kustomize --input-file istio-base.yaml --output-dir ./k8s/istio-base
yamlfmt ./k8s/istio-base/*.yaml
```
