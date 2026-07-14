package session

import "time"

// Session'ın kendisi leaf `session` paketinde yaşar: view'lar (templ) tipe
// oradan erişir; auth paketi handler'ları view'ları import ettiği için tip
// burada tanımlı kalsaydı import döngüsü oluşurdu. Alias sayesinde auth
// içindeki kod auth.Session yazmaya devam eder.
type Session struct {
	ID        string
	UserID    string
	Email     string
	Name      string
	Role      string
	CSRFToken string
	CreatedAt time.Time
	LastSeen  time.Time
}

// Scope, admin ve müşteri session'larını birbirinden tamamen ayırır:
// farklı cookie adı, farklı Redis prefix'i, farklı path.
type Scope struct {
	CookieName  string
	RedisPrefix string
	CookiePath  string
	LoginURL    string // yetkisiz istekte yönlendirilecek sayfa
}

var (
	// AdminScope: cookie sadece /admin path'ine gönderilir — storefront istekleri admin cookie'sini hiç görmez.
	AdminScope = Scope{CookieName: "admin_session", RedisPrefix: "sess:admin:", CookiePath: "/admin", LoginURL: "/admin/login"}
	StoreScope = Scope{CookieName: "store_session", RedisPrefix: "sess:store:", CookiePath: "/", LoginURL: "/login"}
)
