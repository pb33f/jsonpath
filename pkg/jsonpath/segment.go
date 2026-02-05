package jsonpath

import (
    "go.yaml.in/yaml/v4"
    "strings"
)

type segmentKind int

const (
    segmentKindChild       segmentKind = iota // .
    segmentKindDescendant                     // ..
    segmentKindProperyName                    // ~ (extension only)
    segmentKindParent                         // ^ (JSONPath Plus parent selector)
)

type segment struct {
    kind       segmentKind
    child      *innerSegment
    descendant *innerSegment
}

type segmentSubKind int

const (
    segmentDotWildcard   segmentSubKind = iota // .*
    segmentDotMemberName                       // .property
    segmentLongHand                            // [ selector[] ]
)

func (s segment) ToString() string {
    switch s.kind {
    case segmentKindChild:
        if s.child.kind != segmentLongHand {
            return "." + s.child.ToString()
        } else {
            return s.child.ToString()
        }
    case segmentKindDescendant:
        return ".." + s.descendant.ToString()
    case segmentKindProperyName:
        return "~"
    case segmentKindParent:
        return "^"
    }
    panic("unknown segment kind")
}

type innerSegment struct {
    kind      segmentSubKind
    dotName   string
    selectors []*selector
}

func (s innerSegment) ToString() string {
    builder := strings.Builder{}
    switch s.kind {
    case segmentDotWildcard:
        builder.WriteString("*")
        break
    case segmentDotMemberName:
        builder.WriteString(s.dotName)
        break
    case segmentLongHand:
        builder.WriteString("[")
        for i, selector := range s.selectors {
            builder.WriteString(selector.ToString())
            if i < len(s.selectors)-1 {
                builder.WriteString(", ")
            }
        }
        builder.WriteString("]")
        break
    default:
        panic("unknown child segment kind")
    }
    return builder.String()
}

func descendApply(value *yaml.Node, apply func(*yaml.Node)) {
    if value == nil {
        return
    }
    stack := []*yaml.Node{value}
    for len(stack) > 0 {
        n := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        apply(n)
        if len(n.Content) > 0 {
            for i := len(n.Content) - 1; i >= 0; i-- {
                stack = append(stack, n.Content[i])
            }
        }
    }
}
