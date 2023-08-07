{
  description = "zombiezen.com/go/uritemplate Go package";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        packages.go = pkgs.go_1_19;

        devShells.default = pkgs.mkShell {
          packages = [
            self.packages.${system}.go
          ];
        };
      }
    );
}
