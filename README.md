# gh-eco
🦎 github cli (gh) extension to explore the ecosystem


[![GitHub Go Workflow](https://github.com/coloradocolby/gh-eco/actions/workflows/go.yml/badge.svg)](https://github.com/coloradocolby/gh-eco/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/License-MIT-default.svg)](./LICENSE.md)
[![Github Stars](https://img.shields.io/github/stars/coloradocolby/gh-eco)](https://github.com/coloradocolby/gh-eco/stargazers)


## Installation

1. install the `gh` cli - see the [installation](https://github.com/cli/cli#installation)
   
   _Installation requires a minimum version (2.0.0) of the the GitHub CLI that supports extensions._

2. install this extension:

   ```sh
   gh extension install coloradocolby/gh-eco
   ```

<details>
   <summary><strong>Manual Installation</strong></summary>

> If you want to install this extension **manually**, follow these steps:

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

## Contributing

All contributions are **greatly appreciated**.

If you have a suggestion that would make `gh-eco` better, please fork the repo and
create a [pull request](https://github.com/coloradocolby/gh-eco/pulls). You can
also simply open an issue and select `Feature Request`

1. Fork the repo
2. Create your feature branch (`git checkout -b [your_username]/xyz`)
3. Commit your changes (`git commit -m 'add some xyz'`)
4. Rebase off main (`git fetch --all && git rebase origin/main`)
5. Push to your branch (`git push origin [your_username]/xyz`)
6. Fill out pull request template

See the [open issues](https://github.com/coloradocolby/gh-eco/issues) for a full
list of proposed features (and known issues).

## License

Distributed under the MIT License. See [LICENSE.md](./LICENSE.md) for more
information.

## Acknowledgments

Check out these amazing projects that inspired `gh-eco`!

- literally everything by [charm.sh](https://charm.sh/) (what gh-eco is built on)
- [gh-dash](https://github.com/dlvhdr/gh-dash)

## Follow

[![github](https://img.shields.io/github/followers/coloradocolby?style=social)](https://github.com/coloradocolby)

[![twitter](https://img.shields.io/twitter/follow/coloradocolby?color=white&style=social)](https://twitter.com/coloradocolby)

[![youtube](https://img.shields.io/youtube/channel/subscribers/UCEDfokz6igeN4bX7Whq49-g?style=social)](https://youtube.com/user/coloradocolby)
