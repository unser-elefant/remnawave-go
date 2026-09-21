package payload

type Scope string

const (
	ScopeUser            Scope = "user"
	ScopeUserHWIDDevices Scope = "user_hwid_devices"
	ScopeNode            Scope = "node"
	ScopeService         Scope = "service"
	ScopeErrors          Scope = "errors"
	ScopeCRM             Scope = "crm"
	ScopeTorrentBlocker  Scope = "torrent_blocker"
)
