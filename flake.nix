{
  description = "civ6.ch";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        webDeps = pkgs.fetchPnpmDeps {
          pname = "civ6-web";
          version = "0.0.1";
          src = ./web;
          fetcherVersion = 3;
          hash = "sha256-VwBxXTfQzXhHdV1+37v9PEId8OdEj90xjMzfO/oHNBE=";
        };

        web = pkgs.stdenv.mkDerivation {

          pname = "civ6-web";
          version = "0.0.1";
          src = ./web;

          nativeBuildInputs = [
            pkgs.nodejs
            pkgs.pnpmConfigHook
            pkgs.python3
            pkgs.gcc
            pkgs.pkg-config
            pkgs.pnpm
          ];

          buildInputs = [
            pkgs.libsodium
          ];

          pnpmDeps = webDeps;

          buildPhase = ''pnpm run build'';
          installPhase = ''
            mkdir -p $out
            cp -r build/* $out/
            cp -r node_modules $out/node_modules
          '';
        };

        server = pkgs.buildGoModule {
          pname = "civ6";
          version = "0.0.1";
          src = ./.;
          vendorHash = null;
          nativeBuildInputs = [ pkgs.pkg-config ];
          buildInputs = [ pkgs.libwebp ];
        };

      in
      {
        packages.web = web;
        packages.server = server;
        packages.default = web;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            nodejs
            pnpm
            go
            postgresql
            python3
            gcc
            pkg-config
            libwebp
          ];
          shellHook = ''
            export PGDATA=$PWD/.postgres
            export PGHOST=$PWD/.postgres
            export PGDATABASE=civ6
            export PGUSER=$USER

            if [ ! -d $PGDATA ]; then
              initdb --no-locale --encoding=UTF8
              echo "unix_socket_directories = '$PWD/.postgres'" >> $PGDATA/postgresql.conf
              pg_ctl start -l $PGDATA/logfile
              createdb civ6
            else
              pg_ctl start -l $PGDATA/logfile 2>/dev/null || true
            fi
          '';
        };
      }
    ) // {
      nixosModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.services.civ6;
          system = pkgs.stdenv.hostPlatform.system;
          web = self.packages.${system}.web;
          server = self.packages.${system}.server;
        in {
          options.services.civ6.enable = lib.mkEnableOption "civ6.ch";

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ server ];

            services.postgresql = {
              enable = true;
              ensureDatabases = [ "civ6" ];
              ensureUsers = [{
                name = "civ6";
                ensureDBOwnership = true;
              }];
              authentication = pkgs.lib.mkOverride 10 ''
                local all all trust
                host all all 127.0.0.1/32 trust
                host all all ::1/128 trust
              '';
            };

            users.users.civ6 = {
              isSystemUser = true;
              group = "civ6";
            };
            users.groups.civ6 = {};

            systemd.services.civ6-server = {
              description = "civ6.ch Go API server";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" "postgresql.service" ];
              serviceConfig = {
                ExecStart = "${server}/bin/civ6";
                Restart = "on-failure";
                User = "civ6";
                Group = "civ6";
              };
            };

            systemd.services.civ6-web = {
              description = "civ6.ch SvelteKit server";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" "civ6-server.service" ];
              serviceConfig = {
                ExecStart = "${pkgs.nodejs}/bin/node ${web}/index.js";
                Restart = "on-failure";
                User = "civ6";
                Group = "civ6";
                Environment = [
                  "PORT=3000"
                  "PGHOST=/run/postgresql"
                  "PGDATABASE=civ6"
                  "PGUSER=civ6"
                  "NODE_PATH=${web}/node_modules"
                ];
              };
            };

            services.caddy = {
              enable = true;
              virtualHosts."civ6.ch".extraConfig = ''
                reverse_proxy localhost:3000
              '';
            };
          };
        };
    };
}
