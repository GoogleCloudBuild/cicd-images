// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
