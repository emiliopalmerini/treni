{
  description = "Treni - Telegram bot for Italian train tracking";

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
            pname = "trenibot";
            version = "0.2.0";
            src = ./.;
            # First `nix build` will fail and print the real sha256 — paste it here.
            vendorHash = nixpkgs.lib.fakeHash;

            subPackages = [ "cmd/trenibot" ];

            meta = {
              description = "Telegram bot for Italian train tracking";
              mainProgram = "trenibot";
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
              pkgs.golangci-lint
            ];
          };
        }
      );
    };
}
