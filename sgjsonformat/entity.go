package sgjsonformat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func unorderdEqual(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}

type SgJsonFile struct {
	filepath string
	Entities []Entity
	ById     map[string]Entity
	Ids      []string
}

// A Entity is the default type of an SGDatabase entry.
type Entity struct {
	Id            string                   `json:"id"`
	Name          string                   `json:"name"`
	Data          EntityData               `json:"data"`
	Meta          []map[string]interface{} `json:"meta"`
	Selectors     []string                 `json:"selectors"`
	Domain        EntityDomain             `json:"domain"`
	Tags          []string                 `json:"tags"`
	FormatVersion string                   `json:"formatVersion"`
	IsValid       bool
}

type EntityDiff struct {
	a                        *Entity
	b                        *Entity
	HasDifferences           bool
	NameIsDifferent          bool
	DataIsDifferent          bool
	MetaIsDifferent          bool
	SelectorsIsDifferent     bool
	DomainIsDifferent        bool
	TagsIsDifferent          bool
	FormatVersionIsDifferent bool
}

// custom data type to accomodate json array & json object usage of data field
type EntityData []EntityDataEntry

type EntityDomain struct {
	Category string `json:"category"`
	Kind     string `json:"kind"`
}

type EntityDataEntry struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (entity *Entity) Compare(other *Entity) {
	if !entity.IsValid {
		fmt.Printf("%15s: %38s ≠ %38s\n", "Id", "- missing -", other.Id)
	} else if !other.IsValid {
		fmt.Printf("%15s: %38s ≠ %38s\n", "Id", entity.Id, "- missing -")
	} else if !entity.IsValid && !other.IsValid {
		fmt.Println(" both entities invalid !!")
	}

	diff := NewEntityDiff(entity, other)
	if diff.HasDifferences {
		fmt.Printf("%15s: %38s ≠ %38s\n", "Id", entity.Id, other.Id)
	}

}

func RemoveDuplicates[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// custom unmarshaler for EntityData
func (e *EntityData) UnmarshalJSON(b []byte) error {
	var list []EntityDataEntry
	if err := json.Unmarshal(b, &list); err != nil {
		var element EntityDataEntry
		if err := json.Unmarshal(b, &element); err != nil {
			return err
		}
		*e = append(*e, element)
	}
	*e = append(*e, list...)
	return nil
}

func NewEntityDiff(a *Entity, b *Entity) *EntityDiff {
	if a.FormatVersion == "" {
		a.FormatVersion = "1.0"
	}
	if b.FormatVersion == "" {
		b.FormatVersion = "1.0"
	}
	j := &EntityDiff{
		a: a,
		b: b,

		NameIsDifferent: a.Name != b.Name,

		SelectorsIsDifferent:     !unorderdEqual(a.Selectors, b.Selectors),
		DomainIsDifferent:        a.Domain != b.Domain,
		TagsIsDifferent:          !unorderdEqual(a.Tags, b.Tags),
		FormatVersionIsDifferent: a.FormatVersion != b.FormatVersion,
	}
	if j.NameIsDifferent || j.DataIsDifferent || j.MetaIsDifferent || j.SelectorsIsDifferent ||
		j.DomainIsDifferent || j.TagsIsDifferent || j.FormatVersionIsDifferent {
		j.HasDifferences = true
	}
	return j
}

func NewSgJsonFile(path string) *SgJsonFile {
	j := &SgJsonFile{
		filepath: path,
		ById:     make(map[string]Entity),
		Ids:      make([]string, 0),
	}
	return j
}

func (j *SgJsonFile) Read() error {
	jsonFile, err := os.Open(j.filepath)

	if err != nil {
		return fmt.Errorf("read: failed opening json file: %w", err)
	}
	defer jsonFile.Close()

	bytes, _ := io.ReadAll(jsonFile)
	var result []Entity
	err = json.Unmarshal([]byte(bytes), &result)
	if err != nil {
		return fmt.Errorf("read: failed parsing json file %s: %w", j.filepath, err)
	}

	j.Entities = result
	for _, e := range j.Entities {
		e.IsValid = true
		j.ById[e.Id] = e
		j.Ids = append(j.Ids, e.Id)
	}

	return nil
}
