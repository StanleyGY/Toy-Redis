package resp

import (
	"bytes"
	"strconv"
)

const (
	TypeSimpleStrings = "+"
	TypeSimpleErrors  = "-"
	TypeIntegers      = ":"
	TypeBulkStrings   = "$"
	TypeArrays        = "*"
)

type RespValue struct {
	DataType      string
	SimpleStr     string
	BulkStr       string
	IsNullBulkStr bool
	Int           int
	Array         []*RespValue
}

func (rv RespValue) ToByteArray() []byte {
	var buf bytes.Buffer
	buf.WriteString(rv.DataType)

	switch rv.DataType {
	case TypeSimpleStrings:
		buf.WriteString(rv.SimpleStr)
	case TypeBulkStrings:
		if rv.IsNullBulkStr {
			buf.WriteString(strconv.Itoa(-1))
		} else {
			buf.WriteString(strconv.Itoa(len(rv.BulkStr)))
			buf.WriteString("\r\n")
			buf.WriteString(rv.BulkStr)
		}
	}
	buf.WriteString("\r\n")
	return buf.Bytes()
}
