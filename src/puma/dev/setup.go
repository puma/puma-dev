package dev

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kardianos/osext"
	"github.com/mitchellh/go-homedir"
)

func Setup(skipFirewall bool) error {
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

	if !skipFirewall {
		fmt.Printf("* Configuring firewall...\n")

		cmd := exec.Command("pfctl", "-a", "com.apple/250.PumaDevFirewall", "-E", "-f", "-")
		rule := fmt.Sprintf(
			"rdr pass inet proto tcp from any to any port = %d -> 127.0.0.1 port %d\n",
			80, 9280)

		cmd.Stdin = strings.NewReader(rule)

		err = cmd.Run()
		if err != nil {
			return err
		}

		cur, err := exec.Command(
			"pfctl", "-a", "com.apple/250.PumaDevFirewall", "-s", "nat", "-q").Output()
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(cur)) != strings.TrimSpace(rule) {
			return fmt.Errorf("Unable to verify firewall installation was successful")
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

func InstallIntoSystem(skip80 bool) error {
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
     <string>exec '%s'</string>
   </array>
   <key>KeepAlive</key>
   <true/>
   <key>RunAtLoad</key>
   <true/>
   </dict>
</plist>
`

	plist := mustExpand("~/Library/LaunchAgents/io.puma.dev.plist")

	err = ioutil.WriteFile(
		plist,
		[]byte(fmt.Sprintf(userTemplate, binPath)),
		0644,
	)

	if err != nil {
		return err
	}

	err = exec.Command("launchctl", "load", plist).Run()
	if err != nil {
		return err
	}

	fmt.Printf("* Installed puma-dev as LaunchAgent\n")

	var sysTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
   <key>Label</key>
   <string>io.puma.devsetup</string>
   <key>ProgramArguments</key>
   <array>
     <string>zsh</string>
     <string>-l</string>
     <string>-c</string>
     <string>exec '%s' %s</string>
   </array>
   <key>RunAtLoad</key>
   <true/>
	 <key>UserName</key>
	 <string>root</string>
   </dict>
</plist>
`
	opts := "-setup"
	if skip80 {
		opts += " -setup-skip-80"
	}

	plist = "/Library/LaunchDaemons/io.puma.devsetup.plist"

	err = ioutil.WriteFile(
		plist,
		[]byte(fmt.Sprintf(sysTemplate, binPath, opts)),
		0644,
	)

	if err != nil {
		return err
	}

	err = exec.Command("launchctl", "load", plist).Run()
	if err != nil {
		return err
	}

	fmt.Printf("* Installed setup LaunchDaemon\n")

	return nil
}
