{pkgs ? import <nixpkgs> {}}: let
  mdbook-sitemap-generator =
    pkgs.callPackage ./nix/pkgs/mdbook-sitemap-generator.nix {};
in
  with pkgs; {
    html-split = stdenvNoCC.mkDerivation {
      name = "tpm-pills";
      src = lib.cleanSource ./.;

      nativeBuildInputs = [mdbook mdbook-linkcheck mdbook-sitemap-generator];

      buildPhase = ''
        runHook preBuild

        # We can't check external links inside the sandbox, but it's good to check them outside the sandbox.
        substituteInPlace book.toml --replace-fail 'follow-web-links = true' 'follow-web-links = false'
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
