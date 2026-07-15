package tag

import (
	"errors"
	"fmt"
	"strings"

	"ecommerce/internal/admin/views"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/shared/icons"

	"github.com/gofiber/fiber/v3"
)

const (
	tagTitle    = "Etiketler"
	tagSingular = "Etiket"
	tagBasePath = "/admin/tags"
)

// TagHandler — etiket yönetim sayfası. Ortak views.Listing* iskeletini
// kullanır; entity'ler view-model'e (ListRow) burada çevrilir.
type TagHandler struct {
	svc *tagService
}

func NewTagHandler(svc *tagService) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) Mount(app *fiber.App) {
	app.Get(tagBasePath, auth.RequireAdmin, h.list)
	app.Get(tagBasePath+"/new", auth.RequireAdmin, h.newForm)
	app.Post(tagBasePath, auth.RequireAdmin, h.create)
	app.Get(tagBasePath+"/:id/edit", auth.RequireAdmin, h.editForm)
	app.Put(tagBasePath+"/:id", auth.RequireAdmin, h.update)
	app.Delete(tagBasePath+"/:id", auth.RequireAdmin, h.delete)
}

// props — view props'unu oturum/csrf ile birlikte kurar.
func (h *TagHandler) props(c fiber.Ctx, page *TagPage) views.ListingProps {
	rows := make([]views.ListRow, 0, len(page.Items))
	for _, b := range page.Items {
		rows = append(rows, views.ListRow{ID: b.ID.Hex(), Name: b.Name, CreatedAt: b.CreatedAt})
	}
	return views.ListingProps{
		Title:    tagTitle,
		Singular: tagSingular,
		BasePath: tagBasePath,
		Icon:     icons.Tag,
		CSRF:     session.Token(c),
		Sess:     session.FromCtx(c),
		Query:    strings.TrimSpace(fiber.Query[string](c, "q")),
		Page: views.Listing{
			Items:      rows,
			Total:      page.Total,
			Page:       page.Page,
			PerPage:    page.PerPage,
			TotalPages: page.TotalPages,
		},
	}
}

// loadPage — istenen sayfayı getirir; sayfa taşmışsa (ör. son kayıt
// silindi) son geçerli sayfaya çekilir.
func (h *TagHandler) loadPage(c fiber.Ctx, page int) (*TagPage, error) {
	q := strings.TrimSpace(fiber.Query[string](c, "q"))
	res, err := h.svc.List(c.Context(), ListTagsRequest{Page: page, Query: q})
	if err != nil {
		return nil, err
	}
	if res.TotalPages > 0 && res.Page > res.TotalPages {
		return h.svc.List(c.Context(), ListTagsRequest{Page: res.TotalPages, Query: q})
	}
	return res, nil
}

// list — tam sayfa; htmx istekleri (pagination) sadece tablo fragment'ı alır.
func (h *TagHandler) list(c fiber.Ctx) error {
	res, err := h.loadPage(c, fiber.Query(c, "page", 1))
	if err != nil {
		return err
	}
	p := h.props(c, res)
	if httpx.IsHTMX(c) {
		return httpx.Render(c, views.ListingTable(p))
	}
	return httpx.Render(c, views.ListingPage(p))
}

func (h *TagHandler) newForm(c fiber.Ctx) error {
	return httpx.Render(c, views.ListingFormModal(views.ListingFormProps{
		Singular: tagSingular,
		Action:   tagBasePath + httpx.ListQuery(c, 1),
		Method:   "post",
	}))
}

func (h *TagHandler) editForm(c fiber.Ctx) error {
	b, err := h.svc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return httpx.Render(c, views.ListingFormModal(views.ListingFormProps{
			Singular: tagSingular,
			Action:   tagBasePath,
			Method:   "post",
			Err:      friendlyTagErr(err),
		}))
	}
	return httpx.Render(c, views.ListingFormModal(views.ListingFormProps{
		Singular: tagSingular,
		Action:   fmt.Sprintf("%s/%s%s", tagBasePath, b.ID.Hex(), httpx.ListQuery(c, fiber.Query(c, "page", 1))),
		Method:   "put",
		Name:     b.Name,
		IsEdit:   true,
	}))
}

func (h *TagHandler) create(c fiber.Ctx) error {
	req := CreateTagRequest{Name: strings.TrimSpace(c.FormValue("name"))}
	form := views.ListingFormProps{
		Singular: tagSingular,
		Action:   tagBasePath + httpx.ListQuery(c, 1),
		Method:   "post",
		Name:     req.Name,
	}

	if err := req.Validate(); err != nil {
		form.Err = err.Error()
		return httpx.Render(c, views.ListingFormModal(form))
	}
	if _, err := h.svc.Create(c.Context(), req); err != nil {
		form.Err = friendlyTagErr(err)
		return httpx.Render(c, views.ListingFormModal(form))
	}

	// yeni kayıt en üstte görünsün diye 1. sayfa
	return h.renderTableOOB(c, 1)
}

func (h *TagHandler) update(c fiber.Ctx) error {
	page := fiber.Query(c, "page", 1)
	req := UpdateTagRequest{
		ID:   c.Params("id"),
		Name: strings.TrimSpace(c.FormValue("name")),
	}
	form := views.ListingFormProps{
		Singular: tagSingular,
		Action:   fmt.Sprintf("%s/%s%s", tagBasePath, req.ID, httpx.ListQuery(c, page)),
		Method:   "put",
		Name:     req.Name,
		IsEdit:   true,
	}

	if err := req.Validate(); err != nil {
		form.Err = err.Error()
		return httpx.Render(c, views.ListingFormModal(form))
	}
	if _, err := h.svc.Update(c.Context(), req); err != nil {
		form.Err = friendlyTagErr(err)
		return httpx.Render(c, views.ListingFormModal(form))
	}

	return h.renderTableOOB(c, page)
}

func (h *TagHandler) delete(c fiber.Ctx) error {
	err := h.svc.Delete(c.Context(), c.Params("id"))
	// zaten silinmiş/geçersiz kayıt için tabloyu tazelemek yeterli
	if err != nil && !errors.Is(err, ErrTagNotFound) && !errors.Is(err, ErrInvalidID) {
		return err
	}

	res, err := h.loadPage(c, fiber.Query(c, "page", 1))
	if err != nil {
		return err
	}
	return httpx.Render(c, views.ListingTable(h.props(c, res)))
}

// renderTableOOB — form başarı yanıtı: gövdedeki tek içerik hx-swap-oob'lu
// tablodur; formun hedefi (#listing-modal) boş kalan içerikle swap edilip
// modal kapanır, liste out-of-band yenilenir.
func (h *TagHandler) renderTableOOB(c fiber.Ctx, page int) error {
	res, err := h.loadPage(c, page)
	if err != nil {
		return err
	}
	p := h.props(c, res)
	p.OOB = true
	return httpx.Render(c, views.ListingTable(p))
}

// friendlyTagErr — bilinen servis hataları kullanıcıya gösterilir,
// beklenmeyenler generic mesaja çevrilir (İngilizce db hatası forma basılmaz).
func friendlyTagErr(err error) string {
	switch {
	case errors.Is(err, ErrTagNameTaken),
		errors.Is(err, ErrTagNotFound),
		errors.Is(err, ErrInvalidID):
		return err.Error()
	default:
		return "Beklenmeyen bir hata oluştu, lütfen tekrar deneyin"
	}
}
