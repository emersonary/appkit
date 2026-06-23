package permissions

// Action bitmask values for be_action and granted_actions.
const (
	ActionList   = 1
	ActionCreate = 2
	ActionUpdate = 4
	ActionDelete = 8
)

// AllActions is the default full CRUD mask.
const AllActions = ActionList | ActionCreate | ActionUpdate | ActionDelete

// ValidAction reports whether bit is a known single action flag.
func ValidAction(bit int) bool {
	switch bit {
	case ActionList, ActionCreate, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

// MaskAllows reports whether mask includes bit.
func MaskAllows(mask, bit int) bool {
	return bit != 0 && mask&bit != 0
}
