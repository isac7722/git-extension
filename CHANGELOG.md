# Changelog

## [1.2.0](https://github.com/isac7722/git-extension/compare/v1.1.0...v1.2.0) (2026-03-11)


### Features

* **branch:** add interactive branch switcher with `ge branch` ([01d2851](https://github.com/isac7722/git-extension/commit/01d285124cc21bf43b31d2afadd39f737be909f5))
* **pr:** add `ge pr` command for creating GitHub pull requests ([a301505](https://github.com/isac7722/git-extension/commit/a30150511e3f236006551dfcd36da3f501a8ca23))
* **pr:** add `ge pr` command for creating GitHub pull requests ([ee194fd](https://github.com/isac7722/git-extension/commit/ee194fd36325085b2bd867c49d0c64d33b15b490))

## [1.1.0](https://github.com/isac7722/git-extension/compare/v1.0.0...v1.1.0) (2026-03-11)


### Features

* **user:** add `ge user update` command for editing profiles ([f5dc9b2](https://github.com/isac7722/git-extension/commit/f5dc9b23e7b6507542699f13690cdf63d9c6884e))
* **user:** auto-apply profile after `ge user update` ([b00e6d7](https://github.com/isac7722/git-extension/commit/b00e6d7926e8df770218de2728c9207c38ed015e))

## 1.0.0 (2026-03-11)


### ⚠ BREAKING CHANGES

* rewrite ge-cli in Go with TUI and open-source docs
* auto-switch to worktree directory after creation

### Features

* add ge clean command and restructure CLI architecture ([36927f5](https://github.com/isac7722/git-extension/commit/36927f5ac266ace8db7bee4325d51fd1863876e7))
* add interactive branch selector and improve worktree safety ([182d343](https://github.com/isac7722/git-extension/commit/182d343816f916ea5c55603288cacd049be6240f))
* add interactive worktree remove selector with confirmation prompt ([6e064d6](https://github.com/isac7722/git-extension/commit/6e064d6c169d6053040efe31eb6da84b487ca137))
* add interactive worktree selector and worktree remove by branch ([0affb54](https://github.com/isac7722/git-extension/commit/0affb542fb28cd5a303716e307e7f177be1d0dc7))
* auto-switch to worktree directory after creation ([aa36f36](https://github.com/isac7722/git-extension/commit/aa36f3635a81c1bbf22e50ecabf2843429bb7236))
* auto-switch to worktree directory after creation ([8bc1858](https://github.com/isac7722/git-extension/commit/8bc18585877e1dbf55048cb81a687878b390fd62))
* enhance worktree list with status indicators ([9fa2f6c](https://github.com/isac7722/git-extension/commit/9fa2f6c88fe9ff602a0cbcf160a95537c960fa37))
* make `ge user list` an interactive selector with keyboard navigation ([ad5e728](https://github.com/isac7722/git-extension/commit/ad5e728e8bd544ba15f0ee8f659f48348559c20a))
* rewrite ge-cli in Go with TUI and open-source docs ([e7d342e](https://github.com/isac7722/git-extension/commit/e7d342e50ffd6175ec6e4f66ff2328d952a5a7e2))


### Bug Fixes

* correct ANSI escape code padding in interactive user list ([a0e6f08](https://github.com/isac7722/git-extension/commit/a0e6f08aa79e9f106cd82a83f6136bced8da6417))
* move local declarations outside loop in worktree list ([71e3711](https://github.com/isac7722/git-extension/commit/71e3711bd3d251ced0389047e796049921e5aff9))
* resolve golangci-lint errors for errcheck, staticcheck, and unused ([66374fd](https://github.com/isac7722/git-extension/commit/66374fdda9c0d4a9f437cc15dabee6d4acb73dc6))
* **tui:** use WithInputTTY to prevent stdin buffer leak between TUI programs ([28d5682](https://github.com/isac7722/git-extension/commit/28d5682a9771a522af0f25b7afbebb589d97962c))
