package bpmn

import (
	"encoding/xml"
)

type elementType string

const (
	ElementTypeStart            elementType = "Start"
	ElementTypeEnd              elementType = "End"
	ElementTypeTask             elementType = "Task"
	ElementTypeExclusiveGateway elementType = "ExclusiveGateway"
	ElementTypeSequenceFlow     elementType = "SequenceFlow"
)

type Definition struct {
	XMLName         xml.Name    `xml:"definitions"`
	Text            string      `xml:",chardata"`
	ID              string      `xml:"id,attr"`
	TargetNamespace string      `xml:"targetNamespace,attr"`
	Exporter        string      `xml:"exporter,attr"`
	ExporterVersion string      `xml:"exporterVersion,attr"`
	Process         Process     `xml:"process"`
	BPMNDiagram     BPMNDiagram `xml:"BPMNDiagram"`
}

type process struct {
	Text              string             `xml:",chardata"`
	ID                string             `xml:"id,attr"`
	IsExecutable      string             `xml:"isExecutable,attr"`
	StartEvent        StartEvent         `xml:"startEvent"`
	ExclusiveGateways []ExclusiveGateway `xml:"exclusiveGateway"`
	SequenceFlows     []SequenceFlow     `xml:"sequenceFlow"`
	Tasks             []Task             `xml:"task"`
	EndEvents         []EndEvent         `xml:"endEvent"`
}

type Process struct {
	process
	currentElement Element
	elements       map[string]Element
	DAG            *Graph
}

func (p *Process) GetElement(name string) (Element, bool) {
	v, ok := p.elements[name]
	return v, ok
}

type StartEvent struct {
	Text     string `xml:",chardata"`
	ID       string `xml:"id,attr"`
	Outgoing string `xml:"outgoing"`
}

func (e StartEvent) Type() elementType {
	return ElementTypeStart
}

type EndEvent struct {
	Text     string `xml:",chardata"`
	ID       string `xml:"id,attr"`
	Incoming string `xml:"incoming"`
}

func (e EndEvent) Type() elementType {
	return ElementTypeEnd
}

type SequenceFlow struct {
	Text      string `xml:",chardata"`
	ID        string `xml:"id,attr"`
	SourceRef string `xml:"sourceRef,attr"`
	TargetRef string `xml:"targetRef,attr"`
}

func (e SequenceFlow) Type() elementType {
	return ElementTypeSequenceFlow
}

type Task struct {
	Text     string `xml:",chardata"`
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Incoming string `xml:"incoming"`
	Outgoing string `xml:"outgoing"`
}

func (e Task) Type() elementType {
	return ElementTypeTask
}

type ExclusiveGateway struct {
	Text     string   `xml:",chardata"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Incoming string   `xml:"incoming"`
	Outgoing []string `xml:"outgoing"`
}

func (e ExclusiveGateway) Type() elementType {
	return ElementTypeExclusiveGateway
}

// BPMNDiagram contains only spatial information for displaying a BPMN diagram
type BPMNDiagram struct {
	Text      string `xml:",chardata"`
	ID        string `xml:"id,attr"`
	BPMNPlane struct {
		Text        string `xml:",chardata"`
		ID          string `xml:"id,attr"`
		BpmnElement string `xml:"bpmnElement,attr"`
		BPMNShape   []struct {
			Text            string `xml:",chardata"`
			ID              string `xml:"id,attr"`
			BpmnElement     string `xml:"bpmnElement,attr"`
			IsMarkerVisible string `xml:"isMarkerVisible,attr"`
			Bounds          struct {
				Text   string `xml:",chardata"`
				X      string `xml:"x,attr"`
				Y      string `xml:"y,attr"`
				Width  string `xml:"width,attr"`
				Height string `xml:"height,attr"`
			} `xml:"Bounds"`
			BPMNLabel struct {
				Text   string `xml:",chardata"`
				Bounds struct {
					Text   string `xml:",chardata"`
					X      string `xml:"x,attr"`
					Y      string `xml:"y,attr"`
					Width  string `xml:"width,attr"`
					Height string `xml:"height,attr"`
				} `xml:"Bounds"`
			} `xml:"BPMNLabel"`
		} `xml:"BPMNShape"`
		BPMNEdge []struct {
			Text        string `xml:",chardata"`
			ID          string `xml:"id,attr"`
			BpmnElement string `xml:"bpmnElement,attr"`
			Waypoint    []struct {
				Text string `xml:",chardata"`
				X    string `xml:"x,attr"`
				Y    string `xml:"y,attr"`
			} `xml:"waypoint"`
		} `xml:"BPMNEdge"`
	} `xml:"BPMNPlane"`
}

type Element interface {
	Type() elementType
}

type Iterator interface {
	Next() []Element
}

func Unmarshal(data []byte) (*Definition, error) {
	def := &Definition{}
	if err := xml.Unmarshal(data, def); err != nil {
		return def, err
	}

	def.Process.elements = map[string]Element{
		def.Process.StartEvent.ID: def.Process.StartEvent,
	}

	def.Process.DAG = NewGraph()

	for _, elem := range def.Process.SequenceFlows {
		def.Process.elements[elem.ID] = elem
		def.Process.DAG.AddEdge(elem.SourceRef, elem.TargetRef)
	}

	for _, elem := range def.Process.Tasks {
		def.Process.elements[elem.ID] = elem
	}

	for _, elem := range def.Process.ExclusiveGateways {
		def.Process.elements[elem.ID] = elem
	}

	for _, elem := range def.Process.EndEvents {
		def.Process.elements[elem.ID] = elem
	}

	return def, nil
}
