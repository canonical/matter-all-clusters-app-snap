package remote

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"chip-tool-snap-tests/matter"
	"github.com/canonical/matter-snap-testing/utils"
	"golang.org/x/crypto/ssh"
)

var (
	remoteUser           = ""
	remotePassword       = ""
	remoteHost           = ""
	remoteInfraInterface = matter.DefaultInfraInterfaceValue
	remoteRadioUrl       = matter.DefaultRadioUrl

	SSHClient *ssh.Client
)

func Setup(t *testing.T) {
	loadEnvVars()

	connectSSH(t)

	deployOTBRAgent(t)

	deployAllClustersApp(t)
}

func loadEnvVars() {

	if v := os.Getenv(matter.RemoteUserEnv); v != "" {
		remoteUser = v
	}

	if v := os.Getenv(matter.RemotePasswordEnv); v != "" {
		remotePassword = v
	}

	if v := os.Getenv(matter.RemoteHostEnv); v != "" {
		remoteHost = v
	}

	if v := os.Getenv(matter.RemoteInfraInterfaceEnv); v != "" {
		remoteInfraInterface = v
	}

	if v := os.Getenv(matter.RemoteRadioUrlEnv); v != "" {
		remoteRadioUrl = v
	}
}

func connectSSH(t *testing.T) {
	if SSHClient != nil {
		return
	}

	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(remotePassword),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var err error
	SSHClient, err = ssh.Dial("tcp", remoteHost+":22", config)
	if err != nil {
		t.Fatalf("Failed to dial: %s", err)
	}

	t.Cleanup(func() {
		SSHClient.Close()
	})

	t.Logf("SSH: connected to %s", remoteHost)
}

func deployOTBRAgent(t *testing.T) {
	start := time.Now().UTC()

	t.Cleanup(func() {
		dumpLogs(t, "openthread-border-router", start)
		exec(t, "sudo snap remove --purge openthread-border-router")
	})

	commands := []string{
		"sudo snap remove --purge openthread-border-router",
		"sudo snap install openthread-border-router --channel=latest/beta",
		fmt.Sprintf("sudo snap set openthread-border-router %s='%s'", matter.InfraInterfaceKey, remoteInfraInterface),
		fmt.Sprintf("sudo snap set openthread-border-router %s='%s'", matter.RadioUrlKey, remoteRadioUrl),
		// "sudo snap connect openthread-border-router:avahi-control",
		"sudo snap connect openthread-border-router:firewall-control",
		"sudo snap connect openthread-border-router:raw-usb",
		"sudo snap connect openthread-border-router:network-control",
		// "sudo snap connect openthread-border-router:bluetooth-control",
		// "sudo snap connect openthread-border-router:bluez",
		"sudo snap start openthread-border-router",
	}
	for _, cmd := range commands {
		exec(t, cmd)
	}

	WaitForLogMessage(t, matter.OtbrSnap, "Start Thread Border Agent: OK", start)
	t.Log("OTBR on remote device is ready")
}

func deployAllClustersApp(t *testing.T) {
	start := time.Now().UTC()

	t.Cleanup(func() {
		dumpLogs(t, "matter-all-clusters-app", start)
		exec(t, "sudo snap remove --purge matter-all-clusters-app")
	})

	commands := []string{
		// "sudo apt install -y bluez",
		"sudo snap remove --purge matter-all-clusters-app",
		"sudo snap install matter-all-clusters-app --channel=latest/beta",
		"sudo snap set matter-all-clusters-app args='--thread'",
		"sudo snap connect matter-all-clusters-app:avahi-control",
		// "sudo snap connect matter-all-clusters-app:bluez",
		"sudo snap connect matter-all-clusters-app:otbr-dbus-wpan0 openthread-border-router:dbus-wpan0",
		"sudo snap start matter-all-clusters-app",
	}
	for _, cmd := range commands {
		exec(t, cmd)
	}

	WaitForLogMessage(t, "matter-all-clusters-app", "CHIP minimal mDNS started advertising", start)
	t.Log("Matter All Clusters App is ready")
}

func exec(t *testing.T, command string) string {
	t.Helper()

	t.Logf("[exec-ssh] %s", command)

	// Remote commands that require sudo might ask for the password. Always pass it in. See https://stackoverflow.com/a/11955358
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		escapedPassword := strings.ReplaceAll(remotePassword, `"`, `\"`)
		command = fmt.Sprintf(`echo "%s" | sudo -S %s`, escapedPassword, command)
	}

	if SSHClient == nil {
		t.Fatalf("SSH client not initialized. Please connect to remote device first")
	}

	session, err := SSHClient.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := session.Start(command); err != nil {
		t.Fatalf("Failed to start session with command '%s': %v", command, err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read command output: %v", err)
	}

	if err := session.Wait(); err != nil {
		t.Fatalf("Command '%s' failed: %v", command, err)
	}

	return string(output)
}

func WaitForLogMessage(t *testing.T, snap string, expectedLog string, start time.Time) {
	t.Helper()

	const maxRetry = 10
	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: '%s'", i, maxRetry, expectedLog)

		command := fmt.Sprintf("sudo journalctl --utc --since \"%s\" --no-pager | grep \"%s\"|| true", start.UTC().Format("2006-01-02 15:04:05"), snap)
		logs := exec(t, command)
		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: '%s'", expectedLog)
			return
		}
	}

	t.Logf("Time out: reached max %d retries.", maxRetry)
	t.Log(exec(t, "journalctl --no-pager --lines=10 --unit=snap.openthread-border-router.otbr-agent --priority=notice"))
	t.FailNow()
}

func dumpLogs(t *testing.T, label string, start time.Time) error {
	command := fmt.Sprintf("sudo journalctl --utc --since \"%s\" --no-pager | grep \"%s\"|| true", start.UTC().Format("2006-01-02 15:04:05"), label)
	logs := exec(t, command)
	return utils.WriteLogFile(t, "remote-"+label, logs)
}
