{ pkgs ? import <nixpkgs> { } }:
let
  nix-pre-commit-hooks = import (builtins.fetchTarball
    "https://github.com/cachix/git-hooks.nix/tarball/master");
in with pkgs; {
  html-split = stdenvNoCC.mkDerivation {
    name = "tpm-pills";
    src = lib.cleanSource ./.;

    nativeBuildInputs = [ mdbook mdbook-linkcheck ];

    buildPhase = ''
      runHook preBuild

      # We can't check external links inside the sandbox, but it's good to check them outside the sandbox.
      # substituteInPlace book.toml --replace-fail 'follow-web-links = true' 'follow-web-links = false'
      mdbook build

      runHook postBuild
    '';

    installPhase = ''
      runHook preInstall

      dst=$out/tpm-pills
      mkdir -p "$dst"
      mv book/html/* "$dst"/

      runHook postInstall
    '';
  };
  pre-commit-check = nix-pre-commit-hooks.run {
    src = ./.;
    # If your hooks are intrusive, avoid running on each commit with a default_states like this:
    # default_stages = [ "manual" "pre-push" ];
    hooks = {
      # common
      end-of-file-fixer.enable = true;
      # nix
      nixfmt-classic.enable = true;
      # golang
      gofmt.enable = true;
      golangci-lint.enable = true;
      golangci-lint.package = pkgs.golangci-lint;
      golangci-lint.extraPackages = with pkgs; [ go openssl ];
      gotest.enable = true;
      gotest.package = pkgs.go;
      gotest.extraPackages = with pkgs; [ openssl gcc ];
    };
  };
}
