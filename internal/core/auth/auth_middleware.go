package auth

import "ecommerce/internal/core/session"

// session paketindeki scope'lara kısayol: handler'lar AdminScope/StoreScope
// yazabilsin diye. Scope tanımları session'da yaşar çünkü middleware paketi
// de onları kullanır; burada tanımlansaydı middleware→auth→middleware
// döngüsü oluşurdu.
var (
	AdminScope = session.AdminScope
	StoreScope = session.StoreScope
)
