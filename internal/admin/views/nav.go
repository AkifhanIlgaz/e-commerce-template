package views

import (
	"strings"

	"ecommerce/internal/shared/icons"

	"github.com/a-h/templ"
)

// Crumb — breadcrumb öğesi. Href boş bırakılırsa aktif (son) sayfa kabul
// edilir ve link yerine düz metin basılır.
type Crumb struct {
	Label string
	Href  string
}

// IconFunc — internal/shared/icons paketindeki ikon component'lerinin imzası.
type IconFunc = func(templ.Attributes) templ.Component

// NavItem — sidebar menü öğesi. Children doluysa link yerine açılır-kapanır
// grup (SidebarGroup) render edilir; bu durumda Href kullanılmaz. Alt
// linkler girintili listelenir; Icon opsiyoneldir (nil ise sadece metin).
type NavItem struct {
	Label    string
	Href     string
	Icon     IconFunc
	Children []NavItem
}

// NavItems — admin sidebar menüsü. Bu template'i klonlayan projeler menüyü
// SADECE bu listeyi düzenleyerek değiştirir; sidebar markup'ına dokunulmaz.
var NavItems = []NavItem{
	{Label: "Panel", Href: "/admin", Icon: icons.LayoutDashboard},
	{Label: "Ürünler", Icon: icons.Package, Children: []NavItem{
		{Label: "Ürünler", Href: "/admin/products", Icon: icons.List},
		{Label: "Kategoriler", Href: "/admin/categories", Icon: icons.Tags},
		{Label: "Markalar", Href: "/admin/brands", Icon: icons.Tags},
		{Label: "Etiketler", Href: "/admin/tags", Icon: icons.Tags},
	}},
	{Label: "Siparişler", Href: "/admin/orders", Icon: icons.ShoppingCart},
	{Label: "Müşteriler", Href: "/admin/customers", Icon: icons.Users},
	{Label: "Ayarlar", Href: "/admin/settings", Icon: icons.Settings},
}

// navActive — aktif sayfanın linkini işaretler. "/admin" yalnızca tam
// eşleşmede aktiftir; diğer linkler alt sayfalarını da kapsar
// (/admin/products/42 → Ürünler aktif).
func navActive(path, href string) bool {
	if href == "/admin" {
		return path == "/admin"
	}
	return path == href || strings.HasPrefix(path, href+"/")
}

// groupOpen — alt linklerden biri aktifse grup sayfa yüklenirken açık gelir.
func groupOpen(path string, it NavItem) bool {
	for _, c := range it.Children {
		if navActive(path, c.Href) {
			return true
		}
	}
	return false
}
