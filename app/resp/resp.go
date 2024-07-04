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

func (rv RespValue) writeSimpleStrings(buf *bytes.Buffer) {
	buf.WriteString(rv.SimpleStr)
	buf.WriteString("\r\n")
}

func (rv RespValue) writeSimpleErrors(buf *bytes.Buffer) {
	buf.WriteString(rv.SimpleStr)
	buf.WriteString("\r\n")
}

func (rv RespValue) writeIntegers(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(rv.Int))
	buf.WriteString("\r\n")
}

func (rv RespValue) writeBulkStrings(buf *bytes.Buffer) {
	if rv.IsNullBulkStr {
		buf.WriteString(strconv.Itoa(-1))
	} else {
		buf.WriteString(strconv.Itoa(len(rv.BulkStr)))
		buf.WriteString("\r\n")
		buf.WriteString(rv.BulkStr)
	}
	buf.WriteString("\r\n")
}

func (rv RespValue) writeArrays(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(len(rv.Array)))
	buf.WriteString("\r\n")

	for _, arr := range rv.Array {
		arr.writeType(buf)
	}
}

func (rv RespValue) writeType(buf *bytes.Buffer) {
	buf.WriteString(rv.DataType)
	switch rv.DataType {
	case TypeSimpleErrors:
		rv.writeSimpleErrors(buf)
	case TypeSimpleStrings:
		rv.writeSimpleStrings(buf)
	case TypeIntegers:
		rv.writeIntegers(buf)
	case TypeBulkStrings:
		rv.writeBulkStrings(buf)
	case TypeArrays:
		rv.writeArrays(buf)
	}
}

func (rv RespValue) ToByteArray() []byte {
	var buf bytes.Buffer
	rv.writeType(&buf)
	return buf.Bytes()
}

func MakeInt(v int) *RespValue {
	return &RespValue{DataType: TypeIntegers, Int: v}
}

func MakeBulkString(msg string) *RespValue {
	return &RespValue{DataType: TypeBulkStrings, BulkStr: msg}
}

func MakeNilBulkString() *RespValue {
	return &RespValue{DataType: TypeBulkStrings, IsNullBulkStr: true}
}

func MakeErorr(msg string) *RespValue {
	return &RespValue{DataType: TypeSimpleErrors, SimpleStr: msg}
}
