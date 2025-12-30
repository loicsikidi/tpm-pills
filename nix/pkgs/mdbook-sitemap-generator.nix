{
  fetchFromGitHub,
  buildGoModule,
}:
buildGoModule rec {
  pname = "mdbook-sitemap-generator";
  version = "1.2.0";

  src = fetchFromGitHub {
    owner = "loicsikidi";
    repo = "mdbook-sitemap-generator";
    tag = "v${version}";
    hash = "sha256-CGj6sWgOzI2cimX7KE0TkXOR6KlTOk2ZbIb3c5X6TSk=";
  };

  vendorHash = "sha256-WUQW8EDJ7kT2CUZsNtlVUVwwqFRHkpkU6pFmx7/MDGg=";
}
