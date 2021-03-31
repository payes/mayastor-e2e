{ nospdk ? false
, norust ? false
}:
let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs {
    overlays = [
      (_: _: { inherit sources; })
    ];
  };
in
with pkgs;
mkShell {

  # fortify does not work with -O0 which is used by spdk when --enable-debug
  hardeningDisable = [ "fortify" ];
  buildInputs = [
    docker-compose
    kubectl
    kind
    docker
    cowsay
    e2fsprogs
    envsubst # for e2e tests
    gdb
    go
    golangci-lint
    git
    gptfdisk
    kubernetes-helm
    llvmPackages.libclang
    nodejs-12_x
    numactl
    meson
    ninja
    openssl
    pkg-config
    pre-commit
    procps
    python3
    utillinux
    xfsprogs
  ]
  ;

  shellHook = ''
    pre-commit install --hook commit-msg
  '';
}
