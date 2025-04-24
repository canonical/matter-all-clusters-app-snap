package local

import (
	"os"
	"strings"
	"testing"
	"time"

	"all-clusters-tests/shared"
	"github.com/stretchr/testify/require"

	"github.com/canonical/matter-snap-testing/utils"
)

func Setup(t *testing.T) {
	// Clean
	utils.SnapRemove(t, shared.OtbrSnap)
	utils.SnapRemove(t, shared.ChipToolSnap)

	// Install stable chip tool from store
	chipToolInstallTime := time.Now()
	require.NoError(t, utils.SnapInstallFromStore(t, shared.ChipToolSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, chipToolInstallTime, shared.ChipToolSnap)
		utils.SnapRemove(t, shared.ChipToolSnap)
	})

	// Install OTBR
	otbrInstallTime := time.Now()
	require.NoError(t, utils.SnapInstallFromStore(t, shared.OtbrSnap, "latest/beta"))
	t.Cleanup(func() {
		utils.SnapDumpLogs(t, otbrInstallTime, shared.OtbrSnap)
		utils.SnapRemove(t, shared.OtbrSnap)
	})

	// Connect interfaces
	snapInterfaces := []string{"avahi-control", "firewall-control", "raw-usb", "network-control", "bluetooth-control", "bluez"}
	for _, interfaceSlot := range snapInterfaces {
		require.NoError(t, utils.SnapConnect(nil, shared.OtbrSnap+":"+interfaceSlot, ""))
	}

	// Set infra interface
	if v := os.Getenv(shared.LocalInfraInterfaceEnv); v != "" {
		infraInterfaceValue := v
		utils.SnapSet(nil, shared.OtbrSnap, shared.InfraInterfaceKey, infraInterfaceValue)
	} else {
		utils.SnapSet(nil, shared.OtbrSnap, shared.InfraInterfaceKey, shared.DefaultInfraInterfaceValue)
	}

	// Set radio url
	if v := os.Getenv(shared.LocalRadioUrlEnv); v != "" {
		radioUrlValue := v
		utils.SnapSet(nil, shared.OtbrSnap, shared.RadioUrlKey, radioUrlValue)
	} else {
		utils.SnapSet(nil, shared.OtbrSnap, shared.RadioUrlKey, shared.DefaultRadioUrl)
	}

	// Start OTBR
	otbrStartTime := time.Now()
	utils.SnapStart(t, shared.OtbrSnap)
	utils.WaitForLogMessage(t, shared.OtbrSnap, "Start Thread Border Agent: OK", otbrStartTime)

	// Form Thread network
	utils.Exec(t, "sudo "+shared.OTCTL+" dataset init new")
	utils.Exec(t, "sudo "+shared.OTCTL+" dataset commit active")
	utils.Exec(t, "sudo "+shared.OTCTL+" ifconfig up")
	utils.Exec(t, "sudo "+shared.OTCTL+" thread start")
	utils.WaitForLogMessage(t, shared.OtbrSnap, "Thread Network", otbrStartTime)
}

func GetActiveDataset(t *testing.T) string {
	activeDataset, _, _ := utils.Exec(t, "sudo "+shared.OTCTL+" dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}
