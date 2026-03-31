package design

import (
	. "goa.design/goa/v3/dsl"
	cors "goa.design/plugins/v3/cors/dsl"
)

var _ = API("bookstore", func() {
	Title("Bookstore API")
	Description("An api that returns data about books in a bookstore")
	Version("1.0.0")
	Server("bookstore", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("books", func() {
	Description("Service that returns data about books")
	Error("not_found")
	Error("invalid_input")
	Error("invalid_image_format")
	Error("payload_too_large")
	Error("conflict")
	Error("internal_error")

	// CORS is required for Swagger UI to make API calls from the browser.
	cors.Origin("*", func() {
		cors.Methods("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS")
		cors.Headers("Content-Type")
	})

	Method("getBooks", func() {
		Result(ArrayOf(Book))
		Payload(func() {
			Attribute("title", String, func() {
				MinLength(1)
			})
			Attribute("author", String, func() {
				MinLength(1)
			})
			Attribute("publishedAt", String, func() {
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
			Param("publishedAt")
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
			Attribute("publishedAt", String, func() {
				Format(FormatDate)
				Description("Publication time of the book")
			})
			Required("title", "author", "publishedAt")
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
			Attribute("publishedAt", String, func() {
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
			Attribute("cover", Bytes, "Book cover image")
			Required("id", "cover")
		})
		Result(Book)
		HTTP(func() {
			PUT("/books/{id}/cover")
			MultipartRequest()
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

	Files("/uploads/covers/{*path}", "./uploads/covers")
	Files("/openapi.json", "./gen/http/openapi3.json")
	Files("/swagger", "./public/swagger.html")
})
