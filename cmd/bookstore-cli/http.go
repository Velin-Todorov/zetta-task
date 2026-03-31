package main

import (
	"mime/multipart"
	"net/http"
	"time"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	cli "github.com/Velin-Todorov/zetta-task/gen/http/cli/bookstore"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, any, error) {
	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}

	return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		setBookCoverEncoderFunc,
	)
}

func httpUsageCommands() []string {
	return cli.UsageCommands()
}

func httpUsageExamples() string {
	return cli.UsageExamples()
}

// setBookCoverEncoderFunc encodes the cover image into a multipart form for the CLI client.
// This is the client-side counterpart of the server's decoder function.
func setBookCoverEncoderFunc(w *multipart.Writer, p *books.SetBookCoverPayload) error {
	part, err := w.CreateFormFile("cover", "cover")
	if err != nil {
		return err
	}
	_, err = part.Write(p.Cover)
	return err
}
