{ pkgs, ... }:
let
  nix-pre-commit-hooks = import (builtins.fetchTarball
    "https://github.com/cachix/git-hooks.nix/tarball/master");
in nix-pre-commit-hooks.run {
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
    golangci-lint = {
      enable = true;
      package = pkgs.golangci-lint;
      extraPackages = with pkgs; [ go openssl ];
      stages = [ "pre-push" ]; # because it takes a while
    };
    gotest = {
      enable = true;
      package = pkgs.go;
      extraPackages = with pkgs; [ openssl gcc ];
      stages = [ "pre-push" ]; # because it takes a while
    };
  };
}
