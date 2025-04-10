package local

import (
	"testing"
	"time"

	"chip-tool-snap-tests/matter"
	"github.com/canonical/matter-snap-testing/env"
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/require"
)

func InstallAllClusters(t *testing.T) {
	if env.SnapPath() != "" {
		require.NoError(t,
			utils.SnapInstallFromFile(nil, env.SnapPath()),
		)
	} else {
		require.NoError(t,
			utils.SnapInstallFromStore(nil, matter.AllClustersSnap, env.SnapChannel()),
		)
	}
}

func UpgradeAllClusters(t *testing.T) {
	if env.SnapPath() != "" {
		require.NoError(t,
			utils.SnapInstallFromFile(t, env.SnapPath()),
		)
	} else {
		utils.SnapRefresh(t, matter.AllClustersSnap, "latest/edge")
	}
}

func WaitForOnOffHandlingByAllClustersApp(t *testing.T, start time.Time) {
	// 0x6 is the Matter Cluster ID for on-off
	// Using cluster ID here because of a buffering issue in the log stream:
	// https://github.com/canonical/chip-tool-snap/pull/69#issuecomment-2207189962
	utils.WaitForLogMessage(t, matter.AllClustersSnap, "ClusterId = 0x6", start)
}
