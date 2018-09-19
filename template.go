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

func (i *Instruction) IsArray() bool {
	return i.Type == TypeString || i.Type == TypeSequence
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

func ParseXmlTemplate(reader io.Reader) ([]*Template, error) {
	return newXmlParser(reader).Parse()
}

func newXmlParser(reader io.Reader) *xmlParser {
	return &xmlParser{decoder: xml.NewDecoder(reader)}
}

func (p *xmlParser) Parse() (templates []*Template, err error) {
	for {
		token, err := p.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return templates, err
		}

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == tagTemplate{
			template, err := p.parseTemplate(&start)
			if err != nil {
				return templates, err
			}
			templates = append(templates, template)
		}
	}

	return templates, nil
}

func (p *xmlParser) parseTemplate(token *xml.StartElement) (template *Template, err error) {
	template, err = newTemplate(token)
	if err != nil {
		return
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return template, err
		}

		if start, ok := token.(xml.StartElement); ok {
			instruction, err := p.parseInstruction(&start)
			if err != nil {
				return template, err
			}
			template.Instructions = append(template.Instructions, instruction)
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return
}

func (p *xmlParser) parseInstruction(token *xml.StartElement) (instruction *Instruction, err error) {
	instruction, err = newInstruction(token)
	if err != nil {
		return
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return instruction, err
		}

		if start, ok := token.(xml.StartElement); ok {
			if instruction.Type == TypeSequence {

				inner, err := p.parseInstruction(&start)
				if err != nil {
					return instruction, err
				}
				instruction.Instructions = append(instruction.Instructions, inner)

			} else {
				instruction.Opt, instruction.Value = p.parseOption(&start, instruction.Type)
			}
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return
}

func (p *xmlParser) parseOption(token *xml.StartElement, typ InstructionType) (opt InstructionOpt, value interface{}) {
	switch token.Name.Local {
	case tagConstant:
		opt = OptConstant
		value = newValue(token, typ)
	case tagDefault:
		opt = OptDefault
		value = newValue(token, typ)
	case tagCopy:
		opt = OptCopy
	case tagDelta:
		opt = OptDelta
	case tagIncrement:
		opt = OptIncrement
	}

	for {
		token, err := p.decoder.Token()
		if err != nil {
			return opt, err
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return
}

func newInstruction(token *xml.StartElement) (*Instruction, error) {
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
	}

	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case attrName:
			instruction.Name = attr.Value
		case attrID:
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				return instruction, err
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
				return template, err
			}
			template.ID = uint(id)
		}
	}

	return template, nil
}

// TODO add other types
func newValue(token *xml.StartElement, typ InstructionType) interface{} {
	for _, attr := range token.Attr {
		if attr.Name.Local == attrValue {
			switch typ {
			case TypeString:
				return attr.Value
			}
		}
	}
	return nil
}
