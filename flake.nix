{
  description = "helm-to-kustomize: An opinionated tool that converts helm template output into kustomize-ready YAML files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "helm-to-kustomize";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-komX1AmHt2NoF1x6xsNa2RFkfVzOXfYEMPhT0zwMxjw=";
          env.CGO_ENABLED = "0";

          nativeBuildInputs = [ pkgs.installShellFiles ];

          postInstall = ''
            installShellCompletion --cmd helm-to-kustomize \
              --bash <($out/bin/helm-to-kustomize completion bash) \
              --zsh <($out/bin/helm-to-kustomize completion zsh) \
              --fish <($out/bin/helm-to-kustomize completion fish)
          '';
          meta = with pkgs.lib; {
            description = "An opinionated tool that converts helm template output into kustomize-ready YAML files";
            homepage = "https://github.com/snarlysodboxer/helm-to-kustomize";
            license = licenses.mit;
            mainProgram = "helm-to-kustomize";
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            yamlfmt
            helm
          ];

          shellHook = ''
            echo "helm-to-kustomize dev shell"
            echo "  go run . --input-file <file> --output-dir <dir>"

            # Shell completion for the built binary (if it exists)
            if command -v helm-to-kustomize &>/dev/null; then
              source <(helm-to-kustomize completion bash)
            fi
          '';
        };
      }
    );
}
