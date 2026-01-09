{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.buildGoModule {
            pname = "treni";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-f4BWbgHr7etAyhQyflflb+1eHWrlvabyf/ctrChlzxM=";

            subPackages = [ "cmd" ];

            postInstall = ''
              mv $out/bin/cmd $out/bin/treni
            '';

            meta = {
              description = "Train tracking application";
              mainProgram = "treni";
            };
          };
        });

      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            packages = [ pkgs.go_1_25 ];
          };
        });
    };
}
