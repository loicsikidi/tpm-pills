let
  # mdbook pinned to 0.4.45
  # go to https://www.nixhub.io/packages/mdbook to the list of available versions
  nixpkgs = fetchTarball
    "https://github.com/NixOS/nixpkgs/archive/dad564433178067be1fbdfcce23b546254b6d641.tar.gz";
  pkgs = import nixpkgs {
    config = { };
    overlays = [ ];
  };
  pre-commit = import ./default.nix { };
in pkgs.mkShellNoCC {
  packages = with pkgs; [
    mdbook
    mdbook-linkcheck

    go # v1.23.5
    delve # v1.24.0

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
