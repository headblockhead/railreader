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
          src = pkgs.lib.cleanSourceWith {
            src = ./.;
            filter = path: type: !(pkgs.lib.hasSuffix ".nix" path);
          };
          vendorHash = "sha256-4hEiIL6Pkzc5hwsWjOZjv/8DHdo+JdZ43R9YJ+dK/h4=";
        };
        default = railreader;
        railreader-docker = pkgs.dockerTools.buildLayeredImage {
          name = "ghcr.io/headblockhead/railreader";
          compressor = "zstd";
          config.Entrypoint = [ "${railreader}/bin/railreader" ];
        };
      });
      nixosModules.railreader = { config, lib, pkgs, ... }: import ./service.nix {
        inherit config pkgs lib;
        railreader = packages.${pkgs.system}.railreader;
      };
    };
}
