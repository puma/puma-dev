# Puma-dev: A development server for OS X

Puma-dev is the emotional successor to pow. It provides a quick and easy way to manage apps in development on OS X.

## Highlights

* Easy startup and idle shutdown of rack/rails apps
* Easy access to the apps using the `.pdev` subdomain **(configurable)**


### Why not just use pow?

Pow doesn't support rack.hijack and thus not websockets and thus not actioncable. So for all those new Rails 5 apps, pow is a no-go. Puma-dev fills that hole.

### Options

Run: `puma-dev -h`

You have the ability to configure most of the values that you'll use day-to-day.

### Setup

Run: `sudo puma-dev -setup`.

This will configure the bits that require root access. If you're currently using pow and want to just try out puma-dev, I suggest using `sudo puma-dev -setup -setup-skip-80` to not install the port 80 redirect rule that will conflict with pow. You can still access apps, you'll just need to add port `9280` to your requests, such as `http://test.pdev:9280`.

### Quickstart

Run: `puma-dev`

Puma-dev will startup by default using the directory `~/.puma-dev`, looking for symlinks to apps just like pow. Drop a symlink to your app in there as: `cd ~/.puma-dev; ln -s test /path/to/my/app`. You can now access your app as `test.pdev`.

### Coming from Pow

By default, puma-dev uses the domain `.pdev` to manage your apps, so that it doesn't interfer with a pow installation. If you want to have puma-dev take over for pow entirely, just run `puma-dev -pow`. Puma-dev will now use the `.dev` domain and look for apps in `~/.pow`.

### Purging

If you would like to have puma-dev stop all the apps (for resource issues or because an app isn't restarting properly), you can send `puma-dev` the signal `USR1`. The easiest way to do that is:

`pkill -USR1 puma-dev`

## Development

To build puma-dev, follow these steps:

* Install golang (http://golang.org)
* Install gb (http://getgb.io)
* Run `make`
* Run `bin/puma-dev` to use your new binary

Puma-dev use gb (http://getgb.io) to manage dependencies, so if you're working on puma-dev and need to introduce a new dependency, run `gb vendor fetch <package path>` to pull it into `vendor/src`. Then you can use it from within `puma-dev/src`

