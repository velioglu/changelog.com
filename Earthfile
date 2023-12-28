VERSION --global-cache 0.7

# https://github.com/elixir-lang/elixir/tags
ARG --global elixir=1.14.5
# https://github.com/erlang/otp/tags
ARG --global erlang=26.2
# https://hub.docker.com/r/hexpm/elixir/tags?page=1&ordering=last_updated&name=ubuntu-jammy
ARG --global ubuntu=jammy-20231004
# https://nodejs.org/en/about/previous-releases
ARG --global nodejs=20.10.0

# This image is the used by all other environments: dev, test & prod
# It is published as https://github.com/thechangelog/changelog.com/pkgs/container/changelog-runtime
image-runtime:
  FROM hexpm/elixir:$elixir-erlang-$erlang-ubuntu-$ubuntu
  RUN mix --version
  RUN mix local.rebar --force
  RUN mix local.hex --force
  RUN apt-get update
  RUN apt-get install --yes git-core && git --version
  # Install convert (imagemagick), required for image resizing to work...
  RUN apt-get install --yes imagemagick && convert --version
  # Install gcc (build-essential), required to install cmark...
  # https://hexdocs.pm/cmark/readme.html#prerequisites
  RUN apt-get install --yes build-essential && gcc --version
  # Install inotify-tools for dynamic website reloads while developing...
  RUN apt-get install --yes inotify-tools && which inotifywatch
  # Install postgresql-client for manual commands while developing...
  RUN apt-get install --yes postgresql-client && psql --version
  RUN apt-get install --yes curl && curl --version
  ARG nodejs_platform=linux-x64
  ARG nodejs_name=node-v$nodejs-$nodejs_platform
  RUN cd /opt && curl --silent --fail --location --remote-name https://nodejs.org/download/release/v$nodejs/$nodejs_name.tar.xz
  RUN cd /opt && tar -xJf $nodejs_name.tar.xz
  ENV PATH="/opt/$nodejs_name/bin:$PATH"
  RUN node --version && npm --version
  LABEL org.opencontainers.image.description="💜 Elixir v$elixir | 🚜 Erlang v$erlang | ⬢ Node.js v$nodejs | 🐡 Ubuntu $ubuntu"
  LABEL org.opencontainers.image.source=https://github.com/thechangelog/changelog.com
  ARG user=thechangelog
  SAVE IMAGE --push ghcr.io/$user/changelog-runtime:elixir-v$elixir-erlang-v$erlang-nodejs-v$nodejs

app:
  FROM +image-runtime
  COPY --dir config lib test mix.exs mix.lock /app
  COPY --dir priv/repo /app/priv
  WORKDIR /app

test:
  FROM +app
  RUN --mount=type=cache,target=/app/deps
  ENV MIX_ENV=test
  RUN mix deps.get --only $MIX_ENV && ls -lahd /app/deps/*
  RUN --mount=type=cache,target=/app/_build/$MIX_ENV
  RUN mix do deps.compile + compile && ls -lahd /app/_build/$MIX_ENV/lib/*/ebin
