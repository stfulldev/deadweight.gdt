package tscn

import (
	"io"
	"strconv"
)

type valueKind uint8

const (
	valueScalar valueKind = iota
	valueCall
	valueComplex
)

type headerValue struct {
	kind          valueKind
	scalar        string
	function      string
	firstArgument string
	hasFirstArg   bool
	position      Position
}

type sectionHeader struct {
	name       string
	attributes map[string]headerValue
	position   Position
}

type delimiter struct {
	expected tokenType
	position Position
}

type parser struct {
	source    string
	lexer     *lexer
	lookahead *token
	document  *Document
	sections  int
	sawScene  bool
}

// Parse reads the supported Godot 4 TSCN subset from reader.
func Parse(reader io.Reader, source string) (*Document, error) {
	state := &parser{
		source: source,
		lexer:  newLexer(reader, source),
		document: &Document{
			ExtResources: make(map[string]ExtResource),
		},
	}

	if err := state.parse(); err != nil {
		return nil, err
	}

	return state.document, nil
}

func (state *parser) parse() error {
	for {
		if err := state.skipNewlines(); err != nil {
			return err
		}

		current, err := state.peek()
		if err != nil {
			return err
		}
		if current.typeID == tokenEOF {
			break
		}
		if current.typeID != tokenLeftBracket {
			return newParseError(state.source, current.position, "expected a section header, found %s", current.describe())
		}

		header, err := state.parseSectionHeader()
		if err != nil {
			return err
		}

		if state.sections == 0 && header.name != "gd_scene" {
			return newParseError(state.source, header.position, "first section must be [gd_scene], found [%s]", header.name)
		}
		state.sections++

		var bodyNode *Node
		switch header.name {
		case "gd_scene":
			if err := state.applySceneHeader(header); err != nil {
				return err
			}
		case "ext_resource":
			if err := state.addExternalResource(header); err != nil {
				return err
			}
		case "node":
			if err := state.addNode(header); err != nil {
				return err
			}
			bodyNode = &state.document.Nodes[len(state.document.Nodes)-1]
		case "editable":
			state.document.Features.HasEditable = true
		}

		if err := state.parseSectionBody(bodyNode); err != nil {
			return err
		}
	}

	if !state.sawScene {
		return newParseError(state.source, Position{Line: 1, Column: 1}, "missing [gd_scene] header")
	}
	if len(state.document.Nodes) == 0 {
		return newParseError(state.source, Position{Line: 1, Column: 1}, "scene must contain exactly one root node")
	}

	return nil
}

func (state *parser) parseSectionHeader() (sectionHeader, error) {
	opening, err := state.expect(tokenLeftBracket, "[")
	if err != nil {
		return sectionHeader{}, err
	}

	name, err := state.next()
	if err != nil {
		return sectionHeader{}, err
	}
	if name.typeID != tokenIdentifier {
		return sectionHeader{}, newParseError(state.source, name.position, "expected section name, found %s", name.describe())
	}

	header := sectionHeader{
		name:       name.literal,
		attributes: make(map[string]headerValue),
		position:   opening.position,
	}

	for {
		current, err := state.peek()
		if err != nil {
			return sectionHeader{}, err
		}

		switch current.typeID {
		case tokenRightBracket:
			_, _ = state.next()
			if err := state.consumeHeaderLineEnd(); err != nil {
				return sectionHeader{}, err
			}
			return header, nil
		case tokenNewline, tokenEOF:
			return sectionHeader{}, newParseError(state.source, current.position, "unterminated [%s] section header", header.name)
		case tokenIdentifier:
			key, _ := state.next()
			if _, duplicate := header.attributes[key.literal]; duplicate {
				return sectionHeader{}, newParseError(state.source, key.position, "duplicate %q attribute in [%s]", key.literal, header.name)
			}
			if _, err := state.expect(tokenEquals, "="); err != nil {
				return sectionHeader{}, err
			}
			value, err := state.parseHeaderValue()
			if err != nil {
				return sectionHeader{}, err
			}
			header.attributes[key.literal] = value
		default:
			return sectionHeader{}, newParseError(state.source, current.position, "expected attribute or ], found %s", current.describe())
		}
	}
}

