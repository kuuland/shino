package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type meta struct {
	Name        string
	DisplayName string
	Fields      []field
}

type field struct {
	Code     string
	Name     string
	Kind     string
	Type     string
	Enum     string
	IsRef    bool
	IsArray  bool
	FormType string
}

func parse(v string, r interface{}) error {
	return json.Unmarshal([]byte(v), r)
}

func fetchMeta(url string) (result []meta) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("fetching meta:", err)
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading body:", err)
		return
	}
	ret := struct {
		Code int
		Msg  string
		Data []meta
	}{}
	if err := parse(string(b), &ret); err != nil {
		log.Println("parsing body:", err)
		return
	}
	for i, meta := range ret.Data {
		for j, field := range meta.Fields {
			switch field.Type {
			case "boolean":
				field.FormType = "switch"
			case "integer", "number":
				field.FormType = "number"
			case "string":
				field.FormType = "input"
			}
			meta.Fields[j] = field
		}
		ret.Data[i] = meta
	}

	return ret.Data
}

func tableTmpl() string {
	return `
import React from 'react'
import moment from 'moment'
import { FanoTable } from 'fano-antd'

// {{.DisplayName}}
export default class {{.Name}}Table extends React.Component {
  constructor (props) {
    super(props)

    this.state = {
      columns: [
    {{range $i, $v := .Fields}}
      {{ if ne $v.Name "" }}
        {
          title: '{{$v.Name}}',
          dataIndex: '{{$v.Code}}'
        }{{ if notLastField $i (len $.Fields) }},{{ end }}
      {{ end }}
    {{ end }}
      ],
      form: [
    {{range $i, $v := .Fields}}
      {{ if ne $v.Name "" }}
        {
          name: '{{$v.Code}}',
          type: '{{$v.FormType}}',
          label: '{{$v.Name}}',
          props: {
            {{ if eq $v.Type "integer" }}precision: 0,{{ end }}
          }
        }{{ if notLastField $i (len $.Fields) }},{{ end }}
      {{ end }}
    {{ end }}
      ]
    }
  }

  render () {
    const { columns, form } = this.state
    return <FanoTable columns={columns} form={form} url={'/api/{{toLower .Name}}'} />
  }
}`
}

func fanoTable(metaURL string, outDir string) {
	list := fetchMeta(metaURL)
	tmpl := tableTmpl()
	t := template.Must(template.New("table").Funcs(template.FuncMap{
		"notLastField": func(index int, length int) bool {
			return index < (length - 1)
		},
		"toLower": strings.ToLower,
	}).Parse(tmpl))
	ensureDir(outDir)
	for _, meta := range list {
		f, err := os.Create(fmt.Sprintf("%s/%s.jsx", outDir, meta.Name))
		if err = t.Execute(f, meta); err != nil {
			log.Println("executing template:", err)
		}
		if err := f.Close(); err != nil {
			log.Println("out to file:", err)
		}
	}
}
