{
  fetchFromGitHub,
  buildGoModule,
}:
buildGoModule rec {
  pname = "mdbook-sitemap-generator";
  version = "1.3.0";

  src = fetchFromGitHub {
    owner = "loicsikidi";
    repo = "mdbook-sitemap-generator";
    tag = "v${version}";
    hash = "sha256-DNFdHQpvL6BaFMk0LVBkTB5eX93cUXEeKueTBoFIO8o=";
  };

  vendorHash = "sha256-5uDi/9YlUNRDQKFROlN8sLvLRFFOvZvZHcmR6fARG5Q=";
}
