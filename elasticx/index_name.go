package elasticx

import "strings"

type IndexName string

const pathSeparator = "~"

func NewIndexName(elements ...string) IndexName {
	return IndexName(strings.Join(elements, pathSeparator))
}

func (i IndexName) elements() []string {
	return strings.Split(string(i), pathSeparator)
}

func (i IndexName) EngineName() string {
	elems := i.elements()
	if len(elems) < 2 {
		return ""
	}
	return i.elements()[1]
}

func (i IndexName) Name() string {
	elems := i.elements()
	if len(elems) < 3 {
		return ""
	}
	return strings.Join(elems[2:], pathSeparator)
}

func (i IndexName) String() string {
	return string(i)
}
