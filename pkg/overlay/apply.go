package overlay

import (
	"github.com/pb33f/jsonpath/pkg/jsonpath"
	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"go.yaml.in/yaml/v4"
)

// ApplyTo will take an overlay and apply its changes to the given YAML
// document.
func (o *Overlay) ApplyTo(root *yaml.Node) error {
	for _, action := range o.Actions {
		var err error
		if action.Remove {
			err = applyRemoveAction(root, action)
		} else {
			err = applyUpdateAction(root, action)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func applyRemoveAction(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	idx := newParentIndex(root)

	p, err := jsonpath.NewPath(action.Target, config.WithPropertyNameExtension())
	if err != nil {
		return err
	}

	nodes := p.Query(root)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		removeNode(idx, node)
	}

	return nil
}

func removeNode(idx parentIndex, node *yaml.Node) {
	parent := idx.getParent(node)
	if parent == nil {
		return
	}

	for i, child := range parent.Content {
		if child == node {
			switch parent.Kind {
			case yaml.MappingNode:
				if i%2 == 1 {
					// if we select a value, we should delete the key too
					parent.Content = append(parent.Content[:i-1], parent.Content[i+1:]...)
				} else {
					// if we select a key, we should delete the value
					parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
				}
				return
			case yaml.SequenceNode:
				parent.Content = append(parent.Content[:i], parent.Content[i+1:]...)
				return
			}
		}
	}
}

func applyUpdateAction(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	if action.Update.IsZero() {
		return nil
	}

	p, err := jsonpath.NewPath(action.Target, config.WithPropertyNameExtension())
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	if len(nodes) == 0 && action.Upsert {
		return createPath(root, p, &action.Update)
	}

	for _, node := range nodes {
		if err := updateNode(node, &action.Update); err != nil {
			return err
		}
	}

	return nil
}

func updateNode(node *yaml.Node, updateNode *yaml.Node) error {
	mergeNode(node, updateNode)
	return nil
}

func mergeNode(node *yaml.Node, merge *yaml.Node) {
	if node.Kind != merge.Kind {
		*node = *clone(merge)
		return
	}
	switch node.Kind {
	default:
		node.Value = merge.Value
	case yaml.MappingNode:
		mergeMappingNode(node, merge)
	case yaml.SequenceNode:
		mergeSequenceNode(node, merge)
	}
}

// mergeMappingNode will perform a shallow merge of the merge node into the main
// node.
func mergeMappingNode(node *yaml.Node, merge *yaml.Node) {
NextKey:
	for i := 0; i < len(merge.Content); i += 2 {
		mergeKey := merge.Content[i].Value
		mergeValue := merge.Content[i+1]

		for j := 0; j < len(node.Content); j += 2 {
			nodeKey := node.Content[j].Value
			if nodeKey == mergeKey {
				mergeNode(node.Content[j+1], mergeValue)
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], clone(mergeValue))
	}
}

// mergeSequenceNode will append the merge node's content to the original node.
func mergeSequenceNode(node *yaml.Node, merge *yaml.Node) {
	node.Content = append(node.Content, clone(merge).Content...)
}

func clone(node *yaml.Node) *yaml.Node {
	newNode := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
	}
	if node.Alias != nil {
		newNode.Alias = clone(node.Alias)
	}
	if node.Content != nil {
		newNode.Content = make([]*yaml.Node, len(node.Content))
		for i, child := range node.Content {
			newNode.Content[i] = clone(child)
		}
	}
	return newNode
}

func createPath(root *yaml.Node, p *jsonpath.JSONPath, value *yaml.Node) error {
	segments, err := p.GetSegmentInfo()
	if err != nil {
		return err
	}

	if len(segments) == 0 {
		return &EmptyPathError{}
	}

	if err := validatePathForCreation(root, segments); err != nil {
		return err
	}

	current := root
	if root.Kind == yaml.DocumentNode && len(root.Content) == 1 {
		current = root.Content[0]
	}

	for i, seg := range segments[:len(segments)-1] {
		next, err := ensureSegment(current, seg, segments[i+1])
		if err != nil {
			return err
		}
		current = next
	}

	lastSeg := segments[len(segments)-1]
	return setFinalValue(current, lastSeg, value)
}

func validatePathForCreation(root *yaml.Node, segments []jsonpath.SegmentInfo) error {
	current := root
	if root.Kind == yaml.DocumentNode && len(root.Content) == 1 {
		current = root.Content[0]
	}

	for i, seg := range segments[:len(segments)-1] {
		switch seg.Kind {
		case jsonpath.SegmentKindMemberName:
			if current == nil {
				nextSeg := segments[i+1]
				if nextSeg.Kind == jsonpath.SegmentKindArrayIndex && nextSeg.Index != 0 {
					return &CannotCreateArrayAtIndexError{Index: nextSeg.Index}
				}
				continue
			}
			if current.Kind == yaml.MappingNode {
				found := false
				for j := 0; j < len(current.Content); j += 2 {
					if current.Content[j].Value == seg.Key {
						current = current.Content[j+1]
						found = true
						break
					}
				}
				if found {
					continue
				}
			}
			nextSeg := segments[i+1]
			if nextSeg.Kind == jsonpath.SegmentKindArrayIndex {
				if nextSeg.Index < 0 {
					return &NegativeArrayIndexError{Index: nextSeg.Index}
				}
				if nextSeg.Index != 0 {
					return &CannotCreateArrayAtIndexError{Index: nextSeg.Index}
				}
			}
			current = nil

		case jsonpath.SegmentKindArrayIndex:
			if seg.Index < 0 {
				return &NegativeArrayIndexError{Index: seg.Index}
			}
			if current == nil {
				if seg.Index != 0 {
					return &CannotCreateArrayAtIndexError{Index: seg.Index}
				}
				continue
			}
			if current.Kind == yaml.SequenceNode {
				if seg.Index > int64(len(current.Content)) {
					return &ArrayIndexOutOfBoundsError{
						Index:     seg.Index,
						Length:    len(current.Content),
						CanAppend: seg.Index == int64(len(current.Content)),
					}
				}
				if seg.Index < int64(len(current.Content)) {
					current = current.Content[seg.Index]
				} else {
					current = nil
				}
			} else if current.Kind != yaml.MappingNode {
				return &TypeMismatchError{Expected: "sequence node", Actual: kindString(current.Kind)}
			}
		}
	}

	lastSeg := segments[len(segments)-1]
	if lastSeg.Kind == jsonpath.SegmentKindArrayIndex {
		if lastSeg.Index < 0 {
			return &NegativeArrayIndexError{Index: lastSeg.Index}
		}
		if current == nil {
			if lastSeg.Index != 0 {
				return &CannotCreateArrayAtIndexError{Index: lastSeg.Index}
			}
		} else if current.Kind == yaml.SequenceNode {
			if lastSeg.Index > int64(len(current.Content)) {
				return &ArrayIndexOutOfBoundsError{
					Index:     lastSeg.Index,
					Length:    len(current.Content),
					CanAppend: lastSeg.Index == int64(len(current.Content)),
				}
			}
		}
	}

	return nil
}

func kindString(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document node"
	case yaml.SequenceNode:
		return "sequence node"
	case yaml.MappingNode:
		return "mapping node"
	case yaml.ScalarNode:
		return "scalar node"
	case yaml.AliasNode:
		return "alias node"
	default:
		return "unknown node"
	}
}

