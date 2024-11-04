# NYT Reply Bot

A mastodon bot that looks for NYT Games social shares, Cooking links, and articles to remind folks that we're on strike.

## Getting started

### What you'll need

- Go (latest version)
- [Task](https://taskfile.dev/installation/). See [Taskfile](./Taskfile.yml)
- [magic-wormhole](http://magic-wormhole.io) for sharing the `.env` file

### Local development

```shell
task run
```

```shell
task build
```

### Sharing the .env file

On the sending computer:

```shell
brew install magic-wormhole
wormhole send .env
Sending 215 Bytes file named '.env'
Wormhole code is: 4-cellulose-skullcap
On the other computer, please run:

wormhole receive 4-cellulose-skullcap
```

On the receiving computer:

```shell
brew install magic-wormhole
wormhole receive 4-cellulose-skullcap
```
Note that you'll need to repeat the process until everyone has a copy
