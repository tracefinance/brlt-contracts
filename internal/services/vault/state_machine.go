package vault

// stateTransitions defines valid state transitions for a vault
var stateTransitions = map[VaultStatus][]VaultStatus{
	VaultStatusPending: {
		VaultStatusDeploying,
		VaultStatusFailed,
	},
	VaultStatusDeploying: {
		VaultStatusActive,
		VaultStatusFailed,
	},
	VaultStatusActive: {
		VaultStatusRecovering,
		VaultStatusPaused,
	},
	VaultStatusRecovering: {
		VaultStatusActive,
		VaultStatusRecovered,
		VaultStatusFailed,
	},
	VaultStatusPaused: {
		VaultStatusActive,
		VaultStatusRecovering,
	},
	VaultStatusRecovered: {},
	VaultStatusFailed:    {},
}

// CanTransition checks if a transition from one state to another is valid
func CanTransition(from, to VaultStatus) bool {
	validTransitions, exists := stateTransitions[from]
	if !exists {
		return false
	}

	for _, validState := range validTransitions {
		if validState == to {
			return true
		}
	}

	return false
}
