package xmlmodules

import "encoding/xml"

type Settings struct {
	XMLName         xml.Name        `xml:"settings"`
	LocalRepository LocalRepository `xml:"localRepository"`
	Servers         Servers         `xml:"servers"`
}

type LocalRepository struct {
	Path string `xml:",chardata"`
}

type Servers struct {
	Server []Server `xml:"server"`
}

type Server struct {
	ID            string        `xml:"id"`
	Configuration Configuration `xml:"configuration"`
	Username      string        `xml:"username"`
	Password      string        `xml:"password"`
}

type Configuration struct {
	HttpConfiguration HttpConfiguration `xml:"httpConfiguration"`
}

type HttpConfiguration struct {
	Get  bool      `xml:"get>usePreemptive"`
	Head bool      `xml:"head>usePreemptive"`
	Put  PutParams `xml:"put>params"`
}

type UsePreemptive struct {
	Value bool `xml:"value"`
}

type PutParams struct {
	Property []Property `xml:"property"`
}

type Property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}
