package sgjsonformat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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

// SgJsonFile is the representation of one `.sg/*.json` file from the .sg database
// Use the read method to read all stored entities into the Entities field.
// Entites are also mapped by id in the ById map.
type SgJsonFile struct {
	// the filepath to the json this SgJsonFile is constructed from
	filepath string
	Entities []Entity
	ById     map[string]Entity
	// a list of all entity ids that are present in this file
	Ids []string
}

func NewSgJsonFile(path string) *SgJsonFile {
	j := &SgJsonFile{
		filepath: path,
		ById:     make(map[string]Entity),
		Ids:      make([]string, 0),
	}
	return j
}

func (j *SgJsonFile) Write() error {
	jsonFile, err := os.Create(j.filepath)

	if err != nil {
		return fmt.Errorf("write: failed opening json file: %w", err)
	}
	defer jsonFile.Close()
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(j.Entities)
	if err != nil {
		return fmt.Errorf("write: failed encoding entities: %w", err)
	}
	return nil
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

// A Entity is the default type of an SgJsonFile entry.
type Entity struct {
	Id            string       `json:"id"`
	Name          string       `json:"name"`
	Data          EntityData   `json:"data"`
	Meta          EntityMeta   `json:"meta"`
	Selectors     []string     `json:"selectors"`
	Domain        EntityDomain `json:"domain"`
	Tags          []string     `json:"tags"`
	FormatVersion string       `json:"formatVersion"`
	IsValid       bool
}

func (entity *Entity) Merge(other *Entity) Entity {
	if !entity.IsValid {
		return *other
	} else if !other.IsValid {
		return *entity
	} else if !entity.IsValid && !other.IsValid {
		fmt.Println("!!!!!!! both entities invalid !!!!!")
		return Entity{}
	}
	diff := NewEntityDiff(entity, other)
	if !diff.HasDifferences {
		return *entity
	} else {
		return *other
	}
}

func (entity *Entity) Compare(other *Entity) {
	if !entity.IsValid {
		fmt.Printf("%20s: %38s ≠ %-38s\n", "================= Id", "- missing -", other.Id)
		return
	} else if !other.IsValid {
		fmt.Printf("%20s: %38s ≠ %-38s\n", "================= Id", entity.Id, "- missing -")
		return
	} else if !entity.IsValid && !other.IsValid {
		fmt.Println("!!!!!!! both entities invalid !!!!!")
		return
	}

	diff := NewEntityDiff(entity, other)
	if diff.HasDifferences {
		fmt.Printf("%20s: %38s ≠ %-38s\n", "================= Id", entity.Id, other.Id)
		if diff.NameIsDifferent {
			fmt.Printf("%20s: %38s ≠ %-38s\n", "name", entity.Name, other.Name)
		}
		if diff.DataIsDifferent {
			for _, d := range diff.GetDataDifferences() {
				fmt.Print(d)
			}
		}
		if diff.MetaIsDifferent {
			for _, d := range diff.GetMetaDifferences() {
				fmt.Print(d)
			}
		}
		if diff.TagsIsDifferent {
			fmt.Println("tags differ")
		}
		if diff.SelectorsIsDifferent {
			fmt.Println("selectors differ")
		}
	}

}

// A EntityDiff compares two entities with each other, holds references to the
// compared entities and some info about where the differences are
// create with [sgjsonformat.NewEntityDiff]
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

func (diff *EntityDiff) GetDataDifferences() []string {
	return diff.a.Data.Differences(&diff.b.Data)
}

func (diff *EntityDiff) GetMetaDifferences() []string {
	return diff.a.Meta.Differences(&diff.b.Meta)
}

// Compares entity a with entity b and returns an [sgjsonformat.EntityDiff] reference
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

		NameIsDifferent:          a.Name != b.Name,
		DataIsDifferent:          !a.Data.Equal(&b.Data),
		MetaIsDifferent:          !a.Meta.Equal(&b.Meta),
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

// custom data type to accomodate json array & json object usage of data field
type EntityData []EntityDataEntry

func (data *EntityData) Differences(other *EntityData) []string {
	firstDataByKey := make(map[string]EntityDataEntry)
	secondDataByKey := make(map[string]EntityDataEntry)
	keys := make([]string, 0)

	for _, entry := range *data {
		keys = append(keys, entry.Name)
		firstDataByKey[entry.Name] = entry
	}
	for _, entry := range *other {
		keys = append(keys, entry.Name)
		secondDataByKey[entry.Name] = entry
	}
	keys = RemoveDuplicates(keys)

	diffs := make([]string, 0)
	for _, k := range keys {
		a := firstDataByKey[k]
		b := secondDataByKey[k]
		if a != b {
			if a.Type != b.Type {
				diff := fmt.Sprintf("%20s: %38v ≠ %-38v\n", "data."+k+".type", a.Type, b.Type)
				diffs = append(diffs, diff)
			}
			if a.Value != b.Value {
				diff := fmt.Sprintf("%20s: %38v ≠ %-38v\n", "data."+k+".value", a.Value, b.Value)
				diffs = append(diffs, diff)
			}
		}
	}
	return diffs
}

func (data *EntityData) Equal(other *EntityData) bool {
	if len(*data) != len(*other) {
		return false
	}

	firstDataByKey := make(map[string]EntityDataEntry)
	secondDataByKey := make(map[string]EntityDataEntry)
	keys := make([]string, 0)

	for _, entry := range *data {
		keys = append(keys, entry.Name)
		firstDataByKey[entry.Name] = entry
	}
	for _, entry := range *other {
		keys = append(keys, entry.Name)
		secondDataByKey[entry.Name] = entry
	}
	keys = RemoveDuplicates(keys)
	for _, k := range keys {
		if firstDataByKey[k] != secondDataByKey[k] {
			return false
		}
	}
	return true
}

// custom json unmarshaler for EntityData
// will always return a list of EntityDataEntries even when the data field in json
// is used as an object and not as an array.
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

type EntityDataEntry struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EntityDomain struct {
	Category string `json:"category"`
	Kind     string `json:"kind"`
}

type EntityMeta []EntityMetaEntry

func (meta *EntityMeta) Differences(other *EntityMeta) []string {
	firstMetaByKey := make(map[string]string)
	secondMetaByKey := make(map[string]string)
	keys := make([]string, 0)

	for _, entry := range *meta {
		keys = append(keys, entry.Kind)
		firstMetaByKey[entry.Kind] = entry.Value
	}
	for _, entry := range *other {
		keys = append(keys, entry.Kind)
		secondMetaByKey[entry.Kind] = entry.Value
	}
	keys = RemoveDuplicates(keys)
	diffs := make([]string, 0)
	for _, k := range keys {
		a := firstMetaByKey[k]
		b := secondMetaByKey[k]
		if a != b {
			diff := fmt.Sprintf("%20s: %38s ≠ %-38s\n", "meta."+k, a, b)
			diffs = append(diffs, diff)
		}
	}
	return diffs
}

func (meta *EntityMeta) Equal(other *EntityMeta) bool {
	if len(*meta) != len(*other) {
		return false
	}

	firstMetaByKey := make(map[string]string)
	secondMetaByKey := make(map[string]string)
	keys := make([]string, 0)

	for _, entry := range *meta {
		keys = append(keys, entry.Kind)
		firstMetaByKey[entry.Kind] = entry.Value
	}
	for _, entry := range *other {
		keys = append(keys, entry.Kind)
		secondMetaByKey[entry.Kind] = entry.Value
	}
	keys = RemoveDuplicates(keys)
	for _, k := range keys {
		if firstMetaByKey[k] != secondMetaByKey[k] {
			return false
		}
	}
	return true
}

type EntityMetaEntry struct {
	Value string `json:"value"`
	Kind  string `json:"kind"`
}
