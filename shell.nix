{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "go-shell";

  buildInputs = with pkgs; [
    go
    gcc
    sqlite
  ];

  shellHook = ''
    echo "Welcome to your development environment!"
    export GOROOT=${pkgs.go}
    export GOPATH=./modules
  '';
}
