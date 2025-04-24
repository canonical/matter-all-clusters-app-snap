package tests

import (
	"log"
	"testing"
	"time"

	"all-clusters-tests/local"
	"all-clusters-tests/shared"
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) {
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

	// Install all clusters app stable
	require.NoError(t, utils.SnapInstallFromStore(t, shared.AllClustersSnap, "latest/stable"))

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

	t.Run("Control before upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, shared.AllClustersSnap)
		snapRevision := utils.SnapRevision(t, shared.AllClustersSnap)
		log.Printf("%s installed version %s build %s\n", shared.AllClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff on 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, shared.ChipToolSnap, stdout))

		local.WaitForOnOffHandlingByAllClustersApp(t, start)
	})

	t.Run("Upgrade snap", func(t *testing.T) {
		local.UpgradeAllClustersApp(t)
	})

	t.Run("Control after upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, shared.AllClustersSnap)
		snapRevision := utils.SnapRevision(t, shared.AllClustersSnap)
		log.Printf("%s installed version %s build %s\n", shared.AllClustersSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff off 110 1 2>&1")
		assert.NoError(t, utils.WriteLogFile(t, shared.ChipToolSnap, stdout))

		local.WaitForOnOffHandlingByAllClustersApp(t, start)
	})

}
