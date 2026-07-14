// Package httpx, handler'larda tekrar eden küçük HTTP yardımcıları.
package httpx

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

// Render, bir templ component'ini response'a yazar.
func Render(c fiber.Ctx, component templ.Component) error {
	c.Type("html", "utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// IsHTMX, isteğin htmx'ten gelip gelmediğini söyler.
func IsHTMX(c fiber.Ctx) bool {
	return c.Get("HX-Request") == "true"
}

// ListQuery, listing sayfalarının URL query string'ini kurar: verilen sayfa
// numarası + istekteki arama terimi ("q", boşsa atlanır). Handler'lar form
// action ve fragment URL'lerinde aramanın kaybolmaması için kullanır.
func ListQuery(c fiber.Ctx, page int) string {
	v := url.Values{}
	v.Set("page", strconv.Itoa(page))
	if q := strings.TrimSpace(fiber.Query[string](c, "q")); q != "" {
		v.Set("q", q)
	}
	return "?" + v.Encode()
}

// Redirect, htmx isteklerinde HX-Redirect header'ı ile,
// normal isteklerde 303 See Other ile yönlendirir.
func Redirect(c fiber.Ctx, url string) error {
	if IsHTMX(c) {
		c.Set("HX-Redirect", url)
		return c.SendString("")
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To(url)
}
