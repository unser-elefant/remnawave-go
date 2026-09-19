package payload

type Scope string

const (
	ScopeUser            Scope = "user"
	ScopeUserHWIDDevices Scope = "user_hwid_devices"
	ScopeService         Scope = "service"
)
