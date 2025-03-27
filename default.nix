{ pkgs ? import <nixpkgs> { } }:
with pkgs; {
  html-split = stdenvNoCC.mkDerivation {
    name = "tpm-pills";
    src = lib.cleanSource ./.;

    nativeBuildInputs = [ mdbook mdbook-linkcheck ];

    buildPhase = ''
      runHook preBuild

      # We can't check external links inside the sandbox, but it's good to check them outside the sandbox.
      substituteInPlace book.toml 'follow-web-links = true' 'follow-web-links = false'
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
}
