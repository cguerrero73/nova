package fields

import "time"

// Field represents a field definition from eamfields
type Field struct {
	ID        int       `json:"fld_id" db:"fld_id"`
	FieldName string    `json:"fld_fieldname" db:"fld_fieldname"`
	DomainKey string    `json:"fld_domain_key" db:"fld_domain_key"`
	DataType  string    `json:"fld_datatype" db:"fld_datatype"`
	TableName string    `json:"fld_tablename" db:"fld_tablename"`
	CreatedAt time.Time `json:"fld_created_at" db:"fld_created_at"`
	UpdatedAt time.Time `json:"fld_updated_at" db:"fld_updated_at"`
}

// DomainKeyOrName returns the domain key if set, otherwise the raw field name.
func (f *Field) DomainKeyOrName() string {
	if f.DomainKey != "" {
		return f.DomainKey
	}
	return f.FieldName
}

// DataType constants
const (
	DATATYPE_STRING  = "string"
	DATATYPE_NUMBER  = "number"
	DATATYPE_DATE    = "date"
	DATATYPE_BOOLEAN = "boolean"
	DATATYPE_SELECT  = "select"
)
