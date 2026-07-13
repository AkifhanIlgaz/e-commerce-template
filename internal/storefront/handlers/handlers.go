// Package handlers (storefront) — route'lar ve akış sabittir, görünümü
// views/ altındaki templ dosyaları belirler. Yeni müşteride genelde bu
// dosyaya dokunmak gerekmez.
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/cart"
	"ecommerce/internal/core/checkout"
	"ecommerce/internal/core/order"
	"ecommerce/internal/core/payment"
	"ecommerce/internal/core/product"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/shared/middleware"
	views "ecommerce/internal/storefront/views"
	"ecommerce/internal/storefront/views/components"
)

type Handlers struct {
	storeName string
	products  *product.Service
	carts     *cart.Service
	orders    *order.Service
	checkout  *checkout.Service
	users     *auth.Service
	sessions  *auth.SessionManager
	provider  payment.Provider
}

func New(storeName string, products *product.Service, carts *cart.Service, orders *order.Service,
	co *checkout.Service, users *auth.Service, sessions *auth.SessionManager, provider payment.Provider) *Handlers {
	return &Handlers{
		storeName: storeName, products: products, carts: carts, orders: orders,
		checkout: co, users: users, sessions: sessions, provider: provider,
	}
}

// Mount, storefront route'larını bağlar. Tüm route'lar WithSession (varsa
// müşteri session'ını Locals'a koyar) + CSRF zincirinden geçer.
func (h *Handlers) Mount(app *fiber.App) {
	store := app.Group("/",
		middleware.WithSession(h.sessions, auth.StoreScope),
		middleware.CSRF,
	)

	store.Get("/", h.home)
	store.Get("/products", h.productList)
	store.Get("/products/search", h.productSearch)
	store.Get("/products/:idOrSlug/quick-view", h.quickView)
	store.Get("/products/:idOrSlug", h.productDetail)

	store.Post("/cart/add", h.cartAdd)
	store.Get("/cart", h.cartPage)
	store.Post("/cart/update", h.cartUpdate)
	store.Post("/cart/remove", h.cartRemove)

	store.Get("/checkout", h.checkoutPage)
	store.Post("/checkout", h.checkoutSubmit)
	store.Get("/checkout/payment/:orderID/status", h.paymentStatus)
	store.Get("/orders/:id/success", h.orderSuccess)
	store.Get("/orders/:id/failed", h.orderFailed)

	// mock 3DS ekranı (iframe içinde açılır; sadece mock provider'da kullanılır)
	store.Get("/pay/mock-3ds", h.mock3DSPage)
	store.Post("/pay/mock-3ds/resolve", h.mock3DSResolve)

	store.Get("/login", h.loginPage)
	store.Post("/login", h.login)
	store.Get("/register", h.registerPage)
	store.Post("/register", h.register)
	store.Post("/logout", h.logout)
}

// layoutData, her sayfa için ortak veriyi toplar (sepet sayacı, müşteri adı, csrf).
func (h *Handlers) layoutData(c fiber.Ctx) views.LayoutData {
	d := views.LayoutData{
		StoreName: h.storeName,
		CSRFToken: middleware.CSRFToken(c),
	}
	if sess := middleware.SessionFromCtx(c); sess != nil {
		d.CustomerName = sess.Name
	}
	cartID := h.carts.CartID(c)
	if crt, err := h.carts.Get(c.Context(), cartID); err == nil {
		d.CartCount = crt.Count()
	}
	return d
}

// --- sayfalar ---

func (h *Handlers) home(c fiber.Ctx) error {
	list, err := h.products.List(c.Context(), product.ListFilter{ActiveOnly: true})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if len(list) > 8 {
		list = list[:8]
	}
	return httpx.Render(c, views.Home(h.layoutData(c), list))
}

