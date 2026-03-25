{
  description = "Treni - Train tracking web application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "trenid";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-n29vH5TFOBJ4wUEjwLeXYOEW5zyhVlVEuui9uRaV3TY=";

            nativeBuildInputs = [ pkgs.templ ];

            preBuild = ''
              templ generate
            '';

            subPackages = [ "cmd/trenid" ];

            meta = {
              description = "Train tracking web application";
              mainProgram = "trenid";
            };
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.templ
              pkgs.golangci-lint
            ];
          };
        }
      );
    };
}
