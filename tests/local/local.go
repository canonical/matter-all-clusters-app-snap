package local

import (
	"os"
	"strings"
	"testing"
	"time"

	"chip-tool-snap-tests/matter"
	"github.com/stretchr/testify/require"

	"github.com/canonical/matter-snap-testing/utils"
)

func Setup(t *testing.T) {
	// Clean
	utils.SnapRemove(t, matter.OtbrSnap)
	utils.SnapRemove(t, matter.ChipToolSnap)

	// Install stable chip tool from store
	chipToolInstallTime := time.Now()
	require.NoError(t, utils.SnapInstallFromStore(t, matter.ChipToolSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, chipToolInstallTime, matter.ChipToolSnap)
		utils.SnapRemove(t, matter.ChipToolSnap)
	})

	// Install OTBR
	otbrInstallTime := time.Now()
	require.NoError(t, utils.SnapInstallFromStore(t, matter.OtbrSnap, "latest/beta"))
	t.Cleanup(func() {
		utils.SnapDumpLogs(t, otbrInstallTime, matter.OtbrSnap)
		utils.SnapRemove(t, matter.OtbrSnap)
	})

	// Connect interfaces
	snapInterfaces := []string{"avahi-control", "firewall-control", "raw-usb", "network-control", "bluetooth-control", "bluez"}
	for _, interfaceSlot := range snapInterfaces {
		require.NoError(t, utils.SnapConnect(nil, matter.OtbrSnap+":"+interfaceSlot, ""))
	}

	// Set infra interface
	if v := os.Getenv(matter.LocalInfraInterfaceEnv); v != "" {
		infraInterfaceValue := v
		utils.SnapSet(nil, matter.OtbrSnap, matter.InfraInterfaceKey, infraInterfaceValue)
	} else {
		utils.SnapSet(nil, matter.OtbrSnap, matter.InfraInterfaceKey, matter.DefaultInfraInterfaceValue)
	}

	// Set radio url
	if v := os.Getenv(matter.LocalRadioUrlEnv); v != "" {
		radioUrlValue := v
		utils.SnapSet(nil, matter.OtbrSnap, matter.RadioUrlKey, radioUrlValue)
	} else {
		utils.SnapSet(nil, matter.OtbrSnap, matter.RadioUrlKey, matter.DefaultRadioUrl)
	}

	// Start OTBR
	otbrStartTime := time.Now()
	utils.SnapStart(t, matter.OtbrSnap)
	utils.WaitForLogMessage(t, matter.OtbrSnap, "Start Thread Border Agent: OK", otbrStartTime)

	// Form Thread network
	utils.Exec(t, "sudo "+matter.OTCTL+" dataset init new")
	utils.Exec(t, "sudo "+matter.OTCTL+" dataset commit active")
	utils.Exec(t, "sudo "+matter.OTCTL+" ifconfig up")
	utils.Exec(t, "sudo "+matter.OTCTL+" thread start")
	utils.WaitForLogMessage(t, matter.OtbrSnap, "Thread Network", otbrStartTime)
}

func GetActiveDataset(t *testing.T) string {
	activeDataset, _, _ := utils.Exec(t, "sudo "+matter.OTCTL+" dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}
