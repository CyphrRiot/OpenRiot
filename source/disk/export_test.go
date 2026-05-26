package disk

// Exported test bridges for package disk_test.

var FilterDrives = filterDrives

type ActionType = actionType

const (
	ActionDiscover   = actionDiscover
	ActionMount      = actionMount
	ActionUmount     = actionUmount
	ActionFormat     = actionFormat
	ActionEncrypt    = actionEncrypt
	ActionBenchmark  = actionBenchmark
)