func (state *parser) parseHeaderValue() (headerValue, error) {
	current, err := state.next()
	if err != nil {
		return headerValue{}, err
	}

	switch current.typeID {
	case tokenString, tokenInteger, tokenBool:
		return headerValue{kind: valueScalar, scalar: current.literal, position: current.position}, nil
	case tokenIdentifier:
		next, err := state.peek()
		if err != nil {
			return headerValue{}, err
		}
		if next.typeID != tokenLeftParen {
			return headerValue{kind: valueScalar, scalar: current.literal, position: current.position}, nil
		}

		opening, _ := state.next()
		firstArgument, hasFirstArgument, err := state.consumeBalanced(opening)
		if err != nil {
			return headerValue{}, err
		}

		return headerValue{
			kind:          valueCall,
			function:      current.literal,
			firstArgument: firstArgument,
			hasFirstArg:   hasFirstArgument,
			position:      current.position,
		}, nil
	case tokenLeftBracket, tokenLeftParen, tokenLeftBrace:
		if _, _, err := state.consumeBalanced(current); err != nil {
			return headerValue{}, err
		}
		return headerValue{kind: valueComplex, position: current.position}, nil
	default:
		return headerValue{}, newParseError(state.source, current.position, "expected header value, found %s", current.describe())
	}
}

func (state *parser) consumeBalanced(opening token) (string, bool, error) {
	expected, ok := closingToken(opening.typeID)
	if !ok {
		return "", false, newParseError(state.source, opening.position, "internal parser error: %s is not an opening delimiter", opening.describe())
	}

	stack := []delimiter{{expected: expected, position: opening.position}}
	var firstArgument string
	var hasFirstArgument bool

	for len(stack) > 0 {
		current, err := state.next()
		if err != nil {
			return "", false, err
		}
		if current.typeID == tokenEOF {
			return "", false, newParseError(state.source, opening.position, "unclosed %s delimiter", opening.describe())
		}

		if closing, isOpening := closingToken(current.typeID); isOpening {
			stack = append(stack, delimiter{expected: closing, position: current.position})
			continue
		}
		if isClosingToken(current.typeID) {
			top := stack[len(stack)-1]
			if current.typeID != top.expected {
				return "", false, newParseError(state.source, current.position, "mismatched closing delimiter %s", current.describe())
			}
			stack = stack[:len(stack)-1]
			continue
		}

		if len(stack) == 1 && !hasFirstArgument {
			switch current.typeID {
			case tokenString, tokenInteger, tokenBool, tokenIdentifier:
				firstArgument = current.literal
				hasFirstArgument = true
			}
		}
	}

	return firstArgument, hasFirstArgument, nil
}

func (state *parser) applySceneHeader(header sectionHeader) error {
	if state.sawScene {
		return newParseError(state.source, header.position, "duplicate [gd_scene] header")
	}

	format, exists, err := scalarAttribute(state.source, header, "format")
	if err != nil {
		return err
	}
	if !exists {
		return newParseError(state.source, header.position, "[gd_scene] must define format=3")
	}

	formatNumber, err := strconv.ParseInt(format, 10, 32)
	if err != nil {
		return newParseError(state.source, header.attributes["format"].position, "[gd_scene] format must be an integer")
	}
	if formatNumber != 3 {
		return newParseError(state.source, header.attributes["format"].position, "unsupported Godot scene format %d; expected format=3", formatNumber)
	}

	uid, _, err := scalarAttribute(state.source, header, "uid")
	if err != nil {
		return err
	}

	state.document.Header = SceneHeader{Format: int(formatNumber), UID: uid}
	state.sawScene = true
	return nil
}

