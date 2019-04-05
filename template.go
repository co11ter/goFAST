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
	tagGroup      = "group"
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
	TypeLength
	TypeExponent
	TypeMantissa
	TypeDecimal
	TypeAsciiString
	TypeUnicodeString
	TypeByteVector
	TypeSequence
	TypeGroup

	OperatorNone InstructionOperator = iota
	OperatorConstant
	OperatorDelta
	OperatorDefault
	OperatorCopy
	OperatorIncrement // It's applicable to integer field types
	OperatorTail // It's applicable to string and byte vector field types

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
func ParseXMLTemplate(reader io.Reader) ([]*Template, error) {
	return newXMLParser(reader).Parse()
}

func newXMLParser(reader io.Reader) *xmlParser {
	return &xmlParser{decoder: xml.NewDecoder(reader)}
}

func (p *xmlParser) Parse() (templates []*Template, err error) {
	var token xml.Token
	var template *Template
	for {
		token, err = p.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == tagTemplate {
			template, err = p.parseTemplate(&start)
			if err != nil {
				return
			}
			templates = append(templates, template)
		}
	}

	for _, tpl := range templates {
		err = p.postProcessing(tpl.Instructions)
		if err != nil {
			break
		}
	}

	return
}

func (p *xmlParser) postProcessing(instructions []*Instruction) (err error) {
	for _, item := range instructions {
		if !item.isValid() {
			return ErrS2
		}

		item.key = strconv.Itoa(int(item.ID)) + ":" +
			item.Name + ":" +
			strconv.Itoa(int(item.Type))

		err = p.postProcessing(item.Instructions)
		if err != nil {
			return err
		}

		if item.Type != TypeSequence && item.Type != TypeGroup {
			continue
		}

		for _, instruction := range item.Instructions {
			if instruction.hasPmapBit() {
				item.pMapSize++
			}
		}
	}

	return
}

func (p *xmlParser) parseTemplate(token *xml.StartElement) (*Template, error) {
	template, err := newTemplate(token)
	if err != nil {
		return nil, err
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return nil, err
		}

		if start, ok := token.(xml.StartElement); ok {
			instruction, err := p.parseInstruction(&start)
			if err != nil {
				return nil, err
			}
			template.Instructions = append(template.Instructions, instruction)
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return template, nil
}

func (p *xmlParser) parseDecimalInstructionOrOperator(token *xml.StartElement, instruction *Instruction) error {
	inner, err := newInstruction(token)
	if err != nil {
		return err
	}

	if inner.Type != TypeNull {
		inner, err := p.parseInstruction(token)
		if err != nil {
			return err
		}
		inner.ID = instruction.ID
		inner.Name = instruction.Name

		// If the decimal has optional presence, the exponent field is treated as on optional
		//  integer field and the mantissa field is treated as a mandatory integer field.
		if inner.Type == TypeExponent && instruction.Presence == PresenceOptional {
			inner.Presence = instruction.Presence
		}
		instruction.Instructions = append(instruction.Instructions, inner)
	} else {
		err = p.parseOperation(token, instruction)
	}

	return err
}

func (p *xmlParser) parseInstruction(token *xml.StartElement) (*Instruction, error) {
	instruction, err := newInstruction(token)
	if err != nil {
		return nil, err
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return nil, err
		}

		if start, ok := token.(xml.StartElement); ok {
			switch instruction.Type {
			case TypeSequence, TypeGroup:
				inner, err := p.parseInstruction(&start)
				if err != nil {
					return nil, err
				}
				instruction.Instructions = append(instruction.Instructions, inner)
				if inner.Type == TypeLength && instruction.Presence == PresenceOptional {
					inner.Presence = PresenceOptional
				}
			case TypeDecimal:
				err = p.parseDecimalInstructionOrOperator(&start, instruction)
			default:
				err = p.parseOperation(&start, instruction)
			}
		}

		if err != nil {
			return nil, err
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return instruction, nil
}

func (p *xmlParser) parseOperation(token *xml.StartElement, instruction *Instruction) error {
	switch token.Name.Local {
	case tagConstant:
		instruction.Operator = OperatorConstant
	case tagDefault:
		instruction.Operator = OperatorDefault
	case tagCopy:
		instruction.Operator = OperatorCopy
	case tagDelta:
		instruction.Operator = OperatorDelta
	case tagIncrement:
		instruction.Operator = OperatorIncrement
	default:
		instruction.Operator = OperatorNone
	}

	var err error
	instruction.Value, err = newValue(token, instruction.Type)
	if err != nil {
		return err
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return err
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return nil
}

func newInstruction(token *xml.StartElement) (*Instruction, error) {
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
	case tagGroup:
		instruction.Type = TypeGroup
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
				return nil, err
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

	return instruction, nil
}

func newTemplate(token *xml.StartElement) (*Template, error) {
	template := &Template{}

	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case attrName:
			template.Name = attr.Value
		case attrID:
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				return nil, err
			}
			template.ID = uint(id)
		}
	}

	return template, nil
}

func newValue(token *xml.StartElement, typ InstructionType) (value interface{}, err error) {
	for _, attr := range token.Attr {
		if attr.Name.Local == attrValue {
			switch typ {
			case TypeAsciiString, TypeUnicodeString:
				value = attr.Value
			case TypeUint64:
				value, err = strconv.ParseUint(attr.Value, 10, 64)
			case TypeUint32:
				value, err = strconv.ParseUint(attr.Value, 10, 32)
				value = uint32(value.(uint64))
			case TypeInt64, TypeMantissa:
				value, err = strconv.ParseInt(attr.Value, 10, 64)
			case TypeInt32, TypeExponent:
				value, err = strconv.ParseInt(attr.Value, 10, 32)
				value = int32(value.(int64))
			}
			return
		}
	}
	return
}
