{
  description = "ipf";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  };

  outputs = { self, nixpkgs }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs {
      inherit system;
    };
    runHeadless = pkgs.writeShellApplication {
      name = "run-headless";
      runtimeInputs = with pkgs; [ xorgserver x11vnc mesa ];
      text = ''
        if [ -z "''${1:-}" ]; then
          echo "usage: $(basename "$0") <vnc-password>" >&2
          exit 1
        fi

        pw_file=$(mktemp)
        trap 'rm -f "$pw_file"' EXIT

        Xvfb :99 -screen 0 1920x1080x24 >/dev/null 2>&1 &
        x11vnc -storepasswd "$1" "$pw_file"
        x11vnc -display :99 -rfbauth "$pw_file" -localhost -forever -bg >/dev/null 2>&1

        export DISPLAY=:99
        export LIBGL_ALWAYS_SOFTWARE=1
        ./ipf
      '';
    };
  in {
    devShells.${system}.default = pkgs.mkShell {
      packages = with pkgs; [
        go
        gopls
        pkg-config
        libGL
        libX11
        libxrandr
        libxinerama
        libxi
        libxxf86vm
        libxcursor
        # Helper script.
        runHeadless
      ];
    };
  };
}
