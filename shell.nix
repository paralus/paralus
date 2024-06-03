{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/refs/tags/24.05-beta.tar.gz") { } }:
pkgs.mkShell {
  hardeningDisable = [ "fortify" ]; # needed for dlv to work (https://github.com/NixOS/nixpkgs/issues/18995)
  buildInputs = with pkgs; [
    # go
    go_1_22
    buf
    golangci-lint
    go-migrate
    protobuf

    # test
    moq

    #db schema
    atlas

    # protoc plugins
    protoc-gen-go
    protoc-gen-go-grpc
    grpc-gateway # adds protoc-gen-grpc-gateway and protoc-gen-openapiv2 

    # debugging
    delve

    # other
    gnumake
  ];

  GOPRIVATE = "github.com/RafaySystems/*";
}
