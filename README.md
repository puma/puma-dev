# Puma-dev: A fast, zero-config development server for macOS and Linux

[![Build Status](https://travis-ci.org/puma/puma-dev.svg?branch=master)](https://travis-ci.org/puma/puma-dev)

Puma-dev is the emotional successor to pow. It provides a quick and easy way to manage apps in development on macOS and Linux.

## Highlights

* Easy startup and idle shutdown of rack/rails apps
* Easy access to the apps using the `.test` subdomain **(configurable)**
* Run multiple custom domains at the same time, e.g. `.test`, `.puma`.

### Why choose puma-dev?
* __https__ - it Just Works!
* Supports __Rails 5 actioncable__ via rack.hijack websockets
* Supports macOS __and__ Linux
* The honorary `pow` [is no longer maintained](https://github.com/basecamp/pow/commit/310f260d08159cf86a52df7ddb5a3bd53a94614f)

## Installation
First, ensure that the [`puma`](https://github.com/puma/puma) gem is installed. It probably belongs in the Gemfile of the application(s) you're trying to serve via puma-dev.

```ruby
# Gemfile
gem 'puma'
```

### Homebrew on macOS
`brew install puma/puma/puma-dev`

### Pre-built Binaries

You may download binaries for macOS and Linux at [https://github.com/puma/puma-dev/releases](https://github.com/puma/puma-dev/releases)

### Build from Source

```shell
#!/usr/bin/env bash

go version

go get github.com/puma/puma-dev/...
cd $GOPATH/src/github.com/puma/puma-dev/
make && make install

$GOBIN/puma-dev -V
```

------

## macOS Support

### Install & Setup

```shell
# Configure some DNS settings that have to be done as root
sudo puma-dev -setup
# Configure puma-dev to run in the background on ports 80 and 443 with the domain `.test`.
puma-dev -install
```

If you wish to have `puma-dev` use a port other than 80, pass it via the `-install-port`, for example to use port 81: `puma-dev -install -install-port 81`.

*NOTE:* If you installed puma-dev v0.2, please run `sudo puma-dev -cleanup` to remove firewall rules that puma-dev no longer uses (and will conflict with puma-dev working).

If you're currently using `pow`, puma-dev taking control of `.test` will break it. If you want to just try out puma-dev and leave pow working, pass `-d pdev` on `-install` to use the `.pdev` as an alternate development TLD.

*NOTE:* If you had pow installed before in the system, please make sure to run pow's uninstall script. Read more details in [the pow manual](http://pow.cx/manual.html#section_1.2).

### Uninstall

Run: `puma-dev -uninstall`

*NOTE:* If you passed custom options (e.g. `-d test:localhost`) to `-setup`, be sure to pass them to `-uninstall` as well. Otherwise `/etc/resolver/*` might contain orphaned entries.

### Logging

When puma-dev is installed as a user agent (the default mode), it will log output from itself and the apps to `~/Library/Logs/puma-dev.log`. You can refer to there to find out if apps have started and look for errors.

In the future, puma-dev will provide an integrated console for this log output.

------

## Linux Support

Puma-dev supports Linux but requires the following additional installation steps to be followed to make all the features work (`-install` and `-setup` flags for Linux are not provided):

### puma-dev root CA

The puma-dev root CA is generated (in `~/.puma-dev-ssl/`), but you will need to install and trust this as a Certificate Authority by adding it to your operating system's certificate trust store, or by trusting it directly in your favored browser (as some browsers will not share the operating system's trust store).

### Domains (.test or similar)

In order for requests to the `.test` (or any other custom) domain to resolve, install the [dev-tld-resolver](https://github.com/puma/dev-tld-resolver), making sure to use `test` (or the custom TLD you want to use) when configuring TLDs.

### Port 80/443 binding

Linux prevents applications from binding to ports lower that 1024 by default. You don't need to bind to port 80/443 to use puma-dev but it makes using the `.test` domain much nicer (e.g. you'll be able to use the domain as-is in your browser rather than providing a port number)

There are 2 options to allow puma-dev to listen on port 80 and 443:

1. Give puma-dev the capabilities directly:
  ```shell
  sudo setcap CAP\_NET\_BIND\_SERVICE=+eip /path/to/puma-dev
  ```
or
2. Install `authbind`. and invoke puma-dev with it when you want to use it e.g.
  ```shell
  authbind puma-dev -http-port 80 -https-port 443
  ```

There is a shortcut for binding to 80/443 by passing `-sysbind` to puma-dev when starting, which overrides `-http-port` and `-https-port`.

### Systemd (running puma-dev in the background)

On Linux, puma-dev will not automatically run in the background (as per the MacOS `-install` script); you'll need to [run it in the foreground](#running-in-the-foreground). You can set up a system daemon to start up puma-dev in the background yourself.

1. Create `/lib/systemd/system/puma-dev.service` and put in the following:
  ```
  [Unit]
  After=network.target

  [Service]
  User=$USER
  ExecStart=/path/to/puma-dev -sysbind
  Restart=on-failure

  [Install]
  WantedBy=multi-user.target
  ```

  Replace `path/to/puma-dev` with an absolute path to puma-dev
  Replace the `$USER` variable with the name of the user you want to run under.

2. Start puma-dev using systemd:
  ```shell
  sudo systemctl daemon-reload
  sudo systemctl enable puma-dev
  sudo systemctl start puma-dev
  ```

------

## Usage

Simply symlink your app's directory into `~/.puma-dev`! That's it!

You can use the built-in helper subcommand: `puma-dev link [-n name] [dir]` to link app directories into your puma-dev directory (`~/.puma-dev` by default).

### Options
Run: `puma-dev -h`

You have the ability to configure most of the values that you'll use day-to-day.

### Advanced Configuration

Puma-dev supports loading environment variables before puma starts. It checks for the following files in this order:

* `~/.powconfig`
* `.env`
* `.powrc`
* `.powenv`

Additionally, puma-dev uses a few environment variables to control how puma is started that you can overwrite in your loaded shell config.

* `CONFIG`: A puma configuration file to load, usually something like `config/puma-dev.rb`. Defaults to no config.
* `THREADS`: How many threads puma should use concurrently. Defaults to 5.
* `WORKERS`: How many worker processes to start. Defaults to 0, meaning only use threads.

### Important Note On Ports and Domain Names

* Default privileged ports are 80 and 443
* Default domain is `.test`.
  * Previously it was `.dev`, but it is owned by Google and since Dec 2017 **HSTS only** with real websites hosted there.
  * Don't use `.dev` and `.foo`, as they are real TLDs.
* Using pow? To avoid conflicts, use different ports and domain or [uninstall pow properly](http://pow.cx/manual.html#section_1.2).

### Restarting

If you would like to have puma-dev restart *a specific app*, you can run `touch tmp/restart.txt` in that app's directory.

### Purging

If you would like to have puma-dev stop *all the apps* (for resource issues or because an app isn't restarting properly), you can send `puma-dev` the signal `USR1`. The easiest way to do that is:

`puma-dev -stop`

### Running in the foreground

Run: `puma-dev`

Puma-dev will startup by default using the directory `~/.puma-dev`, looking for symlinks to apps just like pow. Drop a symlink to your app in there as: `cd ~/.puma-dev; ln -s /path/to/my/app test`. You can now access your app as `test.test`.

Running `puma-dev` in this way will require you to use the listed http port, which is `9280` by default.

### Coming from v0.2

Puma-dev v0.3 and later use launchd to access privileged ports, so if you installed v0.2, you'll need to remove the firewall rules.

Run: `sudo puma-dev -cleanup`

### Coming from Pow

By default, puma-dev uses the domain `.test` to manage your apps. If you want to have puma-dev look for apps in `~/.pow`, just run `puma-dev -pow`.

### Sub Directories

If you have a more complex set of applications you want puma-dev to manage, you can use subdirectories under `~/.puma-dev` as well. This works by naming the app with a hyphen (`-`) where you'd have a slash (`/`) in the hostname. So for instance if you access `cool-frontend.test`, puma-dev will look for `~/.puma-dev/cool-frontend` and if it finds nothing, try `~/.puma-dev/cool/frontend`.

### Proxy support

Puma-dev can also proxy requests from a nice dev domain to another app. To do so, just write a file (rather than a symlink'd directory) into `~/.puma-dev` with the connection information.

For example, to have port 9292 show up as `awesome.test`: `echo 9292 > ~/.puma-dev/awesome`.

Or to proxy to another host: `echo 10.3.1.2:9292 > ~/.puma-dev/awesome-elsewhere`.

### HTTPS

Puma-dev automatically makes the apps available via SSL as well. When you first run puma-dev, it will have likely caused a dialog to appear to put in your password. What happened there was puma-dev generates its own CA certification that is stored in `~/Library/Application Support/io.puma.dev/cert.pem`.

That CA cert is used to dynamically create certificates for your apps when access to them is requested. It automatically happens, no configuration necessary. The certs are stored entirely in memory so future restarts of puma-dev simply generate new ones.

When `-install` is used (and let's be honest, that's how you want to use puma-dev), then it listens on port 443 by default (configurable with `-install-https-port`) so you can just do `https://blah.test` to access your app via https.

### Websockets

Puma-dev supports websockets natively but you may need to tell your web framework to allow the connections.

In the case of rails, you need to configure rails to allow all websockets or websocket requests from certain domains. The quickest way is to add `config.action_cable.disable_request_forgery_protection = true` to `config/environments/development.rb`. This will allow all websocket connections while in development.

*Do not use disable_request_forgery_protection in production!*

Or you can add something like `config.action_cable.allowed_request_origins = /(\.test$)|^localhost$/` to allow anything under `.test` as well as `localhost`.

### xip.io

Puma-dev supports `xip.io` domains. It will detect them and strip them away, so that your `test` app can be accessed as `test.A.B.C.D.xip.io`.

### Run multiple domains

Puma-dev allows you to run multiple local domains. Handy if you're working with more than one client. Simply set up puma-dev like so: `puma-dev -install -d first-domain:second-domain`

### Static file support

Like pow, puma-dev support serving static files. If an app has a `public` directory, then any urls that match files within that directory are served. The static files have priority over the app.

### Subdomains support

Once a virtual host is installed, it's also automatically accessible from all subdomains of the named host. For example, a `myapp` virtual host could also be accessed at `http://www.myapp.test/` and `http://assets.www.myapp.test/`. You can override this behavior to, say, point `www.myapp.test` to a different application: just create another virtual host symlink named `www.myapp` for the application you want.

### Status API

Puma-dev is starting to evolve a status API that can be used to introspect it and the apps. To access it, send a request with the `Host: puma-dev` and the path `/status`, for example: `curl -H "Host: puma-dev" localhost/status`.

The status includes:
  * If it is booting, running, or dead
  * The directory of the app
  * The last 1024 lines the app output

## Development

To build puma-dev, follow these steps:

* Install [golang](http://golang.org)
* Run `go get github.com/puma/puma-dev/...`
* Run `go get github.com/vektra/errors/...`
* Run `$GOPATH/bin/puma-dev` to use your new binary

Puma-dev uses [govendor](https://github.com/kardianos/govendor) to manage dependencies, so if you're working on puma-dev and need to introduce a new dependency, run `govendor fetch +vendor <package path>` to pull it into `vendor`. Then you can use it from within `puma-dev/src`

### Releasing & Packaging for Homebrew

This script compiles Darwin and Linux binaries, creates a release tag, and uploads the binaries to the release. The Darwin binary can then be referenced in [puma/homebrew-puma](https://github.com/puma/homebrew-puma/blob/bd977276dace11f9b31e6181aa770d676a996b11/puma-dev.rb#L4).

```shell
# Gox is a "A dead simple, no frills Go cross compile tool." It's used inside `make release`.
go get -u github.com/mitchellh/gox
# Ghr can be used to "Upload multiple artifacts to GitHub Release in parallel."
go get -u github.com/tcnksm/ghr

export OWNER="puma"
export REPO="puma-dev"
export RELEASE="0.13"
export GITHUB_TOKEN="$GITHUB_API_TOKEN"

make release

git tag -f "v${RELEASE}"
git push origin "v${RELEASE}"

ghr -u $OWNER  -t $GITHUB_TOKEN -r $REPO  -n "v${RELEASE}" -delete -prerelease "v${RELEASE}" ./pkg/
```
