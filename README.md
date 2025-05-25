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

## License

This work is copyright Loïc Sikidi and licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-nc-sa/4.0/).
