package book

import (
	"context"
	"strconv"
	"time"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
	"honnef.co/go/js/dom"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
)

//go:generate reactGen

// GetBookDef defines the getbook component
type GetBookDef struct {
	r.ComponentDef
	client library.BookServiceClient
}

// GetBookState holds the state for the GetBook component
type GetBookState struct {
	isbnInput string
	book      *library.Book
	err       string
}

// GetBook returns a new GetBookDef
func GetBook(client library.BookServiceClient) *GetBookDef {
	res := &GetBookDef{
		client: client,
	}
	r.BlessElement(res, nil)

	return res
}

// Render renders the GetBook component
func (g *GetBookDef) Render() r.Element {
	st := g.State()
	content := []r.Element{
		r.P(nil, r.S("Search for book by ISBN (for example, 140008381).")),
		r.Form(&r.FormProps{ClassName: "form-inline"},
			r.Div(
				&r.DivProps{ClassName: "form-group"},
				r.Label(&r.LabelProps{ClassName: "sr-only", For: "isnbText"}, r.S("ISBN")),
				r.Input(&r.InputProps{
					Type:      "number",
					ClassName: "form-control",
					ID:        "isnbText",
					Value:     st.isbnInput,
					OnChange:  isbnInputChange{g},
				}),
				r.Button(&r.ButtonProps{
					Type:      "submit",
					ClassName: "btn btn-default",
					OnClick:   triggerGet{g},
				}, r.S("Get Book")),
			),
		),
	}

	if st.book != nil {
		content = append(content, renderBook(st.book))
	}

	if st.err != "" {
		content = append(content,
			r.Div(nil,
				r.HR(nil),
				r.S("Error: "+st.err),
			),
		)
	}

	return r.Div(nil, content...)
}

type isbnInputChange struct{ g *GetBookDef }
type triggerGet struct{ g *GetBookDef }

func (i isbnInputChange) OnChange(se *r.SyntheticEvent) {
	target := se.Target().(*dom.HTMLInputElement)

	newSt := i.g.State()
	newSt.isbnInput = target.Value

	i.g.SetState(newSt)
}

func (t triggerGet) OnClick(se *r.SyntheticMouseEvent) {
	// Wrapped in goroutine because GetBook is blocking
	go func() {
		newSt := t.g.State()
		// Note: defer t.g.SetState(newSt) doesn't work for some reason?
		newSt.err = ""
		newSt.book = nil

		isbn, err := strconv.Atoi(newSt.isbnInput)
		if err != nil {
			newSt.err = "ISBN must not be empty"
			t.g.SetState(newSt)
			return
		}

		// 1 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		bk, err := t.g.client.GetBook(ctx, new(library.GetBookRequest).New(int64(isbn)))
		if err != nil {
			sts := status.FromError(err)
			newSt.err = sts.Message
			t.g.SetState(newSt)
			return
		}

		newSt.book = bk
		t.g.SetState(newSt)
	}()

	se.PreventDefault()
}
