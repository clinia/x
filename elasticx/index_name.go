package elasticx

import "strings"

type IndexName string

const pathSeparator = "~"

func NewIndexName(elements ...string) IndexName {
	return IndexName(strings.Join(elements, pathSeparator))
}

// FullIndexName create an index name with the base clinia segement, the engine name and the app index name
// e.g. "clinia-engines~<engine-name~<...elements>"
func FullIndexName(engineName string, elements ...string) IndexName {
	parts := []string{enginesIndexNameSegment, engineName}
	parts = append(parts, elements...)

	return IndexName(strings.Join(parts, pathSeparator))
}

func (i IndexName) Elements() []string {
	return strings.Split(string(i), pathSeparator)
}

// EngineName returns the engine name from the index name.
// If the index name doesn't contain the base segment, it means that the engine name is at index 0.
// If the index name contains the base segment, the engine name is at index 1.
func (i IndexName) EngineName() string {
	elems := i.Elements()

	// Safeguard against empty index name
	if len(elems) == 0 {
		return ""
	}

	// If the index name contains the base segment, the engine name is at index 1
	if elems[0] == enginesIndexNameSegment {
		return elems[1]
	}

	// Otherwise, the engine name is at index 0
	return elems[0]
}

// Name returns the app index name from the index name.
// If the index name doesn't contain the base segment, it means that the app index name is all the segments from index 1.
// If the index name contains the base segment, the app index name is all the segments from index 2.
func (i IndexName) Name() string {
	elems := i.Elements()

	// Safeguard against empty index name
	if len(elems) == 0 {
		return ""
	}

	// If the index name contains only 2 segments, the app index name is empty
	if len(elems) < 3 {
		return ""
	}

	// If the index name contains the base segment, the app index name is all the segments from index 2
	if elems[0] == enginesIndexNameSegment {
		return strings.Join(elems[2:], pathSeparator)
	}

	// Otherwise, the app index name is all the segments from index 1
	return strings.Join(elems[1:], pathSeparator)
}

func (i IndexName) String() string {
	return string(i)
}
