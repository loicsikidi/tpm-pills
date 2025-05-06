# Install tooling in Your Running System

The goal here is to prepare your environment to run the examples provided in `TPM Pills`, if (like me) you only trust what you see with your own eyes. **It's not a requirement**. Indeed, you can limit yourself to read the content and code snippets. However, I strongly recommend you to read and run the examples to better understand the concepts.

## Disclaimer regarding Windows  

So far, my experience with TPMs has been exclusively limited to a **Linux** context — this is why, I am open to [feedback](https://github.com/loicsikidi/tpm-pills/issues) from Windows users, if they encounter any issues.  

<div class="warning">
<b>Warning</b>

Unfortunately, according to <a href="https://github.com/microsoft/WSL/issues/5178" target="_blank">this issue</a>, TPM is not added to WSL (<em>Windows Subsystem for Linux</em>), so it will be necessary to run commands from the host machine.
</div>

## Prerequisites  

`TPM Pills` will require you to have the following tools (in addition to `git`):  

| Tool | Description | Linux Support | Windows Support | MacOS Support |  
| ---- | ----------- | :-----------: | :-------------: | :-----------: |  
|  **[go](https://go.dev/)** >= `v1.22` | A language that no longer needs an introduction | ✅ | ✅ | ✅ |  
|  **[openssl](https://github.com/openssl/openssl)** | Crypto *Swiss Army Knife* which here is a dependency for using the Software TPM | ✅ | ✅ | ✅ |
|  **[swtpm](https://github.com/stefanberger/swtpm)** | A Software TPM Emulator | ✅ |  ✅ <br><br> ([using Cygwin](https://github.com/stefanberger/swtpm/wiki#compile-and-install-on-cygwin)) |  ✅ |  
|  **[tpm2-tools](https://github.com/tpm2-software/tpm2-tools)** | A CLI (*Command-Line Interface*) for interacting with a TPM | ✅ | ❌ | ❌ |  

`tpm2-tools` is a great tool to have in your *toolbox*! However, since it is not available everywhere, it will be used sparingly.  

> *Note: `PowerShell` provides some [commands](https://learn.microsoft.com/en-us/powershell/module/trustedplatformmodule) to interact with a TPM, but they will not be covered here.* 

### Why `go`?  

Most educational content on the subject is in **C**... but why follow the crowd, right?!

More seriously:  

* I am not an experienced **C** developer, but I am proficient in **Go**  
* [`go-tpm`](https://github.com/google/go-tpm) provides a rich interface for communicating with a TPM  
* In the upcoming *pills*, we will make the TPM interact with a server in gRPC, and **Go** allows me to do this easily
* More and more projects in Golang ecosystem use TPMs (e.g., [spire](https://github.com/spiffe/spire), [sks](https://github.com/facebookincubator/sks), [u-root](https://github.com/u-root/u-root), [constellation](https://github.com/edgelesssys/constellation), etc.)

Fundamentally, since the `TPM 2.0` interface is a standard, all the concepts we will cover here are also applicable in other languages.  

#### TPM 2.0 Clients  

For those interested, here is a (probably non-exhaustive) list of TPM 2.0 clients.  

| Name | Language | Description |
| ---- | -------- | ----------- |
| [tpm2-tss](https://github.com/tpm2-software/tpm2-tss) | C | Intel implementation of TCG's TPM Software Stack (TSS). The current standard meter bar regarding TPM libraries. |
| [ibmtss](https://github.com/kgoldman/ibmtss) | C | IBM implementation of TPM Software Stack (TSS) but not API compatible with TCG TSS. |
| [wolfTPM](https://github.com/wolfSSL/wolfTPM) | C | TPM 2.0 librairy designed for embedded system. |
| [go-tpm](https://github.com/google/go-tpm) | golang | |
| [tpm2-pytss](https://github.com/tpm2-software/tpm2-pytss) | python | Wrapper of `tpm2-tss`. |
| [tpm-rs](https://github.com/tpm-rs/tpm-rs)| rust | |
| [rust-tss-fapi](https://github.com/tpm2-software/rust-tss-fapi)| rust | Wrapper of `libtss2-fapi` which is an upper API provided by `tpm2-tss` named FAPI[^1]. <br><br>***Warning**: project's maintainers underline that the implementation is experimental and shouldn't be use in production*. |
| [TSS.MSR](https://github.com/microsoft/TSS.MSR)| c#, c++, java, nodejs and python | |

## Installation  

### OCI

🚧 TBD 🚧

### Nix  

If you are a Nix user, `TPM Pills` provides a Nix shell (i.e. `shell.nix`) at the root of the repository.  

To install dependencies, simply run the following commands:  

```bash
git clone https://github.com/loicsikidi/tpm-pills.git
cd ./tpm-pills
# launch the derministic shell
nix-shell

# inside the shell
go version
tpm2 --version
```

> *Note: with this method `tpm2-tools` will only be installed on a Linux platform.*

### Devbox  

<div class="info">
<b>Info</b>

 For those who are unfamiliar, <a href="https://github.com/jetify-com/devbox" target="_blank">Devbox</a> is a layer on top of <b>Nix</b> that allows you to obtain a deterministic shell without having to master Nix language.
</div>
  
If you are a Devbox user, `TPM Pills` also provides a configuration (i.e. `devbox.json`) at the root of the repository.  

To install dependencies, simply run the following commands:  

```bash
git clone https://github.com/loicsikidi/tpm-pills.git
cd ./tpm-pills
# launch the derministic shell
devbox shell

# inside the shell
go version
tpm2 --version
```

> *Note: with this method `tpm2-tools` will only be installed on a Linux platform.*

### Manually  

* **go**: Use your preferred *package manager* or download the binary from the [official website](https://go.dev/doc/install)  
* **openssl**: Use your preferred *package manager* or get the sources from the [official website](https://openssl-library.org/source/)
* **swtpm**: Use your preferred package manager or build the sources by following the [official documentation](https://github.com/stefanberger/swtpm/wiki)
* **tpm2-tools**: Use your preferred *package manager* or build the sources by following the [official documentation](https://tpm2-tools.readthedocs.io/en/latest/INSTALL/)

## Example: Validate TPM's version  

Let’s finally get to the serious stuff! We will check the version of the TPM installed on your machine and ensure that it is a `TPM 2.0`. We will able to do this by interacting directly with the TPM using a command called `TPM2_GetCapability`.

<div class="info">
<b>Info</b>

On Linux, access to the Hardware TPM is secured by <b>sudo</b> rights. It is possible to have finer control using a <code class="hljs">udev</code> policy to allow specific users or groups to access it (e.g., the policy available in <a href="https://github.com/NixOS/nixpkgs/blob/88a55dffa4d44d294c74c298daf75824dc0aafb5/nixos/modules/security/tpm2.nix#L10-L21" target="_blank">NixOS</a>).
</div>

### tpm2-tools  

<div class="warning">
<b>Warning</b>

<em>Only works on Linux.</em>
</div>

```bash
# dependending on your config it might require 'sudo'
tpm2_getcap properties-fixed | grep -i pt_family_indicator -A 2
```

You should get the following output:  

![](./images/02-pill/get_cap_cmd.png)

### go  

The script will works on all environments (on `Darwin`, the code relies on a Software TPM).

Run the following command:  

```bash
# dependending on your config it might require 'sudo'
go run github.com/loicsikidi/tpm-pills/examples/02-pill

# output:
# TPM Version: 2.0
```

Depending on your local setup, you can also run the following command:  

```bash
# nix command
nix-shell --run "go run github.com/loicsikidi/tpm-pills/examples/02-pill"
# devox command
devbox run -- go run github.com/loicsikidi/tpm-pills/examples/02-pill
```

## Next pill...

...we will see in much more details how we can interact with a TPM.

---

🚧 `TPM Pills` is in **beta** 🚧

* if you encounter problems 🙏 please report them on the [tpm-pills](https://github.com/loicsikidi/tpm-pills/issues) issue tracker
* if you think that `TPM Pills` should cover a specific topic which isn't in the [roadmap](https://github.com/loicsikidi/tpm-pills/blob/main/ROADMAP.md), let's initiate a [discussion](https://github.com/loicsikidi/tpm-pills/discussions/new?category=ideas) 💬

[^1]: Feature API
