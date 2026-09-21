package payload

type Event string

const (
	EventUserCreated                        Event = "user.created"
	EventUserModified                       Event = "user.modified"
	EventUserDeleted                        Event = "user.deleted"
	EventUserRevoked                        Event = "user.revoked"
	EventUserDisabled                       Event = "user.disabled"
	EventUserEnabled                        Event = "user.enabled"
	EventUserLimited                        Event = "user.limited"
	EventUserExpired                        Event = "user.expired"
	EventUserTrafficReset                   Event = "user.traffic_reset"
	EventUserFirstConnected                 Event = "user.first_connected"
	EventUserBandwidthUsageThresholdReached Event = "user.bandwidth_usage_threshold_reached"
	EventUserNotConnected                   Event = "user.not_connected"
	EventUserExpiration                     Event = "user.expiration"
)

const (
	EventUserHWIDDevicesAdded   Event = "user_hwid_devices.added"
	EventUserHWIDDevicesDeleted Event = "user_hwid_devices.deleted"
)

const (
	EventNodeCreated            Event = "node.created"
	EventNodeModified           Event = "node.modified"
	EventNodeDisabled           Event = "node.disabled"
	EventNodeEnabled            Event = "node.enabled"
	EventNodeDeleted            Event = "node.deleted"
	EventNodeConnectinoLost     Event = "node.connection_lost"
	EventNodeConnectionRestored Event = "node.connection_restored"
	EventNodeTrafficNotify      Event = "node.traffic_notify"
)

const (
	EventServicePanelStarted         Event = "service.panel_started"
	EventServiceLoginAttemptFailed   Event = "service.login_attempt_failed"
	EventServiceLoginAttemptSuccess  Event = "service.login_attempt_success"
	EventServiceSubpageConfigChanged Event = "service.subpage_config_changed"
	EventServiceApiTokenCreated      Event = "service.api_token_created"
	EventServiceApiTokenDeleted      Event = "service.api_token_deleted"
)

const (
	EventBandwidthUsageThresholdReachedMaxNotifications Event = "errors.bandwidth_usage_threshold_reached_max_notifications"
)

const (
	EventInfraBillingNodePaymentIn7Days      Event = "crm.infra_billing_node_payment_in_7_days"
	EventInfraBillingNodePaymentIn24Days     Event = "crm.infra_billing_node_payment_in_24_days"
	EventInfraBillingNodePaymentIn48Days     Event = "crm.infra_billing_node_payment_in_48_days"
	EventInfraBillingNodePaymentDueToday     Event = "crm.infra_billing_node_payment_due_today"
	EventInfraBillingNodePaymentOverdue24Hrs Event = "crm.infra_billing_node_payment_overdue_24hrs"
	EventInfraBillingNodePaymentOverdue48Hrs Event = "crm.infra_billing_node_payment_overdue_48hrs"
	EventInfraBillingNodePaymentOverdue7Days Event = "crm.infra_billing_node_payment_overdue_7_days"
)

const (
	EventTorrentBlockerReport Event = "torrent_blocker.report"
)

func (e Event) RequiresTelegramID() bool {
	switch e {
	case EventUserExpiration, EventUserExpired:
		return true
	default:
		return false
	}
}
