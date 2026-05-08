# TPM Pills

Content and examples for [TPM Pills](https://tpmpills.com), a site that introduces *Trusted Platform Modules* (TPM) with a series of short articles.

## Building

The repository provides a Nix definition which embed everything:

```bash
nix-build -A html-split && open result/tpm-pills/index.html
```

If you are not familiar with Nix, to build the site locally, you will need to have [mdbook](https://github.com/rust-lang/mdBook) + [mdbook-linkcheck](https://github.com/Michael-F-Bryan/mdbook-linkcheck) and run:

```bash
mdbook build && open result/tpm-pills/index.html
```

## Dependency Update Policy

> [!NOTE]
> For those interested in understanding the motivations behind this approach, I recommend reading [Filippo Valsorda's thoughts on Dependabot](https://words.filippo.io/dependabot/).

This project does not rely on automated dependency update tools like Dependabot. When managing multiple projects in parallel, such tools generate more noise than value.

Instead, this project follows a pragmatic, security-first approach:

1. **`govulncheck` runs daily** to detect vulnerable dependencies. When a vulnerability is identified → we bump the affected dependency.
2. **Feature-driven updates**: Dependencies are updated when the project needs a new feature provided by a newer version.
3. **`go test` runs daily** with the latest dependency versions to detect breaking changes early.

This approach balances security with intentionality, ensuring updates happen for concrete reasons rather than on autopilot.

## License

This work is copyright Loïc Sikidi and licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-nc-sa/4.0/).