func (state *parser) addExternalResource(header sectionHeader) error {
	id, exists, err := scalarAttribute(state.source, header, "id")
	if err != nil {
		return err
	}
	if !exists || id == "" {
		return newParseError(state.source, header.position, "[ext_resource] must define a non-empty id")
	}
	if previous, duplicate := state.document.ExtResources[id]; duplicate {
		return newParseError(
			state.source,
			header.attributes["id"].position,
			"duplicate external resource id %q; first declared at line %d",
			id,
			previous.Position.Line,
		)
	}

	resourceType, _, err := scalarAttribute(state.source, header, "type")
	if err != nil {
		return err
	}
	path, _, err := scalarAttribute(state.source, header, "path")
	if err != nil {
		return err
	}
	uid, _, err := scalarAttribute(state.source, header, "uid")
	if err != nil {
		return err
	}

	state.document.ExtResources[id] = ExtResource{
		ID:       id,
		Type:     resourceType,
		UID:      uid,
		Path:     path,
		Position: header.position,
	}
	return nil
}

func (state *parser) addNode(header sectionHeader) error {
	name, exists, err := scalarAttribute(state.source, header, "name")
	if err != nil {
		return err
	}
	if !exists || name == "" {
		return newParseError(state.source, header.position, "[node] must define a non-empty name")
	}

	nodeType, _, err := scalarAttribute(state.source, header, "type")
	if err != nil {
		return err
	}
	parent, hasParent, err := scalarAttribute(state.source, header, "parent")
	if err != nil {
		return err
	}
	owner, _, err := scalarAttribute(state.source, header, "owner")
	if err != nil {
		return err
	}
	placeholder, _, err := scalarAttribute(state.source, header, "instance_placeholder")
	if err != nil {
		return err
	}

	node := Node{
		Name:                name,
		Type:                nodeType,
		Parent:              parent,
		Owner:               owner,
		InstancePlaceholder: placeholder,
		Position:            header.position,
	}

	if index, exists, err := scalarAttribute(state.source, header, "index"); err != nil {
		return err
	} else if exists {
		indexNumber, parseErr := strconv.ParseInt(index, 10, 32)
		if parseErr != nil {
			return newParseError(state.source, header.attributes["index"].position, "node index must be an integer")
		}
		converted := int(indexNumber)
		node.Index = &converted
	}

	if instance, exists := header.attributes["instance"]; exists {
		if instance.kind != valueCall || !instance.hasFirstArg || (instance.function != ResourceRefExternal && instance.function != ResourceRefInternal) {
			return newParseError(state.source, instance.position, "node instance must be ExtResource(...) or SubResource(...)")
		}
		node.Instance = &ResourceRef{Kind: instance.function, ID: instance.firstArgument}
	}
	if node.Instance != nil && node.InstancePlaceholder != "" {
		return newParseError(state.source, header.position, "node cannot define both instance and instance_placeholder")
	}

	if len(state.document.Nodes) == 0 {
		if hasParent {
			return newParseError(state.source, header.attributes["parent"].position, "root node must not define parent")
		}
		if node.Instance != nil {
			state.document.Features.HasInheritedRoot = true
		}
	} else if !hasParent {
		return newParseError(state.source, header.position, "non-root node %q must define parent", node.Name)
	}

	if node.Type == "" && node.Instance == nil && node.InstancePlaceholder == "" {
		state.document.Features.HasOverrideNodes = true
	}

	state.document.Nodes = append(state.document.Nodes, node)
	return nil
}

func (state *parser) parseSectionBody(node *Node) error {
	for {
		if err := state.skipNewlines(); err != nil {
			return err
		}

		current, err := state.peek()
		if err != nil {
			return err
		}
		if current.typeID == tokenEOF || current.typeID == tokenLeftBracket {
			return nil
		}
		if current.typeID != tokenIdentifier {
			return newParseError(state.source, current.position, "expected property name or section header, found %s", current.describe())
		}

		property, _ := state.next()
		if _, err := state.expect(tokenEquals, "="); err != nil {
			return err
		}

		if node != nil && property.literal == "shadow_enabled" {
			if node.ShadowEnabled != nil {
				return newParseError(state.source, property.position, "duplicate shadow_enabled property")
			}
			value, err := state.parseBooleanProperty(property)
			if err != nil {
				return err
			}
			node.ShadowEnabled = &value
			continue
		}

		if err := state.skipPropertyValue(property); err != nil {
			return err
		}
	}
}

