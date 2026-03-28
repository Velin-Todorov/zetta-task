package design

import (
	. "goa.design/goa/v3/dsl"
)

var Book = Type("Book", func() {
	Description("A single book")

	Attribute("id", Int64)
	Attribute("title", String)
	Attribute("author", String)
	Attribute("cover_url", String)
	Attribute("published_at", String, func() {
		Format(FormatDate)
	})

	Required("id", "title", "author", "published_at")
})
