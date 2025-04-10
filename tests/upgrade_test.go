package tests

import (
	"log"
	"testing"
	"time"

	"chip-tool-snap-tests/local"
	"chip-tool-snap-tests/matter"
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) { // Start clean
	utils.SnapRemove(t, matter.ChipToolSnap)
	utils.SnapRemove(t, matter.AllClustersSnap)

	// Set start time for capturing logs after removal, before installing version to test
	start := time.Now()

	// Install stable chip tool from store
	require.NoError(t, utils.SnapInstallFromStore(t, matter.ChipToolSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, matter.ChipToolSnap)
		utils.SnapRemove(t, matter.ChipToolSnap)
	})

	// Install all clusters app stable
	require.NoError(t, utils.SnapInstallFromStore(t, matter.AllClustersSnap, "latest/stable"))

	t.Cleanup(func() {
		utils.SnapDumpLogs(t, start, matter.AllClustersSnap)
		utils.SnapRemove(t, matter.AllClustersSnap)
	})

	// Setup all clusters app
	utils.SnapSet(t, matter.AllClustersSnap, "args", "--wifi")
	require.NoError(t, utils.SnapConnect(t, matter.AllClustersSnap+":avahi-control", ""))
	require.NoError(t, utils.SnapConnect(t, matter.AllClustersSnap+":bluez", ""))

	// Start all clusters app
	utils.SnapStart(t, matter.AllClustersSnap)
	utils.WaitForLogMessage(t,
		matter.AllClustersSnap, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, matter.ChipToolSnap, stdout))
	})

	t.Run("Control before upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, matter.AllClustersSnap)
		snapRevision := utils.SnapRevision(t, matter.AllClustersSnap)
		log.Printf("%s installed version %s build %s\n", matter.AllClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff on 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, matter.ChipToolSnap, stdout))

		local.WaitForOnOffHandlingByAllClustersApp(t, start)
	})

	t.Run("Upgrade snap", func(t *testing.T) {
		local.UpgradeAllClusters(t)
	})

	t.Run("Control after upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, matter.AllClustersSnap)
		snapRevision := utils.SnapRevision(t, matter.AllClustersSnap)
		log.Printf("%s installed version %s build %s\n", matter.AllClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff off 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, matter.ChipToolSnap, stdout))

		local.WaitForOnOffHandlingByAllClustersApp(t, start)
	})

}
