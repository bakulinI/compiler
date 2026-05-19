package parser

import (
	"fmt"
	"strings"
)

type ASTField struct {
	Name  string
	Value string
}

type ASTNode struct {
	Name     string
	Fields   []ASTField
	Children []*ASTNode
	Line     int
	Column   int
}

func NewNode(name string) *ASTNode {
	return &ASTNode{Name: name}
}

func (n *ASTNode) AddField(name, value string) {
	n.Fields = append(n.Fields, ASTField{Name: name, Value: value})
}

func (n *ASTNode) AddChild(child *ASTNode) {
	if child != nil {
		n.Children = append(n.Children, child)
	}
}

func (n *ASTNode) SetPosition(line, column int) {
	n.Line = line
	n.Column = column
}

func (n *ASTNode) String() string {
	if n == nil {
		return "<empty AST>"
	}

	var sb strings.Builder
	n.write(&sb, "", true, true)
	return sb.String()
}

func (n *ASTNode) write(sb *strings.Builder, prefix string, isLast bool, isRoot bool) {
	if isRoot {
		sb.WriteString(n.Name)
		sb.WriteString("\n")
	} else {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("└── ")
		} else {
			sb.WriteString("├── ")
		}
		sb.WriteString(n.Name)
		sb.WriteString("\n")
	}

	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	total := len(n.Fields) + len(n.Children)
	index := 0

	for _, field := range n.Fields {
		index++
		writeLine(sb, childPrefix, index == total, fmt.Sprintf("%s: %s", field.Name, field.Value))
	}

	for _, child := range n.Children {
		index++
		child.write(sb, childPrefix, index == total, false)
	}
}

func writeLine(sb *strings.Builder, prefix string, isLast bool, text string) {
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
	} else {
		sb.WriteString("├── ")
	}
	sb.WriteString(text)
	sb.WriteString("\n")
}
