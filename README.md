# gh-eco

🦎 gh extension to explore the ecosystem

[![GitHub Go
Workflow](https://github.com/jrnxf/gh-eco/actions/workflows/ci.yml/badge.svg)](https://github.com/jrnxf/gh-eco/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-default.svg)](./LICENSE.md) [![Github
Stars](https://img.shields.io/github/stars/jrnxf/gh-eco)](https://github.com/jrnxf/gh-eco/stargazers)

![demo](https://github.com/jrnxf/assets/raw/main/gh-eco/demo.gif)

## Installation

1. Install the `gh` cli - see the [installation](https://github.com/cli/cli#installation)

   _Installation requires a minimum version (2.0.0) of the GitHub CLI that supports extensions._

2. Install this extension:

   ```sh
   gh extension install jrnxf/gh-eco
   ```

<details>
   <summary><strong>Manual Installation</strong></summary>

> If you want to install this extension manually, follow these steps:

1. clone the repo

   ```sh
   # git
   git clone https://github.com/jrnxf/gh-eco

   # GitHub CLI
   gh repo clone jrnxf/gh-eco
   ```

2. `cd` into it

   ```sh
   cd gh-eco
   ```

3. build it

   ```sh
   go build
   ```

4. install it locally
   ```sh
   gh extension install .
   ```
   </details>

## Usage

to run:

```sh
gh eco
```

or view an organisation:

```sh
gh eco <organisation>
```

gh-eco follows your terminal's light or dark background; if it guesses wrong, set `GH_ECO_THEME=light` or `GH_ECO_THEME=dark`.

## Keys

| Key | Action |
| --- | --- |
| `/` | focus input / search |
| `enter` | search |
| `j` / `↓` | move down |
| `k` / `↑` | move up |
| `ctrl+d` | page down (readme) |
| `ctrl+u` | page up (readme) |
| `r` | toggle readme |
| `o` | open in github |
| `s` | star repo |
| `!` | star gh-eco |
| `esc` / `ctrl+c` | quit |

to upgrade:

```sh
gh extension upgrade eco
```

## Roadmap

- user and repo search with partial matches
- "see more" repos
- followers/following lists
- trending

## Contributing

All contributions are greatly appreciated!

If you have a suggestion that would make `gh-eco` better, please fork the repo and create a [pull
request](https://github.com/jrnxf/gh-eco/pulls) or open an issue.

See the [open issues](https://github.com/jrnxf/gh-eco/issues) for a full list of proposed
features (and known bugs).

_Regenerate the gif with [vhs](https://github.com/charmbracelet/vhs) and ffmpeg: `./demo.sh`._

## License

Distributed under the MIT License. See [LICENSE.md](./LICENSE.md) for more information.

## Acknowledgments

Check out these amazing projects that inspired `gh-eco`!

- anything and everything by [charm.sh](https://charm.sh/)
- [gh-dash](https://github.com/dlvhdr/gh-dash)

## Follow

[![github](https://img.shields.io/github/followers/jrnxf?style=social)](https://github.com/jrnxf)
[![twitter](https://img.shields.io/twitter/follow/_jrnxf?color=white&style=social)](https://twitter.com/_jrnxf)
[![youtube](https://img.shields.io/youtube/channel/subscribers/UCEDfokz6igeN4bX7Whq49-g?style=social)](https://www.youtube.com/@jrnxf)
