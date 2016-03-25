// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"log"
	"net/http"

	"github.com/spf13/pflag"

	"encoding/json"

	"bytes"
	"strings"
	"text/template"
)

type StatusEndpoint struct {
	flagSet *pflag.FlagSet
}

func NewStatusEndpoint(flagSet *pflag.FlagSet) *StatusEndpoint {
	return &StatusEndpoint{flagSet: flagSet}
}

func (e *StatusEndpoint) ListFlags(resp http.ResponseWriter, req *http.Request) {
	onlyChanged := req.URL.Query().Get("only_changed") != ""
	onlyDynamic := req.URL.Query().Get("type") == "dynamic"
	onlyStatic := req.URL.Query().Get("type") == "static"

	listJson := &listJson{}
	e.flagSet.VisitAll(func(f *pflag.Flag) {
		if onlyChanged && !f.Changed {
			return
		}
		if onlyDynamic && !IsFlagDynamic(f) {
			return
		}
		if onlyStatic && IsFlagDynamic(f) {
			return
		}
		listJson.Flags = append(listJson.Flags, flagToJson(f))
	})

	if requestIsBrowser(req) && req.URL.Query().Get("format") != "json" {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Add("Content-Type", "text/html")
		if err := flagzListTemplate.Execute(resp, listJson); err != nil {
			log.Fatalf("Bad template evaluation: %v", err)
		}
	} else {
		resp.Header().Add("Content-Type", "application/json")
		out, err := json.MarshalIndent(&listJson, "", "  ")
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.WriteHeader(http.StatusOK)
		resp.Write(out)
	}
}

func requestIsBrowser(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept"), "html")
}

var (
	flagzListTemplate = template.Must(template.New("flagz_list").Parse(
		`
<html><head>
<title>Flagz List</title>
<link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.css" rel="stylesheet">

</head>
<body>
<div class="container-fluid">
<div class="col-md-10 col-md-offset-1">
	<h1>Flagz Debug View</h1>
	<p>
	This page presents the configuration flags of this server (<a href="?format=json">JSON</a>).
	</p>
	<p>
	You can easily filter only <a href="?only_changed=true"><span class="label label-primary">changed</span> flagz</a> or filter flags by type:
	<a href="?type=dynamic"><span class="label label-success">dynamic</span> flags</a> (ones that are tweakable by etcd)
	or <a href="?type=static"><span class="label label-default">static</span> ones</a>.
	</p>
	<p>


	{{range $flag := .Flags }}
		<div class="panel panel-default">
          <div class="panel-heading">
            <code>{{ $flag.Name }}</code>
            {{ if $flag.IsChanged }}<span class="label label-primary">changed</span>{{ end }}
            {{ if $flag.IsDynamic }}
                <span class="label label-success">dynamic</span>
            {{ else }}
                <span class="label label-default">static</span>
            {{ end }}

          </div>
		  <div class="panel-body">
		    <dl class="dl-horizontal" style="margin-bottom: 0px">
			  <dt>Description</dt>
			  <dd><small>{{ $flag.Description }}</small></dd>
			  <dt>Default</dt>
			  <dd><pre style="font-size: 8pt">{{ $flag.DefaultValue }}</pre></dd>
			  <dt>Current</dt>
			  <dd><pre class="success" style="font-size: 8pt">{{ $flag.CurrentValue }}</pre></dd>
		    </dl>
		  </div>
		</div>
	{{end}}
</div></div>
</body>
</html>
`))
)

type listJson struct {
	Flags []*flagJson `json:"flags"`
}

type flagJson struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	CurrentValue string `json:"current_value"`
	DefaultValue string `json:"default_value"`

	IsChanged bool `json:"is_changed"`
	IsDynamic bool `json:"is_dynamic"`
}

func flagToJson(f *pflag.Flag) *flagJson {
	fj := &flagJson{
		Name:         f.Name,
		Description:  f.Usage,
		CurrentValue: f.Value.String(),
		DefaultValue: f.DefValue,
		IsChanged:    f.Changed,
		IsDynamic:    IsFlagDynamic(f),
	}
	if strings.Contains(f.Value.Type(), "json") {
		fj.CurrentValue = prettyPrintJson(fj.CurrentValue)
		fj.DefaultValue = prettyPrintJson(fj.DefaultValue)
	}
	return fj
}

func prettyPrintJson(input string) string {
	out := &bytes.Buffer{}
	if err := json.Indent(out, []byte(input), "", "  "); err != nil {
		return "PRETTY_ERROR"
	}
	return out.String()
}
