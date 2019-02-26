// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"encoding/xml"
	"io"
	"strconv"
)

const (
	tagTemplate = "template"

	tagString     = "string"
	tagInt32      = "int32"
	tagUint32     = "uInt32"
	tagInt64      = "int64"
	tagUint64     = "uInt64"
	tagDecimal    = "decimal"
	tagSequence   = "sequence"
	tagLength     = "length"
	tagExponent   = "exponent"
	tagMantissa   = "mantissa"
	tagByteVector = "byteVector"

	tagIncrement = "increment"
	tagConstant  = "constant"
	tagDefault   = "default"
	tagCopy      = "copy"
	tagDelta     = "delta"
	tagTail      = "tail"

	attrID       = "id"
	attrName     = "name"
	attrPresence = "presence"
	attrValue    = "value"
	attrCharset  = "charset"

	valueMandatory = "mandatory"
	valueOptional  = "optional"
	valueUnicode   = "unicode"
)

// InstructionType specifies the basic encoding of the field.
type InstructionType int

// InstructionOperator specifies ways to optimize the encoding of the field.
type InstructionOperator int

// InstructionPresence specifies presence of the field.
type InstructionPresence int

const (
	// TypeNull mark type of instruction null.
	TypeNull InstructionType = iota
	TypeUint32
	TypeInt32
	TypeUint64
	TypeInt64
	TypeAsciiString
	TypeUnicodeString
	TypeSequence
	TypeDecimal
	TypeLength
	TypeExponent
	TypeMantissa
	TypeByteVector

	OperatorNone InstructionOperator = iota
	OperatorConstant
	OperatorDelta
	OperatorDefault
	OperatorCopy
	OperatorIncrement
	OperatorTail

	PresenceMandatory InstructionPresence = iota
	PresenceOptional
)

// Template collect instructions for this template
type Template struct {
	ID           uint
	Name         string
	Instructions []*Instruction
}

type xmlParser struct {
	decoder *xml.Decoder
}

// ParseXMLTemplate reads xml data from reader and return templates collection.
func ParseXMLTemplate(reader io.Reader) []*Template {
	return newXMLParser(reader).Parse()
}

func newXMLParser(reader io.Reader) *xmlParser {
	return &xmlParser{decoder: xml.NewDecoder(reader)}
}

func (p *xmlParser) Parse() (templates []*Template) {
	for {
		token, err := p.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == tagTemplate {
			template := p.parseTemplate(&start)
			templates = append(templates, template)
		}
	}

	for _, tpl := range templates {
		p.postProcessing(tpl.Instructions)
	}

	return templates
}

func (p *xmlParser) postProcessing(instructions []*Instruction) {
	for _, item := range instructions {
		if item.Type != TypeSequence {
			continue
		}

		for _, instruction := range item.Instructions {
			if instruction.hasPmapBit() {
				item.pMapSize++
			}
		}

		p.postProcessing(item.Instructions)
	}
}

