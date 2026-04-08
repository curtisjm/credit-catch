{
  description = "Credit Catch - Credit card benefits tracker";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        pgdata = ".postgres";
        pgport = "5432";
        pguser = "creditcatch";
        pgdb = "creditcatch";
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go
            go
            gopls
            gotools

            # Node.js
            nodejs_22
            pnpm

            # Database
            postgresql_16

            # Tools
            jq
            curl
          ];

          shellHook = let
            db-start = pkgs.writeShellScriptBin "db-start" ''
              export PGDATA="''${PGDATA:-$PWD/${pgdata}}"
              if [ ! -d "$PGDATA" ]; then
                echo "Initializing postgres..."
                ${pkgs.postgresql_16}/bin/initdb --no-locale --encoding=UTF8 -U ${pguser} > /dev/null
                echo "unix_socket_directories = '$PGDATA'" >> "$PGDATA/postgresql.conf"
                echo "port = ${pgport}" >> "$PGDATA/postgresql.conf"
                echo "listen_addresses = 'localhost'" >> "$PGDATA/postgresql.conf"
              fi
              if ! ${pkgs.postgresql_16}/bin/pg_ctl status -D "$PGDATA" > /dev/null 2>&1; then
                ${pkgs.postgresql_16}/bin/pg_ctl start -D "$PGDATA" -l "$PGDATA/postgres.log" -o "-k $PGDATA"
                sleep 1
                ${pkgs.postgresql_16}/bin/createdb -h "$PGDATA" -p ${pgport} ${pgdb} 2>/dev/null || true
                echo "Postgres running on port ${pgport}"
              else
                echo "Postgres already running"
              fi
            '';
            db-stop = pkgs.writeShellScriptBin "db-stop" ''
              export PGDATA="''${PGDATA:-$PWD/${pgdata}}"
              ${pkgs.postgresql_16}/bin/pg_ctl stop -D "$PGDATA" 2>/dev/null || echo "Postgres not running"
            '';
            db-reset = pkgs.writeShellScriptBin "db-reset" ''
              export PGDATA="''${PGDATA:-$PWD/${pgdata}}"
              ${db-stop}/bin/db-stop
              rm -rf "$PGDATA"
              ${db-start}/bin/db-start
              echo "Database reset"
            '';
          in ''
            export PGDATA="$PWD/${pgdata}"
            export PGHOST="$PWD/${pgdata}"
            export PGPORT="${pgport}"
            export PGUSER="${pguser}"
            export PGDATABASE="${pgdb}"
            export DATABASE_URL="postgres://${pguser}@localhost:${pgport}/${pgdb}?host=$PGHOST&sslmode=disable"
            export JWT_SECRET="dev-secret-do-not-use-in-production"
            export PORT="8080"
            export ENVIRONMENT="development"

            export GOBIN="$PWD/.go/bin"
            export PATH="${db-start}/bin:${db-stop}/bin:${db-reset}/bin:$GOBIN:$PATH"

            if [ ! -f "$GOBIN/migrate" ]; then
              echo "Installing golang-migrate..."
              go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest 2>/dev/null
            fi

            echo "credit-catch dev environment loaded"
            echo "  db-start  — start postgres"
            echo "  db-stop   — stop postgres"
            echo "  db-reset  — wipe and reinit postgres"
          '';
        };
      });
}
