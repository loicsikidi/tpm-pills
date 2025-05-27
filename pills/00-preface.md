# Preface

## Motivation

When I started my journey with TPMs (*Trusted Platform Modules*), I experienced two conflicting emotions: `enthusiasm and confusion`. Indeed, the principle of TPM seemed brilliant to me, but I struggled to understand how I could actually put it into practice.

After several months of perseverance, and thanks to the valuable work of the community (i.e. tools, blog posts, etc.), I reached a level of understanding that could be considered acceptable. In hindsight, I can clearly say that my learning cost was too high, but **it doesn't necessarily have to be for everyone**.

That is why, I am taking on the challenge of producing a (relatively) comprehensive introduction to the subject. In short, the content I would have dreamed of having when I started this journey.

## What's `TPM Pills`?

<div class="info">
<b>Info</b>

<code class="hljs">TPM Pills</code> is a direct tribute to <a href="https://nixos.org/guides/nix-pills" target="_blank"><code class="hljs">Nix Pills</code></a>, who has helped many people discover the <a href="https://nixos.org" target="_blank">nix</a> language!
</div>

A series of articles that gradually introduce the key concepts about a TPM. The goal is that by the end of reading `TPM Pills`, you will have a solid understanding of the functionalities offered by a TPM in order to reduce the complexity of TPM 2.0 specification[^1] if you find yourself reading it. Additionally, each article will be accompanied by a reproductible example to make things more concrete.

> *Note: `TPM Pills` will refer a bunch of time to the TPM 2.0 specification. Knowing that the latter evolves over time, we will use the latest version[^2] available at the time of writing.*  
>
> *The spec is also available in [here](https://github.com/loicsikidi/tpm-pills/tree/main/assets/resources) in the repository.*

Finally, it is important to emphasize that this content is **free**.

## Who is this for?

To anyone who wants to understand TPM and its functionalities. Whether you are a developer, a security expert, or just curious, you will find something to satisfy your curiosity.

<div class="warning">
<b>Warning</b>

A developer background is recommended especially for the implementation part.
</div>

## Other educational resources

If you want to explore the topic further or if the `TPM Pills` approach simply doesn't suit you, be aware that there are other alternatives:

| Resource | Description | Format |
| --- | --- | --- |
| [A Practical Guide to TPM 2.0](https://link.springer.com/book/10.1007/978-1-4302-6584-9) | At the time of writing, the most comprehensive book on the subject (my bedside book)! <br><br>***Note: PDF format is free*** | Book |
|  Trusted Platform Module (TPM) courses | <ul><li>[1101: Introductory usage](https://p.ost2.fyi/courses/course-v1:OpenSecurityTraining2+TC1101_IntroTPM+2024_v2/about)</li><li>[1102: Intermediate usage](https://p.ost2.fyi/courses/course-v1:OpenSecurityTraining2+TC1102_IntermediateTPM+2024_v1/about)</li></ul> ***Note: courses are free*** | Online course |
| [TPM.dev tutorials](https://github.com/tpm2dev/tpm.dev.tutorials) | To share developer-friendly resources about Trusted Platform Modules (TPM) and hardware security, including other Hardware Security Modules (HSM). <br><br>*Note: description from the repo*  | Tutorials |
| [TPM-JS by Google](https://github.com/tpm2dev/tpm.dev.tutorials) | TPM-JS lets you experiment with a software TPM device in your browser. It's an educational tool that teaches you how to use a TPM device to secure your workflows. <br><br>*Note: description from the repo*<br><br>***Warning: the repo is archived since 2022***  | Tutorials |
| [TPMCourse by Nokia](https://github.com/nokia/TPMCourse) | A short course on getting started with understanding how a TPM 2.0 works. In this course we explain a number of the features of the TPM 2.0 through the TPM2_Tools through examples and, optionally, exercises.<br><br>*Note: description from the repo*  | Tutorials |

<p align="center"><b>Table: </b><em>Educational Resources</em></p>

## Who Am I?

I'm **Loïc Sikidi** a *passionate software engineer* from France. I love to learn and share my (little bit of) knowledge with others.

I'm far from being an expert on the subject, but I want to contribute to the democratization of this technology because I'm convinced that the TPM is a powerful tool that can help us to build more secure systems.

---

🚧 `TPM Pills` is in **beta** 🚧

* if you encounter problems 🙏 please report them on the [tpm-pills](https://github.com/loicsikidi/tpm-pills/issues) issue tracker
* if you think that `TPM Pills` should cover a specific topic which isn't in the [roadmap](https://github.com/loicsikidi/tpm-pills/blob/main/ROADMAP.md), let's initiate a [discussion](https://github.com/loicsikidi/tpm-pills/discussions/new?category=ideas) 💬

[^1]: The specification available [here](https://trustedcomputinggroup.org/resource/tpm-library-specification) is a dense and relatively complex document.
[^2]: Version 184 released on March 2025
