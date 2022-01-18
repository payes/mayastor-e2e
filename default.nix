{
    crossSystem ? null
}:

let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs {
    overlays = [
      (_: _: { inherit sources; })
    ];
    inherit crossSystem;
  };
in
pkgs
