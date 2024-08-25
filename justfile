# vim: set tabstop=4 shiftwidth=4 expandtab:
# https://linux.101hacks.com/ps1-examples/prompt-color-using-tput/

BOLD := "$(tput bold)"
RESET := "$(tput sgr0)"
BLACK := "$(tput bold)$(tput setaf 0)"
RED := "$(tput bold)$(tput setaf 1)"
GREEN := "$(tput bold)$(tput setaf 2)"
YELLOW := "$(tput bold)$(tput setaf 3)"
BLUE := "$(tput bold)$(tput setaf 4)"
MAGENTA := "$(tput bold)$(tput setaf 5)"
CYAN := "$(tput bold)$(tput setaf 6)"
WHITE := "$(tput bold)$(tput setaf 7)"
BLACKB := "$(tput bold)$(tput setab 0)"
REDB := "$(tput setab 1)$(tput setaf 0)"
GREENB := "$(tput setab 2)$(tput setaf 0)"
YELLOWB := "$(tput setab 3)$(tput setaf 0)"
BLUEB := "$(tput setab 4)$(tput setaf 0)"
MAGENTAB := "$(tput setab 5)$(tput setaf 0)"
CYANB := "$(tput setab 6)$(tput setaf 0)"
WHITEB := "$(tput setab 7)$(tput setaf 0)"
PGPATH := "PATH=$(brew --prefix)/opt/postgresql@16/bin:" + env_var("PATH")

[private]
default:
    just --list

# Check this file formatting
fmt:
    just --fmt --check --unstable

[private]
brew:
    @which brew >/dev/null \
    || (echo {{ GREEN }}🍺 Installing Homebrew...{{ RESET }} \
        && NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" \
        && echo {{ REDB }}{{ WHITE }} 👆 You must follow NEXT STEPS above before continuing 👆 {{ RESET }})

[private]
brew-linux-shell:
    @echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"'

[private]
just: brew
    @[ -f $(brew--prefix)/bin/just ] \
    || brew install just

[private]
just0:
    curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | sudo bash -s -- --to /usr/local/bin

[private]
imagemagick: brew
    @[ -d $(brew --prefix)/opt/imagemagick ] \
    || brew install imagemagick

PGDATA := "PGDATA=$(brew --prefix)/var/postgresql@16"
PG := PGPATH + " " + PGDATA

[private]
postgres: brew
    @[ -d $(brew --prefix)/opt/postgresql@16 ] \
    || brew install postgresql@16

[private]
gpg: brew
    @[ -d $(brew --prefix)/opt/gpg ] \
    || brew install gpg

# https://tldp.org/LDP/abs/html/exitcodes.html
[private]
asdf:
    @which asdf >/dev/null \
    || (brew install asdf \
        && echo {{ REDB }}{{ WHITE }} 👆 You must follow CAVEATS above before continuing 👆 {{ RESET }})

[private]
asdf-shell: brew
    @echo "source $(brew --prefix)/opt/asdf/libexec/asdf.sh"

# Setup everything needed to run the app locally
local: asdf brew imagemagick postgres gpg
    @awk '{ system("asdf plugin-add " $1) }' < .tool-versions
    @asdf install

export ELIXIR_ERL_OPTIONS := if os() == "linux" { "+fnu" } else { "" }

# Install all app dependencies
deps:
    mix deps.get --only dev
    mix deps.get --only test

[private]
pg_ctl:
    @{{ PG }} which pg_ctl >/dev/null \
    || (echo "Please install Postgres using: {{ BOLD }}just local{{ RESET }}" && exit 127)

# Start Postgres server
postgres-up: pg_ctl
    @({{ PG }} pg_ctl status | grep -q "is running") || {{ PG }} pg_ctl start

# Stop Postgres server
postgres-down: pg_ctl
    @({{ PG }} pg_ctl status | grep -q "no server running") || {{ PG }} pg_ctl stop

[private]
postgres-db db:
    @({{ PG }} psql --list --quiet --tuples-only | grep -q {{ db }}) \
    || {{ PG }} createdb {{ db }}

export DB_USER := `whoami`

[private]
changelog_test: postgres-up (postgres-db "changelog_test")

# Run app tests
test: changelog_test
    mix test

[private]
changelog_dev: postgres-up (postgres-db "changelog_dev")
    mix ecto.setup

[private]
yarn:
    @which yarn >/dev/null \
    || (echo "Please install Node.js & Yarn using: {{ BOLD }}just local{{ RESET }}" && exit 127)

[private]
assets: yarn
    cd assets && yarn install

# Run app in dev mode
dev: changelog_dev assets
    mix phx.server

# Run everything needed for your first contribution
contribute: local
    #!/usr/bin/env zsh
    eval "$(just asdf-shell)"
    just deps
    just test
    just dev

# Run this in a local GitHub Actions Runner container
actions-runner:
    docker run --interactive --tty \
        --volume=changelog-linuxbrew:/home/linuxbrew/.linuxbrew \
        --volume=changelog-asdf:/home/runner/.asdf \
        --volume=.:/home/runner/work --workdir=/home/runner/work \
        --env=HOST=$(hostname) --publish=4000:4000 \
        --pull=always ghcr.io/actions/actions-runner

[linux]
[private]
ubuntu:
    sudo apt update
    DEBIAN_FRONTEND=noninteractive sudo apt install -y build-essential curl git libncurses5-dev libssl-dev inotify-tools

[linux]
do-it:
    #!/usr/bin/env bash
    time just ubuntu brew
    eval "$(just brew-linux-shell)"
    time just asdf
    eval "$(just asdf-shell)"
    time just local
    time just test
    just dev

# just actions-runner
# DEBIAN_FRONTEND=noninteractive sudo apt install -y curl
# eval $(grep j_ust.systems justfile)
# cd ../
# sudo chown -fR runner:runner work
# cd work
# just do-it
#
# OR
#
# time just ubuntu brew
# eval "$(just brew-linux-shell)"
# just asdf
# eval "$(just asdf-shell)"
# time just local
# time just test
# just dev
#
# du -skh _build
# 72M
#
# du -skh deps
# 41M
#
# du -skh /home/linuxbrew/.linuxbrew
# 1.5G
# du -skh ~/.asdf
# 728M
