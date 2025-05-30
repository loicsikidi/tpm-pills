{ fetchFromGitHub, buildGoModule, }:

buildGoModule rec {
  pname = "mdbook-sitemap-generator";
  version = "1.1.0";

  src = fetchFromGitHub {
    owner = "loicsikidi";
    repo = "mdbook-sitemap-generator";
    tag = "v${version}";
    hash = "sha256-BA6qfkdVxAq1XENZ1s2uCi2CGJ9FMmFNbF1YZvansiw=";
  };

  vendorHash = "sha256-h5UXs7ujP0YIBKusstTDXvWBhqKpYs6HtzzdaB1+6Wg=";

}
