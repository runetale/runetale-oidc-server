{
  description = "runetale-oidc-server";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-24.05";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      # Generate a user-friendly version number.
      version = builtins.substring 0 8 self.lastModifiedDate;
      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
    in
    flake-utils.lib.eachSystem supportedSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          containerImage = pkgs.dockerTools.buildLayeredImage
            {
              name = "runetale-oidc-server";
              contents = [
                pkgs.bashInteractive
                pkgs.cacert
                pkgs.coreutils-full
                self.packages.${system}.runetale-oidc-server
              ];
              extraCommands = ''
                install -D -t ./migrations ${./migrations}/*
              '';
              config = {
                Entrypoint = [ "${pkgs.tini}/bin/tini" "--" ];
                Cmd = [ "${self.packages.${system}.runetale-oidc-server}/bin/runetale-oidc-server" ];
              };
            };
          runetale-oidc-server =
            pkgs.buildGoModule {
              pname = "runetale-oidc-server";
              inherit version;
              src = self;
              vendorHash = "sha256-G6VlZfSK4OAnG3SEVS7SDe89wTmq82+Xv5gy8Cg78vc=";
              doCheck = false;
            };
          db-migration =
            pkgs.buildGoModule {
              pname = "db-migration";
              inherit version;
              src = self;
              subPackages = [ "cmd/db" ];
              vendorHash = "sha256-G6VlZfSK4OAnG3SEVS7SDe89wTmq82+Xv5gy8Cg78vc=";
              doCheck = false;
            };
        };
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_21
            gopls
            go-migrate
            postgresql
            tailwindcss
          ];
        };
        defaultPackage = self.packages.${system}.runetale-oidc-server;
      });
}
