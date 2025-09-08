{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forEachSystem = f: nixpkgs.lib.genAttrs systems (system: f (import nixpkgs { inherit system; }));
    in
    rec {
      packages = forEachSystem (pkgs: rec {
        railreader = pkgs.buildGo125Module {
          pname = "railreader";
          version = "0.1";
          src = ./.;
          vendorHash = "sha256-imuE8g3wPjUN79tFy1x3wYmJoSSGJp0DRMAP+HEu4HI=";
        };
        default = railreader;
        railreader-docker = pkgs.dockerTools.buildLayeredImage {
          name = "ghcr.io/headblockhead/railreader";
          compressor = "zstd";
          config.Entrypoint = [ "${railreader}/bin/railreader" ];
        };
      });
      nixosModules.railreader = { config, options, lib, pkgs, ... }: import ./service.nix {
        inherit config options lib;
        railreader = packages.${pkgs.system}.railreader;
      };
    };
}
