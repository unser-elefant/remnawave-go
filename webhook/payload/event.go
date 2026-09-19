package payload

type Event string

const (
	EventUserCreated        Event = "user.created"
	EventUserModified       Event = "user.modified"
	EventUserExpiration     Event = "user.expiration"
	EventUserExpired        Event = "user.expired"
	EventUserFirstConnected Event = "user.first_connected"
)

const (
	EventUserHWIDDevicesAdded   Event = "user_hwid_devices.added"
	EventUserHWIDDevicesDeleted Event = "user_hwid_devices.deleted"
)

const (
	EventServiceLoginAttemptSuccess Event = "service.login_attempt_success"
	EventServiceLoginAttemptFailed  Event = "service.login_attempt_failed"
)

func (e Event) RequiresTelegramID() bool {
	switch e {
	case EventUserExpiration, EventUserExpired:
		return true
	default:
		return false
	}
}
