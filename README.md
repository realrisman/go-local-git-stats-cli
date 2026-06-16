# go-local-git-stats-cli

[![CI](https://github.com/realrisman/go-local-git-stats-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/realrisman/go-local-git-stats-cli/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/realrisman/go-local-git-stats-cli)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![codecov](https://codecov.io/gh/realrisman/go-local-git-stats-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/realrisman/go-local-git-stats-cli)

A small command-line tool that renders a GitHub-style contribution graph for your
**local** Git repositories, right in the terminal. It scans folders for repos, counts
the commits authored by a given email over the last six months, and prints them as a
colored grid of weeks × days.

```
            Jan       Feb       Mar       Apr       May       Jun
      -  -  -  -  -  1  -  -  2  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
 Mon  -  -  3  -  -  -  -  -  1  -  -  -  -  -  -  4  -  -  -  -  -  -  -  -  -  -
 ...
```

## How it works

- Repository paths are stored in a dotfile at `~/.go-local-git-stats-cli` (one path
  per line). The `-add` flag scans a folder tree and appends any Git repos it finds.
- Running without `-add` reads that dotfile and walks each repo's commit history from
  `HEAD`, tallying commits whose author email matches `-email` per day for the last
  183 days.
- The graph is drawn with ANSI colors; each cell's shade reflects the commit count for
  that day, and today's cell is highlighted.

## Installation

Requires Go 1.26+.

```sh
git clone https://github.com/realrisman/go-local-git-stats-cli.git
cd go-local-git-stats-cli
go build -o go-local-git-stats-cli .
```

Or run directly with `go run .`.

## Usage

### 1. Register repositories to track

Point `-add` at a folder; the tool recursively finds every Git repository beneath it
(skipping `vendor` and `node_modules`) and saves their paths to the dotfile.

```sh
go-local-git-stats-cli -add /Users/you/Codes
```

Paths starting with `~` are expanded to your home directory, so quoting is safe:

```sh
go-local-git-stats-cli -add "~/Codes"
```

Run it for as many folders as you like — paths are de-duplicated, so re-adding is safe.

### 2. Show your contribution graph

```sh
go-local-git-stats-cli -email you@example.com
```

Only commits whose author email matches are counted, so set this to the email you use
in your local Git config.

## Flags

| Flag     | Default          | Description                                              |
| -------- | ---------------- | -------------------------------------------------------- |
| `-add`   | `""`             | Folder to scan for Git repositories and register.        |
| `-email` | `your@email.com` | Author email to count commits for when drawing the graph. |

## Color legend

| Commits in a day | Cell color  |
| ---------------- | ----------- |
| 0                | dim / blank |
| 1–4              | white       |
| 5–9              | yellow      |
| 10+              | green       |
| today            | magenta     |

## Notes

- Only the **last ~6 months (183 days)** are graphed.
- The tool reads local history only — nothing is sent anywhere.

## License

[MIT](LICENSE)
