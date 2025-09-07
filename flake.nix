{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
          railreader = pkgs.buildGo125Module {
            pname = "railreader";
            version = "0.1";
            src = ./.;
            vendorHash = "sha256-imuE8g3wPjUN79tFy1x3wYmJoSSGJp0DRMAP+HEu4HI=";
          };
          default = railreader;
          railreader-docker = pkgs.dockerTools.buildLayeredImage {
            name = "ghcr.io/headblockhead/railreader";
            tag = "latest";
            config.Cmd = "${railreader}/bin/railreader";
          };
        }
      );
    };
}
