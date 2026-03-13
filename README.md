# mweb

A command-line dictionary and thesaurus powered by the [Merriam-Webster Collegiate API](https://dictionaryapi.com/).

## Installation

```bash
go install github.com/user/mweb@latest
```

Or clone and build locally:

```bash
git clone <repo-url>
cd dict_cli
go install .
```

## Setup

You need two API keys from [dictionaryapi.com](https://dictionaryapi.com/register/index) — one for the **Collegiate Dictionary** and one for the **Collegiate Thesaurus**. Both are free for personal use.

Create the config file at `~/.config/mweb/config.yaml` (Linux/macOS) or `%AppData%\Roaming\mweb\config.yaml` (Windows):

```yaml
api_key_dict:      "your_collegiate_dictionary_key"
api_key_thesaurus: "your_collegiate_thesaurus_key"
output_format:     "plain"   # or "json"
max_definitions:   5
```

Alternatively, use environment variables:

```bash
export MWEB_API_KEY_DICT=your_dict_key
export MWEB_API_KEY_THESAURUS=your_thesaurus_key
```

## Usage

```bash
mweb def <word>   # define a word
mweb syn <word>   # synonyms
mweb ant <word>   # antonyms
```

### Examples

```
$ mweb def run

run (verb)
  1 a. to go faster than a walk
  1 b. to flee, retreat
  2.  to contend in a race
  ...

run (noun)
  1 a. an act or the action of running
  ...
```

```
$ mweb syn happy

Synonyms for "happy" (adjective):
  glad, joyful, blissful, cheerful, content
```

```
$ mweb ant happy

Antonyms for "happy" (adjective):
  sad, unhappy, depressed, miserable
```

### Flags

| Flag     | Description                  |
|----------|------------------------------|
| `--json` | Output raw JSON from the API |

```bash
mweb --json def serendipity
```

## Project structure

```
dict_cli/
├── main.go
├── cmd/
│   ├── root.go       # Cobra root command, config loading, --json flag
│   ├── def.go
│   ├── syn.go
│   └── ant.go
└── internal/
    ├── api/
    │   ├── client.go      # HTTP client, two-pass JSON decode
    │   ├── dictionary.go  # Dictionary types and Define()
    │   ├── thesaurus.go   # Thesaurus types and Thesaurus()
    │   └── errors.go      # WordNotFoundError
    ├── config/
    │   └── config.go      # Viper-backed config loader
    └── output/
        └── formatter.go   # Plain text and JSON rendering
```

## Development

```bash
go test ./...        # run all tests
go build -o mweb .  # build binary
go install .        # install to $GOPATH/bin
```
