package fast

import (
	"encoding/xml"
	"io"
	"strconv"
)

const (
	tagTemplate = "template"

	tagString = "string"
	tagInt32 = "int32"
	tagUint32 = "uInt32"
	tagInt64 = "int64"
	tagUint64 = "uInt64"
	tagDecimal = "decimal"
	tagSequence = "sequence"
	tagLength = "length"
	tagExponent = "exponent"
	tagMantissa = "mantissa"

	tagIncrement = "increment"
	tagConstant = "constant"
	tagDefault = "default"
	tagCopy = "copy"
	tagDelta = "delta"
	tagTail = "tail"

	attrID = "id"
	attrName = "name"
	attrPresence = "presence"
	attrValue = "value"

	valueMandatory = "mandatory"
	valueOptional = "optional"
)

type InstructionType int
type InstructionOpt int
type InstructionPresence int

const (
	TypeNull InstructionType = iota
	TypeUint32
	TypeInt32
	TypeUint64
	TypeInt64
	TypeString
	TypeSequence
	TypeDecimal
	TypeLength
	TypeExponent
	TypeMantissa

	OptNone InstructionOpt = iota
	OptConstant
	OptDelta
	OptDefault
	OptCopy
	OptIncrement
	OptTail

	PresenceMandatory InstructionPresence = iota
	PresenceOptional
)

type Instruction struct {
	ID uint
	Name string
	Presence InstructionPresence
	Type InstructionType
	Opt InstructionOpt
	Instructions []*Instruction
	Value interface{}
}

func (i *Instruction) IsOptional() bool {
	return i.Presence == PresenceOptional
}

func (i *Instruction) IsNullable() bool {
	return i.IsOptional() && (i.Opt != OptConstant)
}

func (i *Instruction) HasPmapBit() bool {
	return i.Opt > OptDelta || ((i.Opt == OptConstant) && i.Presence == PresenceOptional)
}

type Template struct {
	ID uint
	Name string
	Instructions []*Instruction
}

// --------------------------------------------------------

type xmlParser struct {
	decoder *xml.Decoder
}

func ParseXmlTemplate(reader io.Reader) []*Template {
	return newXmlParser(reader).Parse()
}

func newXmlParser(reader io.Reader) *xmlParser {
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

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == tagTemplate{
			template := p.parseTemplate(&start)
			templates = append(templates, template)
		}
	}

	return templates
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
			} else if instruction.Type == TypeDecimal {
				inner := p.parseInstruction(&start)
				inner.ID = instruction.ID
				inner.Name = instruction.Name

				// If the decimal has optional presence, the exponent field is treated as on optional
				//  integer field and the mantissa field is treated as a mandatory integer field.
				if inner.Type == TypeExponent && instruction.Presence == PresenceOptional {
					inner.Presence = instruction.Presence
				}
				instruction.Instructions = append(instruction.Instructions, inner)
			} else {
				instruction.Opt, instruction.Value = p.parseOperation(&start, instruction.Type)
			}
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return instruction
}

func (p *xmlParser) parseOperation(token *xml.StartElement, typ InstructionType) (opt InstructionOpt, value interface{}) {
	switch token.Name.Local {
	case tagConstant:
		opt = OptConstant
	case tagDefault:
		opt = OptDefault
	case tagCopy:
		opt = OptCopy
	case tagDelta:
		opt = OptDelta
	case tagIncrement:
		opt = OptIncrement
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
	instruction := &Instruction{Opt: OptNone}

	switch token.Name.Local {
	case tagString:
		instruction.Type = TypeString
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
			case TypeString:
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
