package fast

import (
	"encoding/xml"
	"io"
	"strconv"
)

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
	for i:=0; i<12; i++ { // TODO
		token, err := p.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return templates, err
		}

		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "template"{
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
	template = &Template{}

	for _, attr := range token.Attr {
		if attr.Name.Local == "name" {
			template.Name = attr.Value
		}
		if attr.Name.Local == "id" {
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				return template, err
			}
			template.ID = uint(id)
		}
	}

	for i:=0; i<12; i++ { // TODO
		token, err := p.decoder.Token()
		if err != nil {
			return template, err
		}

		if start, ok := token.(xml.StartElement); ok {
			instruction, err := p.parseInstraction(&start)
			if err != nil {
				return template, err
			}
			template.Instructions = append(template.Instructions, instruction)
		}

		if _, ok := token.(xml.EndElement); ok {
			break
		}
	}

	return template, nil
}

func (p *xmlParser) parseInstraction(token *xml.StartElement) (instraction *Instruction, err error) {
	return nil, nil // TODO
}
