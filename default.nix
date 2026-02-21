{pkgs ? import <nixpkgs> {}}: let
  mdbook-sitemap-generator =
    pkgs.callPackage ./nix/pkgs/mdbook-sitemap-generator.nix {};
in
  with pkgs; {
    html-split = stdenvNoCC.mkDerivation {
      name = "tpm-pills";
      src = lib.cleanSource ./.;

      nativeBuildInputs = [mdbook mdbook-sitemap-generator];

      buildPhase = ''
        runHook preBuild

        echo "mdbook version: $(mdbook --version)"

        mdbook build

        runHook postBuild
      '';

      installPhase = ''
        runHook preInstall

        dst=$out/tpm-pills
        mkdir -p "$dst"
        mv book/html/* "$dst"/
        mv book/sitemap-generator/* "$dst"/

        runHook postInstall
      '';
    };
  }
