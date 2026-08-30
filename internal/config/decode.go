package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
)

type configWire struct {
	Version       json.RawMessage `json:"version"`
	Preset        json.RawMessage `json:"preset"`
	Profile       json.RawMessage `json:"profile"`
	FailOnPartial json.RawMessage `json:"fail_on_partial"`
	Budgets       json.RawMessage `json:"budgets"`
	Profiles      json.RawMessage `json:"profiles"`
}

type profileWire struct {
	Name        json.RawMessage `json:"name"`
	Description json.RawMessage `json:"description"`
	Extends     json.RawMessage `json:"extends"`
	Platform    json.RawMessage `json:"platform"`
	Renderer    json.RawMessage `json:"renderer"`
	TargetFPS   json.RawMessage `json:"target_fps"`
	Quality     json.RawMessage `json:"quality"`
	Budgets     json.RawMessage `json:"budgets"`
}

type budgetsWire struct {
	Nodes             json.RawMessage `json:"nodes"`
	TreeDepth         json.RawMessage `json:"tree_depth"`
	SceneInstances    json.RawMessage `json:"scene_instances"`
	MeshInstances     json.RawMessage `json:"mesh_instances"`
	Lights            json.RawMessage `json:"lights"`
	ShadowLights      json.RawMessage `json:"shadow_lights"`
	ExternalResources json.RawMessage `json:"external_resources"`
	SceneDependencies json.RawMessage `json:"scene_dependencies"`
}

// Decode strictly decodes and statically validates one version-one document.
func Decode(reader io.Reader, source string) (Config, error) {
	if reader == nil {
		return Config{}, configError(
			ReasonDecode,
			source,
			"",
			"config reader is required",
			nil,
		)
	}

	raw, err := decodeOneDocument(reader, source)
	if err != nil {
		return Config{}, err
	}
	if isNull(raw) {
		return Config{}, nullError(source, "")
	}

	var wire configWire
	if err := decodeStrict(raw, &wire, source, ""); err != nil {
		return Config{}, err
	}

	version, err := decodeRequiredInt64(wire.Version, source, "version")
	if err != nil {
		return Config{}, err
	}
	preset, err := decodeOptionalString(wire.Preset, source, "preset")
	if err != nil {
		return Config{}, err
	}
	profile, err := decodeOptionalString(wire.Profile, source, "profile")
	if err != nil {
		return Config{}, err
	}
	failOnPartial, err := decodeOptionalBool(wire.FailOnPartial, source, "fail_on_partial")
	if err != nil {
		return Config{}, err
	}
	limits, err := decodeBudgets(wire.Budgets, source, "budgets")
	if err != nil {
		return Config{}, err
	}
	profiles, err := decodeProfiles(wire.Profiles, source)
	if err != nil {
		return Config{}, err
	}

	configuration := Config{
		Version:       int(version),
		Preset:        preset,
		Profile:       profile,
		FailOnPartial: failOnPartial,
		Budgets:       limits,
		Profiles:      profiles,

		failOnPartialDeclared: len(wire.FailOnPartial) > 0,
	}
	if err := Validate(configuration, source); err != nil {
		return Config{}, err
	}

	return configuration.Clone(), nil
}

func decodeOneDocument(reader io.Reader, source string) (json.RawMessage, error) {
	decoder := json.NewDecoder(reader)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		detail := "cannot decode JSON document"
		if err == io.EOF {
			detail = "config document is empty"
		}
		return nil, configError(ReasonDecode, source, "", detail, err)
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, configError(
			ReasonDecode,
			source,
			"",
			"trailing JSON data after config document",
			err,
		)
	}

	return append(json.RawMessage(nil), raw...), nil
}

func decodeStrict(raw json.RawMessage, target any, source, prefix string) error {
	if isNull(raw) {
		return nullError(source, prefix)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		field := prefix
		if unknown := unknownField(err); unknown != "" {
			field = joinField(prefix, unknown)
		}
		return configError(ReasonDecode, source, field, "invalid JSON value", err)
	}

	return nil
}

func decodeProfiles(raw json.RawMessage, source string) (map[string]Profile, error) {
	profiles := make(map[string]Profile)
	if len(raw) == 0 {
		return profiles, nil
	}
	if isNull(raw) {
		return nil, nullError(source, "profiles")
	}

	var rawProfiles map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawProfiles); err != nil {
		return nil, configError(ReasonDecode, source, "profiles", "must be an object", err)
	}
	keys := make([]string, 0, len(rawProfiles))
	for id := range rawProfiles {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		prefix := joinField("profiles", id)
		profile, err := decodeProfile(rawProfiles[id], source, prefix)
		if err != nil {
			return nil, err
		}
		profiles[id] = profile
	}

	return profiles, nil
}

