{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f {
        pkgs = import nixpkgs {
          inherit system;
        };
      });
    in
    {
      packages = forAllSystems (
        { pkgs }:
        rec {
          railreader = pkgs.buildGoModule {
            pname = "railreader";
            version = "0.1";
            src = ./.;
            vendorHash = "";
          };
          default = railreader;
          railreader-docker = pkgs.dockerTools.buildLayeredImage {
            name = "registry.edwardh.dev/railreader"; # TODO
            tag = "latest";
            config.Cmd = "${railreader}/bin/railreader";
          };
        }
      );
    };
}
