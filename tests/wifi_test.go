package tests

import (
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllClustersAppWiFi(t *testing.T) {
	// Start clean
	utils.SnapRemove(t, chipToolSnap)
	utils.SnapRemove(t, allClustersSnap)

	// Set start time for capturing logs after removal, before installing version to test
	start := time.Now()

	// Install stable chip tool from store
	require.NoError(t, utils.SnapInstallFromStore(t, chipToolSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, chipToolSnap)
		utils.SnapRemove(t, chipToolSnap)
	})

	// Install all clusters app
	installAllClusters(t)

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, allClustersSnap)
		utils.SnapRemove(t, allClustersSnap)
	})

	// Setup all clusters app
	utils.SnapSet(t, allClustersSnap, "args", "--wifi")
	require.NoError(t, utils.SnapConnect(t, allClustersSnap+":avahi-control", ""))
	require.NoError(t, utils.SnapConnect(t, allClustersSnap+":bluez", ""))

	// Start all clusters app
	utils.SnapStart(t, allClustersSnap)
	utils.WaitForLogMessage(t,
		allClustersSnap, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, chipToolSnap, stdout))
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, chipToolSnap, stdout))

		waitForOnOffHandlingByAllClustersApp(t, start)
	})

}
