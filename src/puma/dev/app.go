package dev

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
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

	address string
	dir     string

	t tomb.Tomb

	stdout  io.Reader
	lock    sync.Mutex
	pool    *AppPool
	lastUse time.Time
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

func (a *App) Kill() error {
	fmt.Printf("! Killing '%s' (%d)\n", a.Name, a.Command.Process.Pid)
	err := a.Command.Process.Kill()
	if err != nil {
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
				fmt.Fprintf(os.Stdout, "%s[%d]: %s", a.Name, a.Command.Process.Pid, line)
			}

			if err != nil {
				c <- err
				return
			}
		}
	}()

	var err error

	select {
	case err = <-c:
		err = ErrUnexpectedExit
	case <-a.t.Dying():
		a.Kill()
		err = nil
	}

	a.Command.Wait()
	a.pool.remove(a)

	if a.Scheme == "httpu" {
		os.Remove(a.Address())
	}

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
				a.Kill()
			}
			return nil
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

	dev, err := fsevents.DeviceForPath(restart)
	if err != nil {
		return err
	}

	es := &fsevents.EventStream{
		Paths:   []string{restart},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.IgnoreSelf,
	}

	es.Start()

	defer es.Stop()

	for {
		select {
		case events := <-es.Events:
			for _, ev := range events {
				if ev.Flags&fsevents.ItemInodeMetaMod != 0 {
					a.Kill()
				}
			}
		case <-a.t.Dying():
			return nil
		}
	}
}

func (a *App) UpdateUsed() {
	a.lastUse = time.Now()
}

const executionShell = `# puma-dev generated at runtime
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

if test -e Gemfile; then
	exec bundle exec puma -C $CONFIG --tag puma-dev:%s -w $WORKERS -t 0:$THREADS -b unix:%s
fi


exec puma -C $CONFIG --tag puma-dev:%s -w $WORKERS -t 0:$THREADS -b unix:%s
`

func LaunchApp(pool *AppPool, name, dir string) (*App, error) {
	tmpDir := filepath.Join(dir, "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return nil, err
	}

	socket := filepath.Join(tmpDir, fmt.Sprintf("puma-dev-%d.sock", os.Getpid()))

	shell := os.Getenv("SHELL")

	cmd := exec.Command(shell, "-l", "-i", "-c",
		fmt.Sprintf(executionShell, name, socket, name, socket))

	cmd.Dir = dir

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("THREADS=%d", DefaultThreads),
		"WORKERS=0",
		"CONFIG=-",
	)

	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	fmt.Printf("! Booted app '%s' on socket %s\n", name, socket)

	app := &App{
		Name:    name,
		Command: cmd,
		stdout:  stdout,
		dir:     dir,
		pool:    pool,
	}

	app.SetAddress("httpu", socket, 0)

	app.t.Go(app.watch)
	app.t.Go(app.idleMonitor)
	app.t.Go(app.restartMonitor)

	// This is a poor substitute for getting an actual readiness signal
	// from puma but it's good enough.
	for {
		c, err := net.Dial("unix", socket)
		if err == nil {
			c.Close()
			break
		}

		time.Sleep(250 * time.Microsecond)
	}

	return app, nil
}

type AppPool struct {
	Dir      string
	IdleTime time.Duration

	lock sync.Mutex
	apps map[string]*App
}

func (a *AppPool) maybeIdle(app *App) bool {
	a.lock.Lock()
	defer a.lock.Unlock()

	diff := time.Since(app.lastUse)
	if diff > a.IdleTime {
		delete(a.apps, app.Name)
		return true
	}

	return false
}

func (a *AppPool) App(name string) (*App, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.apps == nil {
		a.apps = make(map[string]*App)
	}

	app, ok := a.apps[name]
	if ok {
		app.UpdateUsed()
		return app, nil
	}

	path := filepath.Join(a.Dir, name)

	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Unknown app: %s", name)
	}

	if stat.IsDir() {
		app, err = LaunchApp(a, name, path)
		if err != nil {
			return nil, err
		}
	} else {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		app = &App{
			Name: name,
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

		fmt.Printf("* Generated proxy connection for '%s' to %s://%s\n",
			name, app.Scheme, app.Address())
	}

	app.pool = a

	app.UpdateUsed()
	a.apps[name] = app

	return app, nil
}

func (a *AppPool) remove(app *App) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.apps, app.Name)
}

func (a *AppPool) Purge() {
	a.lock.Lock()

	var apps []*App

	for _, app := range a.apps {
		apps = append(apps, app)
	}

	a.lock.Unlock()

	for _, app := range apps {
		app.t.Kill(nil)
	}

	for _, app := range apps {
		app.t.Wait()
	}
}
