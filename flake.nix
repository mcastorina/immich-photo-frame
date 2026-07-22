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
        # Headless display for running the Fyne app over SSH/VNC.
        xorgserver
        x11vnc
        mesa
        mesa-demos
      ];
    };
  };
}