func (p *xmlParser) parseTemplate(token *xml.StartElement) *Template {
	template := newTemplate(token)

	for {
		token, err := p.decoder.Token()
		if err != nil {
			panic(err)
		}

		if start, ok := token.(xml.StartElement); ok {
			instruction := p.parseInstruction(&start)
			template.Instructions = append(template.Instructions, instruction)
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return template
}

func (p *xmlParser) parseDecimalInstructionOrOperator(token *xml.StartElement, instruction *Instruction) {
	inner := newInstruction(token)
	if inner.Type != TypeNull {
		inner = p.parseInstruction(token)
		inner.ID = instruction.ID
		inner.Name = instruction.Name

		// If the decimal has optional presence, the exponent field is treated as on optional
		//  integer field and the mantissa field is treated as a mandatory integer field.
		if inner.Type == TypeExponent && instruction.Presence == PresenceOptional {
			inner.Presence = instruction.Presence
		}
		instruction.Instructions = append(instruction.Instructions, inner)
	} else {
		instruction.Operator, instruction.Value = p.parseOperation(token, instruction.Type)
	}
}

func (p *xmlParser) parseInstruction(token *xml.StartElement) *Instruction {
	instruction := newInstruction(token)

	for {
		token, err := p.decoder.Token()
		if err != nil {
			panic(err)
		}

		if start, ok := token.(xml.StartElement); ok {
			if instruction.Type == TypeSequence {
				inner := p.parseInstruction(&start)
				instruction.Instructions = append(instruction.Instructions, inner)
				if inner.Type == TypeLength && instruction.Presence == PresenceOptional {
					inner.Presence = PresenceOptional
				}
			} else if instruction.Type == TypeDecimal {
				p.parseDecimalInstructionOrOperator(&start, instruction)
			} else {
				instruction.Operator, instruction.Value = p.parseOperation(&start, instruction.Type)
			}
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return instruction
}

func (p *xmlParser) parseOperation(token *xml.StartElement, typ InstructionType) (opt InstructionOperator, value interface{}) {
	switch token.Name.Local {
	case tagConstant:
		opt = OperatorConstant
	case tagDefault:
		opt = OperatorDefault
	case tagCopy:
		opt = OperatorCopy
	case tagDelta:
		opt = OperatorDelta
	case tagIncrement:
		opt = OperatorIncrement
	default:
		opt = OperatorNone
	}

	value = newValue(token, typ)

	for {
		token, err := p.decoder.Token()
		if err != nil {
			panic(err)
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return
}

func newInstruction(token *xml.StartElement) *Instruction {
	instruction := &Instruction{Operator: OperatorNone}

	switch token.Name.Local {
	case tagString:
		instruction.Type = TypeAsciiString
	case tagInt32:
		instruction.Type = TypeInt32
	case tagInt64:
		instruction.Type = TypeInt64
	case tagUint32:
		instruction.Type = TypeUint32
	case tagUint64:
		instruction.Type = TypeUint64
	case tagDecimal:
		instruction.Type = TypeDecimal
	case tagSequence:
		instruction.Type = TypeSequence
	case tagLength:
		instruction.Type = TypeLength
	case tagExponent:
		instruction.Type = TypeExponent
	case tagMantissa:
		instruction.Type = TypeMantissa
	case tagByteVector:
		instruction.Type = TypeByteVector
	default:
		instruction.Type = TypeNull
	}

	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case attrName:
			instruction.Name = attr.Value
		case attrID:
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				panic(err)
			}
			instruction.ID = uint(id)
		case attrPresence:
			if attr.Value == valueMandatory {
				instruction.Presence = PresenceMandatory
			}
			if attr.Value == valueOptional {
				instruction.Presence = PresenceOptional
			}
		case attrCharset:
			if attr.Value == valueUnicode {
				instruction.Type = TypeUnicodeString
			}
		}
	}

	return instruction
}

func newTemplate(token *xml.StartElement) *Template {
	template := &Template{}

	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case attrName:
			template.Name = attr.Value
		case attrID:
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				panic(err)
			}
			template.ID = uint(id)
		}
	}

	return template
}

func newValue(token *xml.StartElement, typ InstructionType) interface{} {
	for _, attr := range token.Attr {
		if attr.Name.Local == attrValue {
			switch typ {
			case TypeAsciiString, TypeUnicodeString:
				return attr.Value
			case TypeUint64:
				value, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					panic(err)
				}
				return value
			case TypeUint32:
				value, err := strconv.ParseUint(attr.Value, 10, 32)
				if err != nil {
					panic(err)
				}
				return uint32(value)
			case TypeInt64, TypeMantissa:
				value, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					panic(err)
				}
				return value
			case TypeInt32, TypeExponent:
				value, err := strconv.ParseInt(attr.Value, 10, 32)
				if err != nil {
					panic(err)
				}
				return int32(value)
			}
		}
	}
	return nil
}