func decodeProfile(raw json.RawMessage, source, prefix string) (Profile, error) {
	var wire profileWire
	if err := decodeStrict(raw, &wire, source, prefix); err != nil {
		return Profile{}, err
	}

	var result Profile
	fields := []struct {
		name   string
		raw    json.RawMessage
		target **string
	}{
		{name: "name", raw: wire.Name, target: &result.Name},
		{name: "description", raw: wire.Description, target: &result.Description},
		{name: "extends", raw: wire.Extends, target: &result.Extends},
		{name: "platform", raw: wire.Platform, target: &result.Platform},
		{name: "renderer", raw: wire.Renderer, target: &result.Renderer},
		{name: "quality", raw: wire.Quality, target: &result.Quality},
	}
	for _, field := range fields {
		value, err := decodeOptionalString(field.raw, source, joinField(prefix, field.name))
		if err != nil {
			return Profile{}, err
		}
		*field.target = value
	}

	var err error
	result.TargetFPS, err = decodeOptionalInt64(wire.TargetFPS, source, joinField(prefix, "target_fps"))
	if err != nil {
		return Profile{}, err
	}
	result.Budgets, err = decodeBudgets(wire.Budgets, source, joinField(prefix, "budgets"))
	if err != nil {
		return Profile{}, err
	}

	return result, nil
}

func decodeBudgets(raw json.RawMessage, source, prefix string) (budget.Limits, error) {
	if len(raw) == 0 {
		return budget.Limits{}, nil
	}

	var wire budgetsWire
	if err := decodeStrict(raw, &wire, source, prefix); err != nil {
		return budget.Limits{}, err
	}

	result := budget.Limits{}
	fields := []struct {
		name   string
		raw    json.RawMessage
		target **int64
	}{
		{name: "nodes", raw: wire.Nodes, target: &result.Nodes},
		{name: "tree_depth", raw: wire.TreeDepth, target: &result.TreeDepth},
		{name: "scene_instances", raw: wire.SceneInstances, target: &result.SceneInstances},
		{name: "mesh_instances", raw: wire.MeshInstances, target: &result.MeshInstances},
		{name: "lights", raw: wire.Lights, target: &result.Lights},
		{name: "shadow_lights", raw: wire.ShadowLights, target: &result.ShadowLights},
		{name: "external_resources", raw: wire.ExternalResources, target: &result.ExternalResources},
		{name: "scene_dependencies", raw: wire.SceneDependencies, target: &result.SceneDependencies},
	}
	for _, field := range fields {
		value, err := decodeOptionalInt64(field.raw, source, joinField(prefix, field.name))
		if err != nil {
			return budget.Limits{}, err
		}
		*field.target = value
	}

	return result, nil
}

func decodeRequiredInt64(raw json.RawMessage, source, field string) (int64, error) {
	if len(raw) == 0 {
		return 0, configError(ReasonValidation, source, field, "field is required", nil)
	}
	value, err := decodeOptionalInt64(raw, source, field)
	if err != nil {
		return 0, err
	}

	return *value, nil
}

func decodeOptionalInt64(raw json.RawMessage, source, field string) (*int64, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if isNull(raw) {
		return nil, nullError(source, field)
	}

	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, configError(ReasonDecode, source, field, "must be an integer", err)
	}

	return &value, nil
}

func decodeOptionalString(raw json.RawMessage, source, field string) (*string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if isNull(raw) {
		return nil, nullError(source, field)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, configError(ReasonDecode, source, field, "must be a string", err)
	}

	return &value, nil
}

func decodeOptionalBool(raw json.RawMessage, source, field string) (bool, error) {
	if len(raw) == 0 {
		return false, nil
	}
	if isNull(raw) {
		return false, nullError(source, field)
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, configError(ReasonDecode, source, field, "must be a boolean", err)
	}

	return value, nil
}

func isNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func nullError(source, field string) *Error {
	return configError(ReasonDecode, source, field, "must not be null", nil)
}

func unknownField(err error) string {
	const prefix = "json: unknown field "
	message := err.Error()
	if !strings.HasPrefix(message, prefix) {
		return ""
	}

	var field string
	if _, scanErr := fmt.Sscanf(strings.TrimPrefix(message, prefix), "%q", &field); scanErr != nil {
		return ""
	}

	return field
}

func joinField(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if name == "" {
		return prefix
	}

	return prefix + "." + name
}
