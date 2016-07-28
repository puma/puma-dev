package dev

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
	"gopkg.in/tomb.v2"
)

var ErrUnexpectedExit = errors.New("unexpected exit")

type App struct {
	Name    string
	Port    int
	Command *exec.Cmd

	dir string

	t tomb.Tomb

	listener net.Listener

	stdout  io.Reader
	lock    sync.Mutex
	pool    *AppPool
	lastUse time.Time
}

func (a *App) Address() string {
	return fmt.Sprintf("localhost:%d", a.Port)
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
	a.listener.Close()

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

func LaunchApp(pool *AppPool, name, dir string) (*App, error) {
	// Create a listener socket and inject it
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	addr := l.Addr().(*net.TCPAddr)

	shell := os.Getenv("SHELL")

	cmd := exec.Command(shell, "-l", "-i", "-c",
		fmt.Sprintf("exec bundle exec puma -C- --tag puma-dev:%s -b tcp://127.0.0.1:%d",
			name, addr.Port))

	cmd.Dir = dir

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("PUMA_INHERIT_0=3:tcp://127.0.0.1:%d", addr.Port))

	tcpListener := l.(*net.TCPListener)
	socket, err := tcpListener.File()
	if err != nil {
		return nil, err
	}

	cmd.ExtraFiles = []*os.File{socket}

	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	fmt.Printf("! Booted app '%s' on port %d\n", name, addr.Port)

	app := &App{
		Name:     name,
		Port:     addr.Port,
		Command:  cmd,
		listener: l,
		stdout:   stdout,
		dir:      dir,
	}

	app.t.Go(app.watch)
	app.t.Go(app.idleMonitor)
	app.t.Go(app.restartMonitor)

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

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Unknown app: %s", name)
	}

	app, err = LaunchApp(a, name, path)
	if err != nil {
		return nil, err
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
	defer a.lock.Unlock()

	for _, app := range a.apps {
		app.t.Kill(nil)
	}

	for _, app := range a.apps {
		app.t.Wait()
	}
}
