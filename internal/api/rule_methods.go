package api

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/tg-manager/internal/forwarder"
	"github.com/tg-manager/internal/storage"
)

// rules.create
type RulesCreateMethod struct {
	storage *storage.Storage
	engine  *forwarder.Engine
}

type createRuleParams struct {
	SourceChannelID int64  `json:"source_channel_id"`
	SourceName      string `json:"source_name"`
	SourceHash      int64  `json:"source_hash"`
	TargetChannelID int64  `json:"target_channel_id"`
	TargetName      string `json:"target_name"`
	TargetHash      int64  `json:"target_hash"`
	MatchPattern    string `json:"match_pattern"`
}

func (m *RulesCreateMethod) Name() string { return "rules.create" }
func (m *RulesCreateMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p createRuleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.SourceChannelID == 0 || p.TargetChannelID == 0 || p.MatchPattern == "" {
		return nil, fmt.Errorf("source_channel_id, target_channel_id, and match_pattern are required")
	}
	if _, err := regexp.Compile(p.MatchPattern); err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	rule := storage.ForwardRule{
		SourceChannelID: p.SourceChannelID,
		SourceName:      p.SourceName,
		SourceHash:      p.SourceHash,
		TargetChannelID: p.TargetChannelID,
		TargetName:      p.TargetName,
		TargetHash:      p.TargetHash,
		MatchPattern:    p.MatchPattern,
		Enabled:         true,
	}

	if err := m.storage.GetDB().Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}

	_ = m.engine.ReloadRules()
	return rule, nil
}

// rules.list
type RulesListMethod struct {
	storage *storage.Storage
}

func (m *RulesListMethod) Name() string { return "rules.list" }
func (m *RulesListMethod) Execute(ctx context.Context, _ json.RawMessage) (interface{}, error) {
	var rules []storage.ForwardRule
	if err := m.storage.GetDB().Order("id desc").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	return rules, nil
}

// rules.update
type RulesUpdateMethod struct {
	storage *storage.Storage
	engine  *forwarder.Engine
}

type updateRuleParams struct {
	ID              uint   `json:"id"`
	SourceChannelID *int64 `json:"source_channel_id,omitempty"`
	SourceName      *string `json:"source_name,omitempty"`
	SourceHash      *int64  `json:"source_hash,omitempty"`
	TargetChannelID *int64 `json:"target_channel_id,omitempty"`
	TargetName      *string `json:"target_name,omitempty"`
	TargetHash      *int64  `json:"target_hash,omitempty"`
	MatchPattern    *string `json:"match_pattern,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
}

func (m *RulesUpdateMethod) Name() string { return "rules.update" }
func (m *RulesUpdateMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p updateRuleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}

	var rule storage.ForwardRule
	if err := m.storage.GetDB().First(&rule, p.ID).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	updates := make(map[string]interface{})
	if p.SourceChannelID != nil {
		updates["source_channel_id"] = *p.SourceChannelID
	}
	if p.SourceName != nil {
		updates["source_name"] = *p.SourceName
	}
	if p.SourceHash != nil {
		updates["source_hash"] = *p.SourceHash
	}
	if p.TargetChannelID != nil {
		updates["target_channel_id"] = *p.TargetChannelID
	}
	if p.TargetName != nil {
		updates["target_name"] = *p.TargetName
	}
	if p.TargetHash != nil {
		updates["target_hash"] = *p.TargetHash
	}
	if p.MatchPattern != nil {
		if _, err := regexp.Compile(*p.MatchPattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		updates["match_pattern"] = *p.MatchPattern
	}
	if p.Enabled != nil {
		updates["enabled"] = *p.Enabled
	}

	if err := m.storage.GetDB().Model(&rule).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}

	// Reload updated rule
	m.storage.GetDB().First(&rule, p.ID)
	_ = m.engine.ReloadRules()
	return rule, nil
}

// rules.delete
type RulesDeleteMethod struct {
	storage *storage.Storage
	engine  *forwarder.Engine
}

type deleteRuleParams struct {
	ID uint `json:"id"`
}

func (m *RulesDeleteMethod) Name() string { return "rules.delete" }
func (m *RulesDeleteMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p deleteRuleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}

	if err := m.storage.GetDB().Delete(&storage.ForwardRule{}, p.ID).Error; err != nil {
		return nil, fmt.Errorf("delete rule: %w", err)
	}

	_ = m.engine.ReloadRules()
	return map[string]bool{"deleted": true}, nil
}
