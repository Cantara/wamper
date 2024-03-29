// Code generated by templ@v0.2.364 DO NOT EDIT.

package main

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"net/url"
	"strings"
)

func image(site string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		if strings.TrimSpace(site) != "" {
			_, err = templBuffer.WriteString("<h2>")
			if err != nil {
				return err
			}
			var var_2 string = site
			_, err = templBuffer.WriteString(templ.EscapeString(var_2))
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString("</h2> <img id=\"image-result\" width=\"100%\" src=\"")
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString(templ.EscapeString("/site?name=" + url.QueryEscape(site)))
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString("\" alt=\"site image\">")
			if err != nil {
				return err
			}
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}
