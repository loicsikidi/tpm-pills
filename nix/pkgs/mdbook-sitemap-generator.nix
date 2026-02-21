{
  fetchFromGitHub,
  buildGoModule,
}:
buildGoModule rec {
  pname = "mdbook-sitemap-generator";
  version = "1.2.2";

  src = fetchFromGitHub {
    owner = "loicsikidi";
    repo = "mdbook-sitemap-generator";
    tag = "v${version}";
    hash = "sha256-FZ8tfIvosu6L0jWex/uyg+w6QYYxyBSqYWsqu2low+0=";
  };

  vendorHash = "sha256-cEgvwog50izBOyMlCdLI2KvwSHPKZsp1wSw6a59V1yw=";
}
