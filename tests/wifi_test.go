package tests

import (
	"testing"
	"time"

	"all-clusters-tests/local"
	"all-clusters-tests/shared"
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllClustersAppWiFi(t *testing.T) {
	// Start clean
	utils.SnapRemove(t, shared.ChipToolSnap)
	utils.SnapRemove(t, shared.AllClustersSnap)

	// Set start time for capturing logs after removal, before installing version to test
	start := time.Now()

	// Install stable chip tool from store
	require.NoError(t, utils.SnapInstallFromStore(t, shared.ChipToolSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, shared.ChipToolSnap)
		utils.SnapRemove(t, shared.ChipToolSnap)
	})

	// Install all clusters app
	local.InstallAllClustersApp(t)

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, shared.AllClustersSnap)
		utils.SnapRemove(t, shared.AllClustersSnap)
	})

	// Setup all clusters app
	utils.SnapSet(t, shared.AllClustersSnap, "args", "--wifi")
	require.NoError(t, utils.SnapConnect(t, shared.AllClustersSnap+":avahi-control", ""))
	require.NoError(t, utils.SnapConnect(t, shared.AllClustersSnap+":bluez", ""))

	// Start all clusters app
	utils.SnapStart(t, shared.AllClustersSnap)
	utils.WaitForLogMessage(t,
		shared.AllClustersSnap, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, shared.ChipToolSnap, stdout))
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, shared.ChipToolSnap, stdout))

		local.WaitForOnOffHandlingByAllClustersApp(t, start)
	})

}