func ensureSegment(node *yaml.Node, seg, nextSeg jsonpath.SegmentInfo) (*yaml.Node, error) {
	switch seg.Kind {
	case jsonpath.SegmentKindMemberName:
		return ensureMapKey(node, seg.Key, nextSeg)
	case jsonpath.SegmentKindArrayIndex:
		return ensureArrayIndex(node, seg.Index, nextSeg)
	default:
		return nil, &UnknownSegmentKindError{Kind: seg.Kind}
	}
}

func ensureMapKey(node *yaml.Node, key string, nextSeg jsonpath.SegmentInfo) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, &TypeMismatchError{Expected: "mapping node", Actual: kindString(node.Kind)}
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], nil
		}
	}

	var newKind yaml.Kind
	if nextSeg.Kind == jsonpath.SegmentKindArrayIndex {
		newKind = yaml.SequenceNode
	} else {
		newKind = yaml.MappingNode
	}

	newNode := &yaml.Node{
		Kind:    newKind,
		Tag:     "!!map",
		Content: []*yaml.Node{},
	}
	if newKind == yaml.SequenceNode {
		newNode.Tag = "!!seq"
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}
	node.Content = append(node.Content, keyNode, newNode)

	return newNode, nil
}

func ensureArrayIndex(node *yaml.Node, index int64, nextSeg jsonpath.SegmentInfo) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, &TypeMismatchError{Expected: "sequence node", Actual: kindString(node.Kind)}
	}

	if index < 0 {
		return nil, &NegativeArrayIndexError{Index: index}
	}

	if index == int64(len(node.Content)) {
		child := createPlaceholderForSegment(nextSeg)
		node.Content = append(node.Content, child)
		return child, nil
	}

	if index > int64(len(node.Content)) {
		return nil, &ArrayIndexOutOfBoundsError{
			Index:     index,
			Length:    len(node.Content),
			CanAppend: true,
		}
	}

	return node.Content[index], nil
}

func createPlaceholderForSegment(nextSeg jsonpath.SegmentInfo) *yaml.Node {
	var newKind yaml.Kind
	if nextSeg.Kind == jsonpath.SegmentKindArrayIndex {
		newKind = yaml.SequenceNode
	} else {
		newKind = yaml.MappingNode
	}

	newNode := &yaml.Node{
		Kind:    newKind,
		Content: []*yaml.Node{},
	}
	if newKind == yaml.SequenceNode {
		newNode.Tag = "!!seq"
	} else {
		newNode.Tag = "!!map"
	}

	return newNode
}

func setFinalValue(node *yaml.Node, seg jsonpath.SegmentInfo, value *yaml.Node) error {
	switch seg.Kind {
	case jsonpath.SegmentKindMemberName:
		return setMapValue(node, seg.Key, value)
	case jsonpath.SegmentKindArrayIndex:
		return setArrayValue(node, seg.Index, value)
	default:
		return &UnknownSegmentKindError{Kind: seg.Kind}
	}
}

func setMapValue(node *yaml.Node, key string, value *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return &TypeMismatchError{Expected: "mapping node", Actual: kindString(node.Kind)}
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			mergeNode(node.Content[i+1], value)
			return nil
		}
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}
	node.Content = append(node.Content, keyNode, clone(value))
	return nil
}

func setArrayValue(node *yaml.Node, index int64, value *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return &TypeMismatchError{Expected: "sequence node", Actual: kindString(node.Kind)}
	}

	if index < 0 {
		return &NegativeArrayIndexError{Index: index}
	}

	if index == int64(len(node.Content)) {
		node.Content = append(node.Content, clone(value))
		return nil
	}

	if index > int64(len(node.Content)) {
		return &ArrayIndexOutOfBoundsError{
			Index:     index,
			Length:    len(node.Content),
			CanAppend: true,
		}
	}

	mergeNode(node.Content[index], value)
	return nil
}
