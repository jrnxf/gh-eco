# gh-eco

🦎 gh extension to explore the ecosystem

[![GitHub Go
Workflow](https://github.com/jrnxf/gh-eco/actions/workflows/ci.yml/badge.svg)](https://github.com/jrnxf/gh-eco/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-default.svg)](./LICENSE.md) [![Github
Stars](https://img.shields.io/github/stars/jrnxf/gh-eco)](https://github.com/jrnxf/gh-eco/stargazers)

![demo](https://github.com/jrnxf/assets/raw/main/gh-eco/demo.gif)

_Regenerate it with [vhs](https://github.com/charmbracelet/vhs) and ffmpeg: `./demo.sh`._

## Installation

1. Install the `gh` cli - see the [installation](https://github.com/cli/cli#installation)

   _Installation requires a minimum version (2.0.0) of the the GitHub CLI that supports extensions._

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

3. add dependencies and build it

   ```sh
   go get && go build
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

gh-eco follows your terminal's light or dark background; if it guesses wrong, set `GH_ECO_THEME=light` or `GH_ECO_THEME=dark`.

to upgrade:

```sh
gh extension upgrade eco
```

## Roadmap

**🎨 Custom Configurations**

Allowing users to customize the colors of all displayed elements is definitely a priority. Right now
the colors used should be adaptive between standard light/dark themed terminal profiles, however
it's entirely possible for the colors to still clash! Beyond colors, having the ability to select
what elements are displayed, their ordering, custom keymaps, etc. would be awesome!

**⚡️ Interactions**

I'm intentionally releasing `gh-eco` while I still have ideas bouncing around in my head for what
could come next. There are some immediate limitations that I need to sort out. One of which is
dealing with gh permissions. Right now all permissions are inherited from the `gh` cli, which only
requests a small subset of available scopes. One interaction I have built out is to follow/unfollow
users (see `feature/follow-users` branch), but the interaction is blocked due to missing scopes. I
haven't found any documentation or examples showcasing what extending permissions for a `gh`
extension might look like, so I may need to contact support. I'd rather not make it a full blown
standalone app that requires it's own auth, as I think that might complicate things? If you have
opinions please don't hesitate to reach out!

Other interactions I think would be nice include:

- searching for users and seeing results (i.e. perfect match not needed)
  - searching for repos as well (possibly using `u/{username}` and `r/{repo}` to differentiate
    searches)
- "see more" option to view more than just pinned repos
- viewing and/or filtering through a list of the followed/following users to jump to their profile
- repo contributors and related stats
- creative ways to show more user info while still keeping things clean (e.g. company, isHireable,
  status, etc). These are easy to display but I'm intentionally keeping things lightweight for now
- trending users, repos, etc.

## Limitations

- Some emoji are drawn at different widths by different terminals, so a card border next to one
  can sit a cell or two off.

## Contributing

All contributions are greatly appreciated!

If you have a suggestion that would make `gh-eco` better, please fork the repo and create a [pull
request](https://github.com/jrnxf/gh-eco/pulls) or open an issue.

See the [open issues](https://github.com/jrnxf/gh-eco/issues) for a full list of proposed
features (and known bugs).

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