func (h *Handlers) productList(c fiber.Ctx) error {
	q := c.Query("q")
	list, err := h.products.List(c.Context(), product.ListFilter{ActiveOnly: true, Search: q})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return httpx.Render(c, views.ProductList(h.layoutData(c), list, q))
}

// productSearch — canlı aramanın çağırdığı fragment endpoint'i.
func (h *Handlers) productSearch(c fiber.Ctx) error {
	list, err := h.products.List(c.Context(), product.ListFilter{ActiveOnly: true, Search: c.Query("q")})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return httpx.Render(c, views.ProductGrid(list))
}

func (h *Handlers) productDetail(c fiber.Ctx) error {
	p, err := h.products.GetBySlug(c.Context(), c.Params("idOrSlug"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, views.ProductDetail(h.layoutData(c), *p))
}

func (h *Handlers) quickView(c fiber.Ctx) error {
	p, err := h.products.GetByID(c.Context(), c.Params("idOrSlug"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, components.QuickView(*p))
}

// --- sepet ---

// cartAdd — cevapta "Eklendi" butonu + out-of-band sepet sayacı döner.
func (h *Handlers) cartAdd(c fiber.Ctx) error {
	cartID := h.carts.CartID(c)
	productID := c.FormValue("product_id")
	if err := h.carts.Add(c.Context(), cartID, productID, 1); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	crt, err := h.carts.Get(c.Context(), cartID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if err := httpx.Render(c, components.AddedToCartButton(productID, "")); err != nil {
		return err
	}
	return httpx.Render(c, views.CartCounter(crt.Count(), true))
}

func (h *Handlers) cartPage(c fiber.Ctx) error {
	crt, err := h.carts.Get(c.Context(), h.carts.CartID(c))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return httpx.Render(c, views.CartPage(h.layoutData(c), crt))
}

func (h *Handlers) cartUpdate(c fiber.Ctx) error {
	cartID := h.carts.CartID(c)
	qty, _ := strconv.Atoi(c.FormValue("qty"))
	if err := h.carts.SetQty(c.Context(), cartID, c.FormValue("product_id"), qty); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return h.renderCart(c, cartID)
}

func (h *Handlers) cartRemove(c fiber.Ctx) error {
	cartID := h.carts.CartID(c)
	if err := h.carts.Remove(c.Context(), cartID, c.FormValue("product_id")); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return h.renderCart(c, cartID)
}

func (h *Handlers) renderCart(c fiber.Ctx, cartID string) error {
	crt, err := h.carts.Get(c.Context(), cartID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return httpx.Render(c, views.CartContents(crt))
}

// --- checkout / ödeme ---

func (h *Handlers) checkoutPage(c fiber.Ctx) error {
	crt, err := h.carts.Get(c.Context(), h.carts.CartID(c))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if len(crt.Items) == 0 {
		return c.Redirect().Status(fiber.StatusSeeOther).To("/cart")
	}
	var email, name string
	if sess := middleware.SessionFromCtx(c); sess != nil {
		email, name = sess.Email, sess.Name
	}
	return httpx.Render(c, views.CheckoutPage(h.layoutData(c), crt, email, name, ""))
}

func (h *Handlers) checkoutSubmit(c fiber.Ctx) error {
	cartID := h.carts.CartID(c)
	in := checkout.PlaceOrderInput{
		CartID:   cartID,
		Email:    c.FormValue("email"),
		FullName: c.FormValue("full_name"),
		Address:  c.FormValue("address"),
	}
	if sess := middleware.SessionFromCtx(c); sess != nil {
		in.UserID = sess.UserID
	}
	o, pay, err := h.checkout.PlaceOrder(c.Context(), in)
	if err != nil {
		crt, cerr := h.carts.Get(c.Context(), cartID)
		if cerr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, cerr.Error())
		}
		return httpx.Render(c, views.CheckoutPage(h.layoutData(c), crt, in.Email, in.FullName, err.Error()))
	}
	return httpx.Render(c, views.PaymentPage(h.layoutData(c), o.ID.Hex(), pay.ThreeDSURL))
}

// paymentStatus — htmx polling endpoint'i. Ödeme sonuçlanınca HX-Redirect
// ile sonuç sayfasına yönlendirir; sonuçlanmadıysa spinner'ı korur.
func (h *Handlers) paymentStatus(c fiber.Ctx) error {
	res, err := h.checkout.CheckPayment(c.Context(), c.Params("orderID"), h.carts.CartID(c))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if res.Done {
		if res.Succeeded {
			c.Set("HX-Redirect", "/orders/"+res.OrderID+"/success")
		} else {
			c.Set("HX-Redirect", "/orders/"+res.OrderID+"/failed")
		}
		return c.SendString("")
	}
	c.Type("html", "utf-8")
	return c.SendString(`<span class="loading loading-spinner loading-sm"></span> Ödeme durumu kontrol ediliyor...`)
}

func (h *Handlers) orderSuccess(c fiber.Ctx) error {
	o, err := h.orders.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, views.OrderSuccessPage(h.layoutData(c), o))
}

func (h *Handlers) orderFailed(c fiber.Ctx) error {
	return httpx.Render(c, views.OrderFailedPage(h.layoutData(c), c.Params("id")))
}

// --- mock 3DS ---

func (h *Handlers) mock3DSPage(c fiber.Ctx) error {
	return httpx.Render(c, views.Mock3DSPage(c.Query("payment_id"), middleware.CSRFToken(c)))
}

func (h *Handlers) mock3DSResolve(c fiber.Ctx) error {
	mock, ok := h.provider.(*payment.MockProvider)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "mock provider aktif değil")
	}
	approved := c.FormValue("decision") == "approve"
	if err := mock.Resolve(c.FormValue("payment_id"), approved); err != nil {
		return httpx.Render(c, views.Mock3DSDone(err.Error()))
	}
	if approved {
		return httpx.Render(c, views.Mock3DSDone("Ödeme onaylandı. Ana pencerede yönlendiriliyorsunuz..."))
	}
	return httpx.Render(c, views.Mock3DSDone("Ödeme reddedildi."))
}

