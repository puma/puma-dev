package dev

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/puma/puma-dev/linebuffer"
	"github.com/puma/puma-dev/watch"
	"github.com/vektra/errors"
	"gopkg.in/tomb.v2"
)

const DefaultThreads = 5

var ErrUnexpectedExit = errors.New("unexpected exit")

type App struct {
	Name    string
	Scheme  string
	Host    string
	Port    int
	Command *exec.Cmd
	Public  bool
	Events  *Events

	lines       linebuffer.LineBuffer
	lastLogLine string

	address string
	dir     string

	t tomb.Tomb

	stdout  io.Reader
	pool    *AppPool
	lastUse time.Time

	lock sync.Mutex

	booting bool

	readyChan chan struct{}
}

func (a *App) eventAdd(name string, args ...interface{}) {
	args = append([]interface{}{"app", a.Name}, args...)

	str := a.Events.Add(name, args...)
	a.lines.Append("#event " + str)
}

func (a *App) SetAddress(scheme, host string, port int) {
	a.Scheme = scheme
	a.Host = host
	a.Port = port

	if a.Port == 0 {
		a.address = host
	} else {
		a.address = fmt.Sprintf("%s:%d", a.Host, a.Port)
	}
}

func (a *App) Address() string {
	if a.Port == 0 {
		return a.Host
	}

	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *App) Kill(reason string) error {
	a.eventAdd("killing_app",
		"pid", a.Command.Process.Pid,
		"reason", reason,
	)

	fmt.Printf("! Killing '%s' (%d)\n", a.Name, a.Command.Process.Pid)
	err := a.Command.Process.Signal(syscall.SIGTERM)
	if err != nil {
		a.eventAdd("killing_error",
			"pid", a.Command.Process.Pid,
			"error", err.Error(),
		)
		fmt.Printf("! Error trying to kill %s: %s", a.Name, err)
	}
	return err
}

func (a *App) watch() error {
	c := make(chan error)

	go func() {
		r := bufio.NewReader(a.stdout)

		for {
			line, err := r.ReadString('\n')
			if line != "" {
				a.lines.Append(line)
				a.lastLogLine = line
				fmt.Fprintf(os.Stdout, "%s[%d]: %s", a.Name, a.Command.Process.Pid, line)
			}

			if err != nil {
				c <- err
				return
			}
		}
	}()

	var err error

	reason := "detected interval shutdown"

	select {
	case err = <-c:
		reason = "stdout/stderr closed"
		err = fmt.Errorf("%s:\n\t%s", ErrUnexpectedExit, a.lastLogLine)
	case <-a.t.Dying():
		err = nil
	}

	a.Kill(reason)
	a.Command.Wait()
	a.pool.remove(a)

	if a.Scheme == "httpu" {
		os.Remove(a.Address())
	}

	a.eventAdd("shutdown")

	fmt.Printf("* App '%s' shutdown and cleaned up\n", a.Name)

	return err
}

func (a *App) idleMonitor() error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.pool.maybeIdle(a) {
				a.Kill("app is idle")
				return nil
			}
		case <-a.t.Dying():
			return nil
		}
	}

	return nil
}

func (a *App) restartMonitor() error {
	tmpDir := filepath.Join(a.dir, "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return err
	}

	restart := filepath.Join(tmpDir, "restart.txt")

	f, err := os.Create(restart)
	if err != nil {
		return err
	}
	f.Close()

	return watch.Watch(restart, a.t.Dying(), func() {
		a.Kill("restart.txt touched")
	})
}

func (a *App) WaitTilReady() error {
	select {
	case <-a.readyChan:
		// double check we aren't also dying
		select {
		case <-a.t.Dying():
			return a.t.Err()
		default:
			a.lastUse = time.Now()
			return nil
		}
	case <-a.t.Dying():
		return a.t.Err()
	}
}

const (
	Booting = iota
	Running
	Dead
)

