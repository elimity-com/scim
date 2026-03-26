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
          goarrange = pkgs.buildGoModule {
            pname = "goarrange";
            version = "1.0.0";
            src = pkgs.fetchFromGitHub {
              owner = "jdeflander";
              repo = "goarrange";
              rev = "v1.0.0";
              hash = "sha256-V03BgTeWcAspMHGUHlAgSbiTaoZ42hgb/Zb/yqZ2m+k=";
            };
            vendorHash = "sha256-Xhxfiw1WeXFHrYIYvUytEtMzMbSxOrignmUC5kVna0o=";
          };
          lint = pkgs.writeShellScriptBin "lint" ''
            set -euo pipefail
            echo "--- format ---"
            go fmt ./...
            goarrange run -r
            echo "--- lint ---"
            golangci-lint run -E misspell,godot,whitespace ./...
            echo "--- tidy ---"
            go mod tidy
          '';
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go_1_26
              pkgs.golangci-lint
              goarrange
              lint
            ];
          };
        }
      );
    };
}
