# E-Ticaret Template

Go + Fiber v3 + templ + htmx + Tailwind CSS + DaisyUI ile e-ticaret site şablonu.
Her yeni müşteri için bu repo kopyalanır; **admin paneli sabit kalır**,
**storefront** (müşteri yüzü) özelleştirilir.

## Stack

| Katman | Teknoloji |
|---|---|
| HTTP router | Fiber v3 |
| Template | templ |
| Frontend etkileşim | htmx (vendor'lanmış, `static/js/`) |
| CSS | Tailwind (browser build) + DaisyUI (vendor'lanmış, `static/css/`) |
| Veritabanı | MongoDB (ürün, sipariş, kullanıcı) |
| Cache/Session/Sepet | Redis |
| Ödeme | `PaymentProvider` interface + mock 3DS akışı |

## Hızlı Başlangıç

```bash
cp .env.example .env      # gerekirse değerleri düzenle
make infra                # Mongo + Redis (docker compose up -d)
make run                  # templ generate + go run ./cmd/server
```

- Storefront: http://localhost:8080
- Admin: http://localhost:8080/admin — ilk açılışta `.env`'deki `ADMIN_EMAIL` / `ADMIN_PASSWORD` ile admin kullanıcısı otomatik oluşturulur (varsayılan: `admin@example.com` / `admin123`).
- İlk açılışta 4 demo ürün otomatik eklenir (veritabanı boşsa).

> `templ` CLI gerekli: `go install github.com/a-h/templ/cmd/templ@latest`

## Mimari

```
cmd/server/            # main: config + bağlantılar + route mount
internal/
  config/              # TÜM ayarlar env variable'dan (.env desteği dahil)
  core/                # PAYLAŞILAN iş mantığı — müşteriye göre DEĞİŞMEZ
    product/           # model + mongo repository + service
    cart/              # Redis tabanlı sepet (misafir destekli, cookie ile)
    checkout/          # sepet -> ödeme -> sipariş orkestrasyonu
    order/             # sipariş modeli + durum makinesi + dashboard istatistikleri
    auth/              # kullanıcı (bcrypt) + Redis session (admin/store ayrı scope)
    payment/           # PaymentProvider interface + mock 3DS implementasyonu
  admin/               # SABİT — jenerik DaisyUI arayüzü, temadan bağımsız
    handlers/          # /admin/* route'ları (RequireRole + CSRF korumalı)
    views/             # templ dosyaları (kendi layout'u: admin Layout)
  storefront/          # MÜŞTERİYE ÖZEL — her kopyada burası değişir
    handlers/          # route'lar sabit, genelde dokunulmaz
    views/             # "MÜŞTERİYE GÖRE DÜZENLE" yorumlu templ dosyaları
      home.templ       #   <- en çok değişecek dosya
      components/      #   ürün kartı, hızlı bakış modalı
  shared/              # db bağlantıları, middleware (auth/CSRF), httpx helpers
static/
  css/theme-config.css # <- TÜM marka renkleri/fontları TEK dosyada
  css/daisyui.css      # vendor (dokunma)
  js/htmx.min.js       # vendor (dokunma)
```

Admin ve storefront **aynı core servisleri paylaşır** — admin'den eklenen ürün
storefront'ta anında görünür; sadece render ettikleri templ farklıdır.

## ✅ Yeni Müşteri İçin Checklist

Repoyu kopyaladıktan (clone/fork) sonra sırasıyla:

1. **`.env` oluştur** — `cp .env.example .env`, sonra düzenle:
   - `STORE_NAME` → müşterinin marka adı (navbar, başlıklar, footer'a otomatik yansır)
   - `ADMIN_EMAIL` / `ADMIN_PASSWORD` → müşterinin admin girişi
   - `MONGO_DB` → müşteriye özel db adı (örn. `musteri_x`)
   - Prod'da: `APP_ENV=prod` (Secure cookie zorunlu olur, varsayılan admin şifresi reddedilir)

2. **`static/css/theme-config.css`** → marka renkleri, köşe yuvarlaklıkları, font.
   Tek dosya, CSS variable'lar; DaisyUI component'lerine otomatik yansır.
   Admin paneli bu dosyayı yüklemediği için etkilenmez.

3. **`internal/storefront/views/home.templ`** → anasayfa (hero metni, kampanya
   bölümleri). En çok değişecek dosya.

4. **`internal/storefront/views/layout.templ`** → navbar/footer, menü linkleri,
   Google Fonts `<link>`'i (gerekiyorsa).

5. **İhtiyaca göre** (hepsi `// MÜŞTERİYE GÖRE DÜZENLE` yorumu taşır):
   - `views/components/product_card.templ` — ürün kartı görünümü
   - `views/components/quick_view.templ` — hızlı bakış modalı
   - `views/product_list.templ`, `product_detail.templ`, `cart.templ`, `checkout.templ`, `auth.templ`

6. **Gerçek ödeme sağlayıcısı** (hazır olunca):
   - `internal/core/payment/` altına `Provider` interface'ini implemente eden dosya ekle
   - `provider.go` içindeki `NewProvider` factory'sine bir `case` ekle
   - `.env`'de `PAYMENT_PROVIDER`, `PAYMENT_API_KEY`, `PAYMENT_SECRET_KEY`, `PAYMENT_BASE_URL` doldur
   - Handler'lar ve checkout akışı değişmez.

**Dokunma listesi:** `internal/core/`, `internal/admin/`, `internal/shared/`,
`internal/storefront/handlers/`, `static/css/daisyui.css`, `static/js/`.

## Akışlar

### Admin → Storefront doğrulama akışı
1. `/admin/login` → giriş yap
2. Ürünler → "+ Yeni Ürün" → kaydet (veya tablodaki "Düzenle" ile inline edit)
3. Storefront `/products` → ürün anında görünür (aynı product service)

### Ödeme (mock 3DS)
1. Sepete ekle → `/checkout` → form doldur → "Ödemeye Devam Et"
2. Sahte banka ekranı iframe'de açılır; sayfa aynı anda htmx ile 2 sn'de bir durum sorar (polling)
3. iframe'de **Onayla** → polling `HX-Redirect` ile başarı sayfasına götürür, sipariş `paid` olur, stok düşer, sepet temizlenir
4. **Reddet** → sipariş `cancelled`, sepet durur, tekrar denenebilir

### Güvenlik
- Admin ve müşteri session'ları tamamen ayrı: farklı cookie (`admin_session` sadece `/admin` path'inde), farklı Redis prefix'i, rol kontrolü
- Session ID `crypto/rand`, HttpOnly + SameSite=Lax (+ prod'da Secure) cookie
- Session fixation önlemi: login'de eski session silinir, yeni ID üretilir
- Idle timeout (varsayılan 30 dk) + absolute timeout (12 saat) — `.env`'den ayarlanır
- CSRF: `<meta name="csrf-token">` + htmx `configRequest` header'ı; misafirde double-submit cookie
- Şifreler bcrypt ile hash'lenir

## Komutlar

```bash
make infra        # Mongo + Redis konteynerlerini başlat
make infra-down   # konteynerleri durdur
make run          # templ generate + go run
make build        # bin/server binary'si üret
templ generate --watch   # geliştirmede templ'leri izle
```
