{ fetchFromGitHub, buildGoModule, }:

buildGoModule rec {
  pname = "mdbook-sitemap-generator";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "loicsikidi";
    repo = "mdbook-sitemap-generator";
    tag = "v${version}";
    hash = "sha256-V6/FYyE5+ESPk8ybUK4eGsX9CVPh3640CkoI/7lKc3Y=";
  };

  vendorHash = "sha256-CVycV7wxo7nOHm7qjZKfJrIkNcIApUNzN1mSIIwQN0g=";

}