func (a *App) Status() int {
	// These are done in order as separate selects because go's
	// select does not execute case's sequentially, it runs bodies
	// after sampling all channels and picking a random body.
	select {
	case <-a.t.Dying():
		return Dead
	default:
		select {
		case <-a.readyChan:
			return Running
		default:
			return Dead
		}
	}
}

func (a *App) Log() string {
	var buf bytes.Buffer
	a.lines.WriteTo(&buf)
	return buf.String()
}

const executionShell = `exec bash -c '
cd %s

if test -e ~/.powconfig; then
	source ~/.powconfig
fi

if test -e .env; then
	source .env
fi

if test -e .powrc; then
	source .powrc
fi

if test -e .powenv; then
	source .powenv
fi

if test -e Gemfile && bundle exec puma -V &>/dev/null; then
	exec bundle exec puma -C $CONFIG --tag puma-dev:%s -w $WORKERS -t 0:$THREADS -b unix:%s
fi

exec puma -C $CONFIG --tag puma-dev:%s -w $WORKERS -t 0:$THREADS -b unix:%s'
`

func (pool *AppPool) LaunchApp(name, dir string) (*App, error) {
	tmpDir := filepath.Join(dir, "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return nil, err
	}

	socket := filepath.Join(tmpDir, fmt.Sprintf("puma-dev-%d.sock", os.Getpid()))

	shell := os.Getenv("SHELL")

	if shell == "" {
		fmt.Printf("! SHELL env var not set, using /bin/bash by default")
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell, "-l", "-i", "-c",
		fmt.Sprintf(executionShell, dir, name, socket, name, socket))

	cmd.Dir = dir

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("THREADS=%d", DefaultThreads),
		"WORKERS=0",
		"CONFIG=-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Stderr = cmd.Stdout

	err = cmd.Start()
	if err != nil {
		return nil, errors.Context(err, "starting app")
	}

	fmt.Printf("! Booting app '%s' on socket %s\n", name, socket)

	app := &App{
		Name:      name,
		Command:   cmd,
		Events:    pool.Events,
		stdout:    stdout,
		dir:       dir,
		pool:      pool,
		readyChan: make(chan struct{}),
		lastUse:   time.Now(),
	}

	app.eventAdd("booting_app", "socket", socket)

	stat, err := os.Stat(filepath.Join(dir, "public"))
	if err == nil {
		app.Public = stat.IsDir()
	}

	app.SetAddress("httpu", socket, 0)

	app.t.Go(app.watch)
	app.t.Go(app.idleMonitor)
	app.t.Go(app.restartMonitor)

	app.t.Go(func() error {
		// This is a poor substitute for getting an actual readiness signal
		// from puma but it's good enough.

		app.eventAdd("waiting_on_app")

		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-app.t.Dying():
				app.eventAdd("dying_on_start")
				fmt.Printf("! Detecting app '%s' dying on start\n", name)
				return fmt.Errorf("app died before booting")
			case <-ticker.C:
				c, err := net.Dial("unix", socket)
				if err == nil {
					c.Close()
					app.eventAdd("app_ready")
					fmt.Printf("! App '%s' booted\n", name)
					close(app.readyChan)
					return nil
				}
			}
		}
	})

	return app, nil
}

func (pool *AppPool) readProxy(name, path string) (*App, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	app := &App{
		Name:      name,
		Events:    pool.Events,
		pool:      pool,
		readyChan: make(chan struct{}),
		lastUse:   time.Now(),
	}

	data = bytes.TrimSpace(data)

	port, err := strconv.Atoi(string(data))
	if err == nil {
		app.SetAddress("http", "127.0.0.1", port)
	} else {
		u, err := url.Parse(string(data))
		if err != nil {
			return nil, err
		}

		var (
			sport, host string
			port        int
		)

		host, sport, err = net.SplitHostPort(u.Host)
		if err == nil {
			port, err = strconv.Atoi(sport)
			if err != nil {
				return nil, err
			}
		} else {
			host = u.Host
		}

		app.SetAddress(u.Scheme, host, port)
	}

	app.eventAdd("proxy_created",
		"destination", fmt.Sprintf("%s://%s", app.Scheme, app.Address()))

	fmt.Printf("* Generated proxy connection for '%s' to %s://%s\n",
		name, app.Scheme, app.Address())

	// to satisfy the tomb
	app.t.Go(func() error {
		<-app.t.Dying()
		return nil
	})

	close(app.readyChan)

	return app, nil
}

