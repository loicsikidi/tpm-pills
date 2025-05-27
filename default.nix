{ pkgs ? import <nixpkgs> { }, domain ? "" }:
let
  mdbook-sitemap-generator =
    pkgs.callPackage ./nix/pkgs/mdbook-sitemap-generator.nix { };
  defaultInputs = [ pkgs.mdbook pkgs.mdbook-linkcheck ];
  conditionalElement =
    if domain != "" then [ mdbook-sitemap-generator ] else [ ];
  nativeBuildInputs = defaultInputs ++ conditionalElement;
in with pkgs; {
  html-split = stdenvNoCC.mkDerivation {
    name = "tpm-pills";
    src = lib.cleanSource ./.;

    nativeBuildInputs = nativeBuildInputs;

    buildPhase = ''
      runHook preBuild

      # We can't check external links inside the sandbox, but it's good to check them outside the sandbox.
      substituteInPlace book.toml --replace-fail 'follow-web-links = true' 'follow-web-links = false'
      mdbook build

      if [ -n "${domain}" ]; then
        echo "Generating sitemap.xml for domain: ${domain}"
        mdbook-sitemap-generator \
          --domain "${domain}" \
          --output book/html/sitemap.xml 
      fi

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
  pre-commit-check = callPackage ./nix/pre-commit.nix { };
}
