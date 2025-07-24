{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = { systems, nixpkgs, ...}@inputs:
  let
    eachSystem = f: nixpkgs.lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
  in {
    devShells = eachSystem (pkgs: {
      default = pkgs.mkShell {
        buildInputs = [
          pkgs.gnumake
          pkgs.go
          pkgs.insomnia
        ];
        shellHook = ''
          export CONTAINER_NAME='guac-api-dev'

          command -v docker > /dev/null 2>&1 && export CONTAINER_PROG='docker'
          command -v podman > /dev/null 2>&1 && export CONTAINER_PROG='podman'

          if ! $CONTAINER_PROG ps --format '{{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
            echo "Starting development containers..."

            export WORKING_DIR=$(mktemp -d)
            chmod 0755 $WORKING_DIR

            export POSTGRES_PW=$(tr -dc 'A-Za-z0-9' </dev/urandom | head -c 20; echo)

            $CONTAINER_PROG run --rm docker.io/guacamole/guacamole:latest /opt/guacamole/bin/initdb.sh --postgresql > $WORKING_DIR/initdb.sql

            while :; do
              CONTAINER_PORT=$(( ( RANDOM % 64512 ) + 1024 ))
              if ! ss -tuln | grep -q ":$CONTAINER_PORT\b"; then
                break
              fi
            done

            $CONTAINER_PROG network create $CONTAINER_NAME-net

            $CONTAINER_PROG run -d --rm \
              --name $CONTAINER_NAME-postgres \
              --network $CONTAINER_NAME-net \
              -e POSTGRES_DB=guacamole_db \
              -e POSTGRES_USER=guacamole_user \
              -e POSTGRES_PASSWORD=$POSTGRES_PW \
              -v $WORKING_DIR:/docker-entrypoint-initdb.d \
              docker.io/postgres:17-alpine

            $CONTAINER_PROG run -d --rm \
              --name $CONTAINER_NAME-guacd \
              --network $CONTAINER_NAME-net \
              docker.io/guacamole/guacd:latest

            $CONTAINER_PROG run -d --rm \
              --name $CONTAINER_NAME \
              --network $CONTAINER_NAME-net \
              -e GUACD_HOSTNAME=$CONTAINER_NAME-guacd \
              -e POSTGRES_HOSTNAME=$CONTAINER_NAME-postgres \
              -e POSTGRES_DATABASE=guacamole_db \
              -e POSTGRES_USER=guacamole_user \
              -e POSTGRES_PASSWORD=$POSTGRES_PW \
              -p $CONTAINER_PORT:8080 \
              docker.io/guacamole/guacamole:latest
          fi

          export CONTAINER_PORT=$($CONTAINER_PROG port $CONTAINER_NAME 8080 | cut -d: -f2)

          export GUACAMOLE_URL=http://localhost:$CONTAINER_PORT/guacamole
          export GUACAMOLE_USERNAME=guacadmin
          export GUACAMOLE_PASSWORD=guacadmin

          echo -e "Guacamole Address:\n\t$GUACAMOLE_URL\n\t$GUACAMOLE_USERNAME : $GUACAMOLE_PASSWORD"

          nohup ${pkgs.insomnia}/bin/insomnia > /dev/null 2>&1 &
          nohup xdg-open "$GUACAMOLE_URL" > /dev/null 2>&1 &

          trap 'echo "Stopping development containers..."; $CONTAINER_PROG network rm -f $CONTAINER_NAME-net' EXIT
        '';
      };
    });
  };
}