type AppPool struct {
	Dir      string
	IdleTime time.Duration
	Debug    bool
	Events   *Events

	AppClosed func(*App)

	lock sync.Mutex
	apps map[string]*App
}

func (a *AppPool) maybeIdle(app *App) bool {
	a.lock.Lock()
	defer a.lock.Unlock()

	diff := time.Since(app.lastUse)
	if diff > a.IdleTime {
		app.eventAdd("idle_app", "last_used", diff.String())
		delete(a.apps, app.Name)
		return true
	}

	return false
}

var ErrUnknownApp = errors.New("unknown app")

func (a *AppPool) App(name string) (*App, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.apps == nil {
		a.apps = make(map[string]*App)
	}

	app, ok := a.apps[name]
	if ok {
		return app, nil
	}

	path := filepath.Join(a.Dir, name)

	a.Events.Add("app_lookup", "path", path)

	stat, err := os.Stat(path)
	destPath, _ := os.Readlink(path)

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Check there might be a link there but it's not valid
		_, err := os.Lstat(path)
		if err == nil {
			fmt.Printf("! Bad symlink detected '%s'. Destination '%s' doesn't exist\n", path, destPath)
			a.Events.Add("bad_symlink", "path", path, "dest", destPath)
		}

		// If possible, also try expanding - to / to allow for apps in subdirs
		possible := strings.Replace(name, "-", "/", -1)
		if possible == name {
			return nil, ErrUnknownApp
		}

		path = filepath.Join(a.Dir, possible)

		a.Events.Add("app_lookup", "path", path)

		stat, err = os.Stat(path)
		destPath, _ = os.Readlink(path)

		if err != nil {
			if !os.IsNotExist(err) && err.Error() != "not a directory" {
				return nil, err
			}

			// Check there might be a link there but it's not valid
			_, err := os.Lstat(path)
			if err == nil {
				fmt.Printf("! Bad symlink detected '%s'. Destination '%s' doesn't exist\n", path, destPath)
				a.Events.Add("bad_symlink", "path", path, "dest", destPath)
			}

			return nil, ErrUnknownApp
		}
	}

	canonicalName := name
	aliasName := ""

	// Handle multiple symlinks to the same app
	destStat, err := os.Stat(destPath)
	if err == nil {
		destName := destStat.Name()
		if destName != canonicalName {
			canonicalName = destName
			aliasName = name
		}
	}

	app, ok = a.apps[canonicalName]

	if !ok {
		if stat.IsDir() {
			app, err = a.LaunchApp(canonicalName, path)
		} else {
			app, err = a.readProxy(canonicalName, path)
		}
	}

	if err != nil {
		a.Events.Add("error_starting_app", "app", canonicalName, "error", err.Error())
		return nil, err
	}

	a.apps[canonicalName] = app

	if aliasName != "" {
		a.apps[aliasName] = app
	}

	return app, nil
}

func (a *AppPool) remove(app *App) {
	a.lock.Lock()
	defer a.lock.Unlock()

	// Find all instance references so aliases are removed too
	for name, candidate := range a.apps {
		if candidate == app {
			delete(a.apps, name)
		}
	}

	if a.AppClosed != nil {
		a.AppClosed(app)
	}
}

func (a *AppPool) ForApps(f func(*App)) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, app := range a.apps {
		f(app)
	}
}

func (a *AppPool) Purge() {
	a.lock.Lock()

	var apps []*App

	for _, app := range a.apps {
		apps = append(apps, app)
	}

	a.lock.Unlock()

	for _, app := range apps {
		app.eventAdd("purging_app")
		app.t.Kill(nil)
	}

	for _, app := range apps {
		app.t.Wait()
	}

	a.Events.Add("apps_purged")
}
