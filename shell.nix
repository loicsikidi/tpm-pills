let
  # mdbook pinned to 0.4.52
  # go to https://www.nixhub.io/packages/mdbook to the list of available versions
  nixpkgs =
    fetchTarball
    "https://github.com/NixOS/nixpkgs/archive/ee09932cedcef15aaf476f9343d1dea2cb77e261.tar.gz";
  pkgs = import nixpkgs {
    config = {};
    overlays = [];
  };
  mdbook-sitemap-generator =
    pkgs.callPackage ./nix/pkgs/mdbook-sitemap-generator.nix {};

  helpers = import (builtins.fetchTarball
    "https://github.com/loicsikidi/nix-shell-toolbox/tarball/main") {
    inherit pkgs;
    hooksConfig = {
      gofmt.enable = false;
      treefmt = {
        enable = true;
        stages = ["pre-push"];
      };
      gotest.settings.flags = "-race";
      lychee = {
        settings.flags = "./pills"; # check only pills directory
        stages = ["pre-commit"];
      };
    };
  };

  # see https://github.com/NixOS/nixpkgs/issues/355486#issuecomment-2746187811
  # to remember why we did this hack to fix the following error in github actions:
  # '# runtime/cgo'
  # 'Multiple conflicting values defined for DEVELOPER_DIR_arm64_apple_darwin
  mkShellNoCC = pkgs.mkShellNoCC.override {
    stdenv = pkgs.stdenvNoCC.override {extraBuildInputs = [];};
  };
in
  mkShellNoCC {
    packages = with pkgs;
      [
        mdbook
        mdbook-sitemap-generator

        # required to run TPM simulator
        # source: https://github.com/google/go-tpm-tools/tree/main/simulator
        gcc
        openssl

        swtpm

        # we install tpm2-tools only in supported platforms (i.e. linux)
        (lib.optional (lib.elem stdenv.system tpm2-tools.meta.platforms) tpm2-tools)
      ]
      ++ helpers.packages;
    # we disable the hardening due to this error: https://github.com/tpm2-software/tpm2-tools/issues/1561
    # fix found here: https://github.com/NixOS/nixpkgs/issues/18995#issuecomment-249748307
    hardeningDisable = ["fortify"];

    shellHook = ''
      ${helpers.shellHook}
      echo "Development environment ready!"
      echo "  - Go version: $(go version)"
    '';

    env = {
      CGO_ENABLED = "1";

      # Disable warnings from TPM simulator C code
      CGO_CFLAGS = "-Wno-array-bounds -Wno-stringop-overflow";
    };
  }
