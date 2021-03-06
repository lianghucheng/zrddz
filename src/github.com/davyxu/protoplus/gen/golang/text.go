package golang

// 报错行号+7
const TemplateText = `// Generated by github.com/davyxu/protoplus
// DO NOT EDIT!
package {{.PackageName}}

import (
	"github.com/davyxu/protoplus/proto"		
	"unsafe"
)
var (
	_ *proto.Buffer			
	_ unsafe.Pointer
)

{{range $a, $enumobj := .Enums}}
type {{.Name}} int32
const (	{{range .Fields}}
	{{$enumobj.Name}}_{{.Name}} {{$enumobj.Name}} = {{TagNumber $enumobj .}} {{end}}
)

var (
{{$enumobj.Name}}MapperValueByName = map[string]int32{ {{range .Fields}}
	"{{.Name}}": {{TagNumber $enumobj .}}, {{end}}
}

{{$enumobj.Name}}MapperNameByValue = map[int32]string{ {{range .Fields}}
	{{TagNumber $enumobj .}}: "{{.Name}}" , {{end}}
}
)

func (self {{$enumobj.Name}}) String() string {
	return {{$enumobj.Name}}MapperNameByValue[int32(self)]
}
{{end}}

{{range $a, $obj := .Structs}}
{{ObjectLeadingComment .}}
type {{.Name}} struct{	{{range .Fields}}
	{{GoFieldName .}} {{ProtoTypeName .}} {{GoStructTag .}}{{FieldTrailingComment .}} {{end}}
}

func (self *{{.Name}}) String() string { return proto.CompactTextString(self) }

func (self *{{.Name}}) Size() (ret int) {
{{range .Fields}}
	{{if IsStructSlice .}}
	if len(self.{{GoFieldName .}}) > 0 {
		for _, elm := range self.{{GoFieldName .}} {
			ret += proto.SizeStruct({{TagNumber $obj .}}, &elm)
		}
	}
	{{else if IsStruct .}}
	ret += proto.Size{{CodecName .}}({{TagNumber $obj .}}, &self.{{GoFieldName .}})
	{{else if IsEnum .}}
	ret += proto.Size{{CodecName .}}({{TagNumber $obj .}}, int32(self.{{GoFieldName .}}))
	{{else if IsEnumSlice .}}
	ret += proto.Size{{CodecName .}}({{TagNumber $obj .}}, *(*[]int32)(unsafe.Pointer(&self.{{GoFieldName .}})))
	{{else}}
	ret += proto.Size{{CodecName .}}({{TagNumber $obj .}}, self.{{GoFieldName .}})
	{{end}}
{{end}}
	return
}

func (self *{{.Name}}) Marshal(buffer *proto.Buffer) error {
{{range .Fields}}
	{{if IsStructSlice .}}
		for _, elm := range self.{{GoFieldName .}} {
			proto.MarshalStruct(buffer, {{TagNumber $obj .}}, &elm)
		}
	{{else if IsStruct .}}
		proto.Marshal{{CodecName .}}(buffer, {{TagNumber $obj .}}, &self.{{GoFieldName .}})
	{{else if IsEnum .}}
		proto.Marshal{{CodecName .}}(buffer, {{TagNumber $obj .}}, int32(self.{{GoFieldName .}}))
	{{else if IsEnumSlice .}}
		proto.Marshal{{CodecName .}}(buffer, {{TagNumber $obj .}}, *(*[]int32)(unsafe.Pointer(&self.{{GoFieldName .}})))
	{{else}}	
		proto.Marshal{{CodecName .}}(buffer, {{TagNumber $obj .}}, self.{{GoFieldName .}})
	{{end}}
{{end}}
	return nil
}

func (self *{{.Name}}) Unmarshal(buffer *proto.Buffer, fieldIndex uint64, wt proto.WireType) error {
	switch fieldIndex {
	{{range .Fields}} case {{TagNumber $obj .}}: {{if IsStructSlice .}}
		var elm {{.Type}}
		if err := proto.UnmarshalStruct(buffer, wt, &elm); err != nil {
			return err
		} else {
			self.{{GoFieldName .}} = append(self.{{GoFieldName .}}, elm)
			return nil
		}{{else if IsEnum .}}
		return proto.Unmarshal{{CodecName .}}(buffer, wt, (*int32)(&self.{{GoFieldName .}})) {{else if IsEnumSlice .}}
		return proto.Unmarshal{{CodecName .}}(buffer, wt, (*[]int32)(unsafe.Pointer(&self.{{GoFieldName .}}))) {{else}}
		return proto.Unmarshal{{CodecName .}}(buffer, wt, &self.{{GoFieldName .}}) {{end}}
		{{end}}
	}

	return proto.ErrUnknownField
}
{{end}}


`
