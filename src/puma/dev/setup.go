package dev

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/kardianos/osext"
	"github.com/mitchellh/go-homedir"
)

func Setup() error {
	err := os.MkdirAll(etcDir, 0755)
	if err != nil {
		return err
	}

	var ok bool

	sudo := os.Getenv("SUDO_USER")
	if sudo != "" {
		uid, err1 := strconv.Atoi(os.Getenv("SUDO_UID"))
		gid, err2 := strconv.Atoi(os.Getenv("SUDO_GID"))

		if err1 == nil && err2 == nil {
			fmt.Printf("* Configuring %s to be owned by %s\n", etcDir, sudo)

			err := os.Chown(etcDir, uid, gid)
			if err != nil {
				return err
			}

			err = os.Chmod(etcDir, 0755)
			if err != nil {
				return err
			}

			ok = true
		}
	}

	if !ok {
		fmt.Printf("* Configuring %s to be world writable\n")
		err := os.Chmod(etcDir, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func mustExpand(str string) string {
	str, err := homedir.Expand(str)
	if err != nil {
		panic(err)
	}

	return str
}

func Cleanup() {
	oldSetup := "/Library/LaunchDaemons/io.puma.devsetup.plist"

	exec.Command("launchctl", "unload", oldSetup).Run()
	os.Remove(oldSetup)
	exec.Command("pfctl", "-F", "nat", "-a", "com.apple/250.PumaDevFirewall").Run()

	fmt.Printf("* Expunged old puma dev system rules\n")

	// Fix perms of the LaunchAgent
	uid, err1 := strconv.Atoi(os.Getenv("SUDO_UID"))
	gid, err2 := strconv.Atoi(os.Getenv("SUDO_GID"))

	if err1 == nil && err2 == nil {
		plist := mustExpand("~/Library/LaunchAgents/io.puma.dev.plist")
		os.Chown(plist, uid, gid)

		fmt.Printf("* Fixed permissions of user LaunchAgent\n")
	}
}

func InstallIntoSystem(listenPort int) error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}

	err = os.MkdirAll(mustExpand("~/bin"), 0755)
	if err != nil {
		return err
	}

	fmt.Printf("* Copying %s to ~/bin/puma-dev...\n", path)

	binPath := mustExpand("~/bin/puma-dev")

	err = exec.Command("cp", path, binPath).Run()
	if err != nil {
		return err
	}

	var userTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
   <key>Label</key>
   <string>io.puma.dev</string>
   <key>ProgramArguments</key>
   <array>
     <string>zsh</string>
     <string>-l</string>
     <string>-c</string>
     <string>exec '%s' -launchd</string>
   </array>
   <key>KeepAlive</key>
   <true/>
   <key>RunAtLoad</key>
   <true/>
   <key>Sockets</key>
   <dict>
       <key>Socket</key>
       <dict>
           <key>SockNodeName</key>
           <string>0.0.0.0</string>
           <key>SockServiceName</key>
           <string>%d</string>
       </dict>
   </dict>
   <key>StandardOutPath</key>
   <string>%s</string>
   <key>StandardErrorPath</key>
   <string>%s</string>
</dict>
</plist>
`

	logPath := mustExpand("~/Library/Logs/puma-dev.log")

	plist := mustExpand("~/Library/LaunchAgents/io.puma.dev.plist")

	err = ioutil.WriteFile(
		plist,
		[]byte(fmt.Sprintf(userTemplate, binPath, listenPort, logPath, logPath)),
		0644,
	)

	if err != nil {
		return err
	}

	// Unload a previous one if need be.
	exec.Command("launchctl", "unload", plist).Run()

	err = exec.Command("launchctl", "load", plist).Run()
	if err != nil {
		return err
	}

	fmt.Printf("* Installed puma-dev on port %d\n", listenPort)

	return nil
}

func Uninstall() {
	plist := mustExpand("~/Library/LaunchAgents/io.puma.dev.plist")

	// Unload a previous one if need be.
	exec.Command("launchctl", "unload", plist).Run()

	os.Remove(plist)

	fmt.Printf("* Removed puma-dev from automatically running\n")
}
