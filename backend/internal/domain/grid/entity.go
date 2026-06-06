package grid

import (
	"strconv"
	"strings"
	"time"
)

// Grid represents a grid configuration from eamgrids
type Grid struct {
	ID          int    `json:"grd_id" db:"grd_id"`
	Name        string `json:"grd_name" db:"grd_name"`
	Description string `json:"grd_desc" db:"grd_desc"`

	BaseQuery string `json:"grd_base_query" db:"grd_base_query"`

	KeyFields       string `json:"grd_key_fields" db:"grd_key_fields"`
	FilterableList  string `json:"grd_filterable_list" db:"grd_filterable_list"`
	SortableList    string `json:"grd_sortable_list" db:"grd_sortable_list"`
	DisplayableList string `json:"grd_displayable_list" db:"grd_displayable_list"`

	OrgColumn   string `json:"grd_org_column" db:"grd_org_column"`
	BotFunction string `json:"grd_bot_function" db:"grd_bot_function"`
	SecEntity   string `json:"grd_sec_entity" db:"grd_sec_entity"`
	Hints       string `json:"grd_hints" db:"grd_hints"`
	GridType    int    `json:"grd_type" db:"grd_type"`

	CreatedAt time.Time `json:"grd_created_at" db:"grd_created_at"`
	UpdatedAt time.Time `json:"grd_updated_at" db:"grd_updated_at"`
}

// ParseCSV converts comma-separated string to int slice
// Example: "1,3,5" -> [1, 3, 5]
func (g *Grid) ParseCSV(list string) []int {
	if list == "" {
		return nil
	}
	parts := strings.Split(list, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.Atoi(p); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetKeyFields returns key fields as int slice
func (g *Grid) GetKeyFields() []int {
	return g.ParseCSV(g.KeyFields)
}

// GetFilterableFields returns filterable fields as int slice
func (g *Grid) GetFilterableFields() []int {
	return g.ParseCSV(g.FilterableList)
}

// GetSortableFields returns sortable fields as int slice
func (g *Grid) GetSortableFields() []int {
	return g.ParseCSV(g.SortableList)
}

// GetDisplayableFields returns displayable fields as int slice
func (g *Grid) GetDisplayableFields() []int {
	return g.ParseCSV(g.DisplayableList)
}

// IsFieldFilterable checks if a field ID is filterable
func (g *Grid) IsFieldFilterable(fieldID int) bool {
	filterable := g.GetFilterableFields()
	for _, f := range filterable {
		if f == fieldID {
			return true
		}
	}
	return false
}

// IsFieldSortable checks if a field ID is sortable
func (g *Grid) IsFieldSortable(fieldID int) bool {
	sortable := g.GetSortableFields()
	for _, s := range sortable {
		if s == fieldID {
			return true
		}
	}
	return false
}

// IsFieldDisplayable checks if a field ID is displayable
func (g *Grid) IsFieldDisplayable(fieldID int) bool {
	displayable := g.GetDisplayableFields()
	for _, d := range displayable {
		if d == fieldID {
			return true
		}
	}
	return false
}
