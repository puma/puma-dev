# Puma-dev: A development server for OS X

Puma-dev is the emotional successor to pow. It provides a quick and easy way to manage apps in development on OS X.

## Highlights

* Easy startup and idle shutdown of rack/rails apps
* Easy access to the apps using the `.dev` subdomain **(configurable)**

### Why not just use pow?

Pow doesn't support rack.hijack and thus not websockets and thus not actioncable. So for all those new Rails 5 apps, pow is a no-go. Puma-dev fills that hole.

## Install

* Via Homebrew is the easiest: `brew install puma/puma/puma-dev`
* Or download the latest release from https://github.com/puma/puma-dev/releases
* If you haven't run puma-dev before, run: `sudo puma-dev -setup` to configure some DNS settings that have to be done as root
* Run `puma-dev -install` to configure puma-dev to run in the background on ports 80 and 443 with the domain `.dev`.
  * If you're currently using pow, puma-dev taking control of `.dev` will break it. If you want to just try out puma-dev and leave pow working, pass `-d pdev` on `-install` to use `.pdev` instead.

### Options

Run: `puma-dev -h`

You have the ability to configure most of the values that you'll use day-to-day.

### Setup

Run: `sudo puma-dev -setup`.

This configures the bits that require root access, which allows your user access to the `/etc/resolver` directory.

### Coming from v0.2

Puma-dev v0.3 and later use launchd to access privileged ports, so if you installed v0.2, you'll need to remove the firewall rules.

Run: `sudo puma-dev -cleanup`

### Background Install/Upgrading for port 80 access

If you want puma-dev to run in the background while you're logged in and on a common port, then you'll need to install it.

*NOTE:* If you installed puma-dev v0.2, please run `sudo puma-dev -cleanup` to remove firewall rules that puma-dev no longer uses (and will conflict with puma-dev working)

Run `puma-dev -install`.

If you wish to have `puma-dev` use a port other than 80, pass it via the `-install-port`, for example to use port 81: `puma-dev -install -install-port 81`.

### Running in the foreground

Run: `puma-dev`

Puma-dev will startup by default using the directory `~/.puma-dev`, looking for symlinks to apps just like pow. Drop a symlink to your app in there as: `cd ~/.puma-dev; ln -s /path/to/my/app test`. You can now access your app as `test.dev`.

Running `puma-dev` in this way will require you to use the listed http port, which is `9280` by default.

### Coming from Pow

By default, puma-dev uses the domain `.dev` to manage your apps, so that it doesn't interfere with a pow installation. If you want to have puma-dev take over for pow entirely, just run `puma-dev -pow`. Puma-dev will now use the `.dev` domain and look for apps in `~/.pow`.

### Configuration

Puma-dev supports loading environment variables before puma starts. It checks for the following files in this order:

* `.env`
* `.powrc`
* `.powenv`

Additionally, puma-dev uses a few environment variables to control how puma is started that you can overwrite in your loaded shell config.

* `CONFIG`: A puma configuration file to load, usually something like `config/puma-dev.rb`. Defaults to no config.
* `THREADS`: How many threads puma should use concurrently. Defaults to 5.
* `WORKERS`: How many worker processes to start. Defaults to 0, meaning only use threads.

### Purging

If you would like to have puma-dev stop all the apps (for resource issues or because an app isn't restarting properly), you can send `puma-dev` the signal `USR1`. The easiest way to do that is:

`pkill -USR1 puma-dev`

### Uninstall

Run: `puma-dev -uninstall`

## App usage

Simply symlink your apps directory into `~/.puma-dev`! That's it!

### Proxy support

Puma-dev can also proxy requests from a nice dev domain to another app. To do so, just write a file (rather than a symlink'd directory) into `~/.puma-dev` with the connection information.

For example, to have port 9292 show up as `awesome.dev`: `echo 9292 > ~/.puma-dev/awesome`.

Or to proxy to another host: `echo 10.3.1.2:9292 > ~/.puma-dev/awesome-elsewhere`.

### HTTPS

Puma-dev automatically makes the apps available via SSL as well. When you first run puma-dev, it will have likely caused a dialog to appear to put in your password. What happened there was puma-dev generates it's own CA certification that is stored in `~/Library/Application Support/io.puma.dev/cert.pem`.

That CA cert is used to dynamically create certificates for your apps when access to them is requested. It automatically happens, no configuration necessary. The certs are stored entirely in memory so future restarts of puma-dev simply generate new ones.

When `-install` is used (and let's be honest, that's how you want to use puma-dev), then it listens on port 443 by default (configurable with `-install-https-port`) so you can just do `https://blah.dev` to access your app via https.

## Development

To build puma-dev, follow these steps:

* Install golang (http://golang.org)
* Install gb (http://getgb.io)
* Run `make`
* Run `bin/puma-dev` to use your new binary

Puma-dev uses gb (http://getgb.io) to manage dependencies, so if you're working on puma-dev and need to introduce a new dependency, run `gb vendor fetch <package path>` to pull it into `vendor/src`. Then you can use it from within `puma-dev/src`

