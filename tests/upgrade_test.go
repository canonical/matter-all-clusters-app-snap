package tests

import (
	"log"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) { // Start clean
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

	// Install all clusters app stable
	require.NoError(t, utils.SnapInstallFromStore(t, allClustersSnap, "latest/stable"))

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

	t.Run("Control before upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, allClustersSnap)
		snapRevision := utils.SnapRevision(t, allClustersSnap)
		log.Printf("%s installed version %s build %s\n", allClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff on 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, chipToolSnap, stdout))

		waitForOnOffHandlingByAllClustersApp(t, start)
	})

	t.Run("Upgrade snap", func(t *testing.T) {
		upgradeAllClusters(t)
	})

	t.Run("Control after upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, allClustersSnap)
		snapRevision := utils.SnapRevision(t, allClustersSnap)
		log.Printf("%s installed version %s build %s\n", allClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff off 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, chipToolSnap, stdout))

		waitForOnOffHandlingByAllClustersApp(t, start)
	})

}
