//go:build darwin

package applehv

import (
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// TODO the following functions were taken from pkg/qemu/claim_darwin.go and
// should be refactored.  I'm thinking even something in pkg/machine/

func dockerClaimSupported() bool {
	return true
}

func dockerClaimHelperInstalled() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}

	labelName := fmt.Sprintf("com.github.containers.podman.helper-%s", u.Username)
	fileName := filepath.Join("/Library", "LaunchDaemons", labelName+".plist")
	info, err := os.Stat(fileName)
	return err == nil && info.Mode().IsRegular()
}

func claimDockerSock() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}

	helperSock := fmt.Sprintf("/var/run/podman-helper-%s.socket", u.Username)
	con, err := net.DialTimeout("unix", helperSock, time.Second*5)
	if err != nil {
		return false
	}
	_ = con.SetWriteDeadline(time.Now().Add(time.Second * 5))
	_, err = fmt.Fprintln(con, "GO")
	if err != nil {
		return false
	}
	_ = con.SetReadDeadline(time.Now().Add(time.Second * 5))
	read, err := io.ReadAll(con)

	return err == nil && string(read) == "OK"
}

func findClaimHelper() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}

	return filepath.Join(filepath.Dir(exe), "podman-mac-helper")
}

func checkSockInUse(sock string) bool {
	if info, err := os.Stat(sock); err == nil && info.Mode()&fs.ModeSocket == fs.ModeSocket {
		_, err = net.DialTimeout("unix", dockerSock, dockerConnectTimeout)
		return err == nil
	}

	return false
}

func alreadyLinked(target string, link string) bool {
	read, err := os.Readlink(link)
	return err == nil && read == target
}
