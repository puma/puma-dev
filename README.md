# Puma-dev: A fast, zero-config development server for OS X and Linux

Puma-dev is the emotional successor to pow. It provides a quick and easy way to manage apps in development on OS X and Linux.

## Highlights

* Easy startup and idle shutdown of rack/rails apps
* Easy access to the apps using the `.dev` subdomain **(configurable)**

### Why not just use pow?

Pow doesn't support rack.hijack and thus not websockets and thus not actioncable. So for all those new Rails 5 apps, pow is a no-go. Puma-dev fills that hole. Puma-dev also goes one step further and provides zero-config https support to your development environment, as well as offering Linux support.

## Install

* Via Homebrew is the easiest: `brew install puma/puma/puma-dev`
* Or download the latest release from https://github.com/puma/puma-dev/releases
* If you haven't run puma-dev before, run: `sudo puma-dev -setup` to configure some DNS settings that have to be done as root
* Run `puma-dev -install` to configure puma-dev to run in the background on ports 80 and 443 with the domain `.dev`.
  * If you're currently using pow, puma-dev taking control of `.dev` will break it. If you want to just try out puma-dev and leave pow working, pass `-d pdev` on `-install` to use `.pdev` instead.

*NOTE:* if you had pow installed before in the system, please make sure to run
pow's uninstall script. Read more details in [the pow manual](http://pow.cx/manual.html#section_1.2).

### Install on Linux

* Puma-dev supports linux but requires additional installation to make all the features work.
* You can either build from source or download a binary from https://github.com/puma/puma-dev/releases

#### .dev domain

Install the [dev-tld-resolver](https://github.com/puma/dev-tld-resolver) to make domains resolve.

#### Port 80/443 binding

There are 2 options to allow puma-dev to listen on port 80 and 443.

1. `sudo setcap CAP\_NET\_BIND\_SERVICE=+eip /path/to/puma-dev`
2. Use `authbind`.

You don't need to bind to port 80/443 to use puma-dev but obviously it makes using the `.dev` domain much nicer.

There is a shortcut for binding to 80/443 by passing `-sysbind` which overrides `-http-port` and `-https-port`.

### Options

Run: `puma-dev -h`

You have the ability to configure most of the values that you'll use day-to-day.

### Setup (OS X only)

Run: `sudo puma-dev -setup`.

This configures the bits that require root access, which allows your user access to the `/etc/resolver` directory.

### Coming from v0.2

Puma-dev v0.3 and later use launchd to access privileged ports, so if you installed v0.2, you'll need to remove the firewall rules.

Run: `sudo puma-dev -cleanup`

### Background Install/Upgrading for port 80 access (OS X only)

If you want puma-dev to run in the background while you're logged in and on a common port, then you'll need to install it.

*NOTE:* If you installed puma-dev v0.2, please run `sudo puma-dev -cleanup` to remove firewall rules that puma-dev no longer uses (and will conflict with puma-dev working)

Run `puma-dev -install`.

If you wish to have `puma-dev` use a port other than 80, pass it via the `-install-port`, for example to use port 81: `puma-dev -install -install-port 81`.

### Running in the foreground

Run: `puma-dev`

Puma-dev will startup by default using the directory `~/.puma-dev`, looking for symlinks to apps just like pow. Drop a symlink to your app in there as: `cd ~/.puma-dev; ln -s /path/to/my/app test`. You can now access your app as `test.dev`.

Running `puma-dev` in this way will require you to use the listed http port, which is `9280` by default.

### Coming from Pow

By default, puma-dev uses the domain `.dev` to manage your apps. If you want to have puma-dev look for apps in `~/.pow`, just run `puma-dev -pow`.

## Configuration

Puma-dev supports loading environment variables before puma starts. It checks for the following files in this order:

* `.env`
* `.powrc`
* `.powenv`

Additionally, puma-dev uses a few environment variables to control how puma is started that you can overwrite in your loaded shell config.

* `CONFIG`: A puma configuration file to load, usually something like `config/puma-dev.rb`. Defaults to no config.
* `THREADS`: How many threads puma should use concurrently. Defaults to 5.
* `WORKERS`: How many worker processes to start. Defaults to 0, meaning only use threads.

## Purging

If you would like to have puma-dev stop all the apps (for resource issues or because an app isn't restarting properly), you can send `puma-dev` the signal `USR1`. The easiest way to do that is:

`puma-dev -stop`

### Uninstall (OS X only)

Run: `puma-dev -uninstall`

## App usage

Simply symlink your apps directory into `~/.puma-dev`! That's it!

### Proxy support

Puma-dev can also proxy requests from a nice dev domain to another app. To do so, just write a file (rather than a symlink'd directory) into `~/.puma-dev` with the connection information.

For example, to have port 9292 show up as `awesome.dev`: `echo 9292 > ~/.puma-dev/awesome`.

Or to proxy to another host: `echo 10.3.1.2:9292 > ~/.puma-dev/awesome-elsewhere`.

### HTTPS

Puma-dev automatically makes the apps available via SSL as well. When you first run puma-dev, it will have likely caused a dialog to appear to put in your password. What happened there was puma-dev generates its own CA certification that is stored in `~/Library/Application Support/io.puma.dev/cert.pem`.

That CA cert is used to dynamically create certificates for your apps when access to them is requested. It automatically happens, no configuration necessary. The certs are stored entirely in memory so future restarts of puma-dev simply generate new ones.

When `-install` is used (and let's be honest, that's how you want to use puma-dev), then it listens on port 443 by default (configurable with `-install-https-port`) so you can just do `https://blah.dev` to access your app via https.

### OS X Logging

When puma-dev is installed as a user agent (the default mode), it will log output from itself and the apps to `~/Library/Logs/puma-dev.log`. You can refer to there to find out if apps have started and look for errors.

In the future, puma-dev will provide an integrated console for this log output.

### Websockets

Puma-dev supports websockets natively but you may need to tell your web framework to allow the connections.

In the case of rails, you need to configure rails to allow all websockets or websocket requests from certain domains. The quickest way is to add `config.action_cable.disable_request_forgery_protection = true` to `config/environments/development.rb`. This will allow all websocket connections while in development.

*Do not use disable_request_forgery_protection in production!*

Or you can add something like `config.action_cable.allowed_request_origins = /(\.dev$)|^localhost$/` to allow anything under `.dev` as well as `localhost`.

### xip.io

Puma-dev supports `xip.io` domains. It will detect them and strip them away, so that your `test` app can be accessed as `test.A.B.C.D.xip.io`.

### Static file support

Like pow, puma-dev support serving static files. If an app has a `public` directory, then any urls that match files within that directory are served. The static files have priority over the app.

### Status API

Puma-dev is starting to evolve a status API that can be used to introspect it and the apps. To access it, send a request with the `Host: puma-dev` and the path `/status`, for example: `curl -H "Host: puma-dev" localhost/status`.

The status includes:
  * If it is booting, running, or dead
  * The directory of the app
  * The last 1024 lines the app output

## Subcommands

### `puma-dev link [-n name] [dir]`

Creates links to app directories into your puma-dev directory (`~/.puma-dev` by default).

## Development

To build puma-dev, follow these steps:

* Install [golang](http://golang.org)
* Run `go get github.com/puma/puma-dev/...`
* Run `$GOPATH/bin/puma-dev` to use your new binary

Puma-dev uses [govendor](https://github.com/kardianos/govendor) to manage dependencies, so if you're working on puma-dev and need to introduce a new dependency, run `govendor fetch +vendor <package path>` to pull it into `vendor`. Then you can use it from within `puma-dev/src`