// --- müşteri auth ---

func (h *Handlers) loginPage(c fiber.Ctx) error {
	return httpx.Render(c, views.CustomerLogin(h.layoutData(c), ""))
}

func (h *Handlers) login(c fiber.Ctx) error {
	u, err := h.users.Authenticate(c.Context(), c.FormValue("email"), c.FormValue("password"), auth.RoleCustomer)
	if err != nil {
		return httpx.Render(c, views.CustomerLogin(h.layoutData(c), "E-posta veya şifre hatalı"))
	}
	// session fixation önlemi: eski session'ı sil, yeni ID üret
	h.sessions.Destroy(c.Context(), c, auth.StoreScope)
	if _, err := h.sessions.Create(c.Context(), c, auth.StoreScope, u); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "oturum oluşturulamadı")
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To("/")
}

func (h *Handlers) registerPage(c fiber.Ctx) error {
	return httpx.Render(c, views.CustomerRegister(h.layoutData(c), ""))
}

func (h *Handlers) register(c fiber.Ctx) error {
	u, err := h.users.Register(c.Context(), c.FormValue("email"), c.FormValue("password"), c.FormValue("name"), auth.RoleCustomer)
	if err != nil {
		return httpx.Render(c, views.CustomerRegister(h.layoutData(c), err.Error()))
	}
	h.sessions.Destroy(c.Context(), c, auth.StoreScope)
	if _, err := h.sessions.Create(c.Context(), c, auth.StoreScope, u); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "oturum oluşturulamadı")
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To("/")
}

func (h *Handlers) logout(c fiber.Ctx) error {
	h.sessions.Destroy(c.Context(), c, auth.StoreScope)
	return httpx.Redirect(c, "/")
}