func (state *parser) parseBooleanProperty(property token) (bool, error) {
	value, err := state.next()
	if err != nil {
		return false, err
	}
	if value.typeID != tokenBool {
		return false, newParseError(state.source, value.position, "%s must be true or false", property.literal)
	}

	ending, err := state.next()
	if err != nil {
		return false, err
	}
	if ending.typeID != tokenNewline && ending.typeID != tokenEOF {
		return false, newParseError(state.source, ending.position, "unexpected token after %s boolean value", property.literal)
	}

	return value.literal == "true", nil
}

func (state *parser) skipPropertyValue(property token) error {
	var stack []delimiter
	sawValue := false

	for {
		current, err := state.next()
		if err != nil {
			return err
		}

		switch current.typeID {
		case tokenEOF:
			if len(stack) > 0 {
				return newParseError(state.source, stack[len(stack)-1].position, "unclosed delimiter in %s property", property.literal)
			}
			if !sawValue {
				return newParseError(state.source, property.position, "property %s has no value", property.literal)
			}
			return nil
		case tokenNewline:
			if len(stack) == 0 {
				if !sawValue {
					return newParseError(state.source, property.position, "property %s has no value", property.literal)
				}
				return nil
			}
			continue
		}

		sawValue = true
		if closing, isOpening := closingToken(current.typeID); isOpening {
			stack = append(stack, delimiter{expected: closing, position: current.position})
			continue
		}
		if isClosingToken(current.typeID) {
			if len(stack) == 0 || stack[len(stack)-1].expected != current.typeID {
				return newParseError(state.source, current.position, "mismatched closing delimiter %s in %s property", current.describe(), property.literal)
			}
			stack = stack[:len(stack)-1]
		}
	}
}

func (state *parser) consumeHeaderLineEnd() error {
	current, err := state.next()
	if err != nil {
		return err
	}
	if current.typeID != tokenNewline && current.typeID != tokenEOF {
		return newParseError(state.source, current.position, "unexpected token after section header")
	}
	return nil
}

func (state *parser) skipNewlines() error {
	for {
		current, err := state.peek()
		if err != nil {
			return err
		}
		if current.typeID != tokenNewline {
			return nil
		}
		_, _ = state.next()
	}
}

func (state *parser) expect(expected tokenType, description string) (token, error) {
	current, err := state.next()
	if err != nil {
		return token{}, err
	}
	if current.typeID != expected {
		return token{}, newParseError(state.source, current.position, "expected %s, found %s", description, current.describe())
	}
	return current, nil
}

func (state *parser) peek() (token, error) {
	if state.lookahead == nil {
		current, err := state.lexer.nextToken()
		if err != nil {
			return token{}, err
		}
		state.lookahead = &current
	}
	return *state.lookahead, nil
}

func (state *parser) next() (token, error) {
	if state.lookahead != nil {
		current := *state.lookahead
		state.lookahead = nil
		return current, nil
	}
	return state.lexer.nextToken()
}

func scalarAttribute(source string, header sectionHeader, name string) (string, bool, error) {
	value, exists := header.attributes[name]
	if !exists {
		return "", false, nil
	}
	if value.kind != valueScalar {
		return "", false, newParseError(source, value.position, "attribute %q in [%s] must be a scalar", name, header.name)
	}
	return value.scalar, true, nil
}

func closingToken(opening tokenType) (tokenType, bool) {
	switch opening {
	case tokenLeftBracket:
		return tokenRightBracket, true
	case tokenLeftParen:
		return tokenRightParen, true
	case tokenLeftBrace:
		return tokenRightBrace, true
	default:
		return tokenInvalid, false
	}
}

func isClosingToken(value tokenType) bool {
	return value == tokenRightBracket || value == tokenRightParen || value == tokenRightBrace
}
