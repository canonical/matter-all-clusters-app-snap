package matter

const (
	AllClustersSnap = "matter-all-clusters-app"
	ChipToolSnap    = "chip-tool"
	OtbrSnap        = "openthread-border-router"
	OTCTL           = OtbrSnap + ".ot-ctl"

	DefaultInfraInterfaceValue = "wlan0"
	InfraInterfaceKey          = "infra-if"
	LocalInfraInterfaceEnv     = "LOCAL_INFRA_IF"
	RemoteInfraInterfaceEnv    = "REMOTE_INFRA_IF"

	DefaultRadioUrl   = "spinel+hdlc+uart:///dev/ttyACM0"
	RadioUrlKey       = "radio-url"
	LocalRadioUrlEnv  = "LOCAL_RADIO_URL"
	RemoteRadioUrlEnv = "REMOTE_RADIO_URL"

	RemoteHostEnv     = "REMOTE_HOST"
	RemoteUserEnv     = "REMOTE_USER"
	RemotePasswordEnv = "REMOTE_PASSWORD"
)
