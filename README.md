# gh-eco

🦎 github cli (gh) extension to explore the ecosystem

[![GitHub Go
Workflow](https://github.com/coloradocolby/gh-eco/actions/workflows/go.yml/badge.svg)](https://github.com/coloradocolby/gh-eco/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/License-MIT-default.svg)](./LICENSE.md) [![Github
Stars](https://img.shields.io/github/stars/coloradocolby/gh-eco)](https://github.com/coloradocolby/gh-eco/stargazers)

![demo](./assets/demo.gif)

## Installation

1. Install the `gh` cli - see the [installation](https://github.com/cli/cli#installation)

   _Installation requires a minimum version (2.0.0) of the the GitHub CLI that supports extensions._

2. Install this extension:

   ```sh
   gh extension install coloradocolby/gh-eco
   ```

<details>
   <summary><strong>Manual Installation</strong></summary>

> If you want to install this extension manually, follow these steps:

1. clone the repo

   ```sh
   # git
   git clone https://github.com/coloradocolby/gh-eco

   # GitHub CLI
   gh repo clone coloradocolby/gh-eco
   ```

2. `cd` into it

   ```sh
   cd gh-eco
   ```

3. build it

   ```sh
   gh build
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

**🧪 Tests**

To date only a small amount of tests have been written. I feel comfortable writing tests for smaller
utility functions, but I'd love to do more integration type testing on the underlying [Bubble
Tea](https://github.com/charmbracelet/bubbletea) implementation. If you have any experience with
this please feel free to open an issue / pull request and start the discussion!

## Limitations

- Misalignment can occur when emojis are present. Since emoji widths vary, I don't believe there is
  any way to deal with this. Definitely reach out if you have any ideas.

- Smaller terminal screens will almost definitely cause rendering issues. For `gh-eco` to work
  properly make sure to give your terminal as much screen real estate as possible!

## Contributing

All contributions are greatly appreciated!

If you have a suggestion that would make `gh-eco` better, please fork the repo and create a [pull
request](https://github.com/coloradocolby/gh-eco/pulls) or open an issue.

See the [open issues](https://github.com/coloradocolby/gh-eco/issues) for a full list of proposed
features (and known bugs).

## License

Distributed under the MIT License. See [LICENSE.md](./LICENSE.md) for more information.

## Acknowledgments

Check out these amazing projects that inspired `gh-eco`!

- anything and everything by [charm.sh](https://charm.sh/)
- [gh-dash](https://github.com/dlvhdr/gh-dash)

## Follow

[![github](https://img.shields.io/github/followers/coloradocolby?style=social)](https://github.com/coloradocolby)

[![twitter](https://img.shields.io/twitter/follow/coloradocolby?color=white&style=social)](https://twitter.com/coloradocolby)

[![youtube](https://img.shields.io/youtube/channel/subscribers/UCEDfokz6igeN4bX7Whq49-g?style=social)](https://youtube.com/user/coloradocolby)
