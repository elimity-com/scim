{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          ci = pkgs.writeShellScriptBin "ci" ''
            set -euo pipefail
            echo "--- test ---"
            go test -v ./...
            echo "--- lint ---"
            golangci-lint run -E misspell,godot,whitespace ./...
            echo "--- arrange ---"
            command -v goarrange >/dev/null || go install github.com/jdeflander/goarrange@v1.0.0
            test -z "$(goarrange run -r -d)"
            echo "--- tidy ---"
            go mod tidy
            git diff --quiet go.mod go.sum
          '';
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go_1_26
              pkgs.golangci-lint
              ci
            ];
          };
        }
      );
    };
}
