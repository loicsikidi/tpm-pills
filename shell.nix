let
  # mdbook pinned to 0.4.50
  # go to https://www.nixhub.io/packages/mdbook to the list of available versions
  nixpkgs = fetchTarball
    "https://github.com/NixOS/nixpkgs/archive/bdac72d387dca7f836f6ef1fe547755fb0e9df61.tar.gz";
  pkgs = import nixpkgs {
    config = { };
    overlays = [ ];
  };
  pre-commit = import ./default.nix { };
  mdbook-sitemap-generator =
    pkgs.callPackage ./nix/pkgs/mdbook-sitemap-generator.nix { };

  # see https://github.com/NixOS/nixpkgs/issues/355486#issuecomment-2746187811
  # to remember why we did this hack to fix the following error in github actions:
  # '# runtime/cgo'
  # 'Multiple conflicting values defined for DEVELOPER_DIR_arm64_apple_darwin
  mkShellNoCC = pkgs.mkShellNoCC.override {
    stdenv = pkgs.stdenvNoCC.override { extraBuildInputs = [ ]; };
  };
in mkShellNoCC {
  packages = with pkgs; [
    mdbook
    mdbook-linkcheck
    mdbook-sitemap-generator

    go # v1.24.3
    delve # v1.24.2

    # required to run TPM simulator
    # source: https://github.com/google/go-tpm-tools/tree/main/simulator
    gcc
    openssl

    swtpm

    # we install tpm2-tools only in supported platforms (i.e. linux)
    (lib.optional (lib.elem stdenv.system tpm2-tools.meta.platforms) tpm2-tools)
  ];
  # we disable the hardening due to this error: https://github.com/tpm2-software/tpm2-tools/issues/1561
  # fix found here: https://github.com/NixOS/nixpkgs/issues/18995#issuecomment-249748307
  hardeningDisable = [ "fortify" ];

  shellHook = ''
    ${pre-commit.pre-commit-check.shellHook}
  '';
  buildInputs = pre-commit.pre-commit-check.enabledPackages;
}
