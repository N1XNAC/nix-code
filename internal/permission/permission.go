package permission

import (
	"fmt"
	"sync"
)

type PermissionLevel string

const (
	Allow PermissionLevel = "allow"
	Deny  PermissionLevel = "deny"
	Ask   PermissionLevel = "ask"
)

type PermissionService struct {
	mu          sync.RWMutex
	rules       map[string]PermissionLevel
	pendingAsk  map[string]bool
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		rules:      make(map[string]PermissionLevel),
		pendingAsk: make(map[string]bool),
	}
}

func (ps *PermissionService) SetRule(toolName string, level PermissionLevel) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.rules[toolName] = level
}

func (ps *PermissionService) GetRule(toolName string) PermissionLevel {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if level, ok := ps.rules[toolName]; ok {
		return level
	}
	return Allow
}

func (ps *PermissionService) Check(toolName string, sessionID string) (bool, error) {
	level := ps.GetRule(toolName)
	switch level {
	case Allow:
		return true, nil
	case Deny:
		return false, fmt.Errorf("tool '%s' is denied in current mode", toolName)
	case Ask:
		ps.mu.Lock()
		ps.pendingAsk[sessionID+":"+toolName] = true
		ps.mu.Unlock()
		return false, fmt.Errorf("tool '%s' requires approval. Type 'y' to allow or 'n' to deny", toolName)
	}
	return true, nil
}

func (ps *PermissionService) Approve(sessionID string, toolName string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.pendingAsk, sessionID+":"+toolName)
}

func (ps *PermissionService) SetAgentPermissions(agentName string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if agentName == "think" {
		ps.rules = map[string]PermissionLevel{
			"read":     Allow,
			"grep":     Allow,
			"glob":     Allow,
			"webfetch": Allow,
			"write":    Deny,
			"edit":     Deny,
			"bash":     Ask,
		}
	} else {
		ps.rules = map[string]PermissionLevel{
			"read":      Allow,
			"write":     Allow,
			"edit":      Allow,
			"bash":      Ask,
			"grep":      Allow,
			"glob":      Allow,
			"todowrite": Allow,
			"webfetch":  Allow,
		}
	}
}

func (ps *PermissionService) LoadFromConfig(permissions map[string]string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for tool, level := range permissions {
		ps.rules[tool] = PermissionLevel(level)
	}
}
