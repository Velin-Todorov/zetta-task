package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("bookstore", func() {
	Title("Bookstore API")
	Description("An api that returns data about books in a bookstore")
	Version("1.0.0")
})

var _ = Service("books", func() {
	Description("Service that returns data about books")
	Error("not_found", String, "Book not found")
	Error("invalid_input", String, "Invalid input")
	Error("invalid_image_format", String, "Unsupported image format")
	Error("payload_too_large", String, "File size exceeds limit")
	Error("conflict", String, "Book already exists")
	Error("internal_error", String, "Internal server error")

	Method("getBooks", func() {
		Result(ArrayOf(Book))
		Payload(func() {
			Attribute("title", String, func() {
				MinLength(1)
			})
			Attribute("author", String, func() {
				MinLength(1)
			})
			Attribute("published_at", String, func() {
				Format(FormatDate)
			})
			Attribute("published_after", String, func() {
				Format(FormatDate)
			})
			Attribute("published_before", String, func() {
				Format(FormatDate)
			})
			Attribute("limit", UInt64)
			Attribute("offset", UInt64)
		})

		HTTP(func() {
			GET("/books")
			Param("title")
			Param("author")
			Param("published_at")
			Param("published_after")
			Param("published_before")
			Param("limit")
			Param("offset")
			Response(StatusOK, func() {
				ContentType("application/json")
			})
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("getBook", func() {
		Payload(func() {
			Attribute("id", Int64, "ID of the book")
			Required("id")
		})
		Result(Book)
		HTTP(func() {
			GET("/books/{id}")
			Response(StatusOK, func() {
				ContentType("application/json")
			})
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("createBook", func() {
		Payload(func() {
			Attribute("title", String, "Title of the book")
			Attribute("author", String, "Author of the book")
			Attribute("published_at", String, func() {
				Format(FormatDate)
				Description("Publication time of the book")
			})
			Required("title", "author", "published_at")
		})
		Result(Book)
		HTTP(func() {
			POST("/books")
			Response(StatusCreated, func() {
				ContentType("application/json")
			})
			Response("conflict", StatusConflict)
			Response("invalid_input", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("updateBook", func() {
		Payload(func() {
			Attribute("id", Int64, "ID of the book")
			Attribute("title", String, "Title of the book")
			Attribute("author", String, "Author of the book")
			Attribute("published_at", String, func() {
				Format(FormatDate)
				Description("Publication time of the book")
			})
			Required("id")
		})
		Result(Book)
		HTTP(func() {
			PATCH("/books/{id}")
			Response(StatusOK, func() {
				ContentType("application/json")
			})
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("invalid_input", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("setBookCover", func() {
		Payload(func() {
			Attribute("id", Int64, "ID of the book")
			Required("id")
		})
		Result(Book)
		HTTP(func() {
			PUT("/books/{id}/cover")
			SkipRequestBodyEncodeDecode()
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("invalid_image_format", StatusBadRequest)
			Response("payload_too_large", StatusRequestEntityTooLarge)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("deleteBook", func() {
		Payload(func() {
			Attribute("id", Int64, "ID of the book")
			Required("id")
		})
		HTTP(func() {
			DELETE("/books/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})

})
