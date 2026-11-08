//go:build !race

package message

func auditRaceEnabled() bool { return false }
