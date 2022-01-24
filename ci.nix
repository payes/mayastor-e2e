{
    crossSystem ? null
}:

let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs {
    overlays =
      [ (_: _: { inherit sources; }) ];
  };
in
with pkgs;
mkShell {
  name = "mayastor-e2e-shell";
  buildInputs = [
    git
    pre-commit
    python3
  ];

}
