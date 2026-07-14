# E-Commerce Template

Go + templ + htmx + Tailwind v4 (browser build) + MongoDB. `make dev` ile çalışır;
templ dosyaları değişince `templ generate` (veya `make templ`) gerekir.

## Tasarım Sistemi Kuralları

Tek kaynak: **`styles/theme.css`**. Bu dosya Go embed ile (`styles/embed.go` →
`internal/shared/ui/theme.templ` içindeki `ThemeHead()`) her sayfanın `<head>`'ine
`<style type="text/tailwindcss">` olarak gömülür — Tailwind v4 browser build
`<link>`'i işleyemediği için bu yol zorunludur. daisyUI KULLANILMAZ (kaldırıldı).

1. **Ham Tailwind renk/spacing/font class'ı YASAK.** `bg-blue-600`, `font-bold`,
   `text-4xl`, `shadow-xl`, `text-[11px]` gibi hardcoded değerler kullanılmaz.
   Default renk paleti, font-weight ve text/radius/shadow skalaları theme.css'te
   `--color-*: initial` vb. ile SİLİNMİŞTİR — bu class'lar zaten üretilmez.
2. **Yeni görsel ihtiyaç → önce token.** Yeni renk/boyut/gölge gerekiyorsa önce
   `styles/theme.css`'e token eklenir, sonra component'te kullanılır. Component
   koduna asla direkt değer yazılmaz.
3. **Content çifti kuralı.** Renkli zemin üstüne yazı her zaman ilgili content
   token'ıyla yazılır: `bg-danger` üstünde `text-danger-content`. Asla
   `text-white`/`text-black` yazılmaz (zaten üretilmez).
4. **Mod'dan bağımsız token'lar** (font, boyut, radius, z-index, spacing...)
   sadece `@theme` içinde tanımlanır; `:root`/`.dark`'a girmez.
   **Mod'a bağımlı token'lar** (renk, gölge, overlay, ring, skeleton) `:root`
   (light) ve `.dark` içinde AYRI AYRI tanımlanır ve `@theme inline` bloğunda
   utility'ye bağlanır (`--color-primary: var(--primary)` pattern'i).
5. **Dark mode:** `<html>`'e `.dark` class'ı ile. Tercih `localStorage("theme")`'de;
   ilk ziyarette `prefers-color-scheme` varsayılır. FOUC guard script'i
   `ThemeHead()` içindedir ve style'dan önce çalışır. Toggle: `ui.ThemeToggle()`
   (iki layout'un navbar'ında). `dark:` variant'ı `@custom-variant` ile class-tabanlıdır.
6. Layout/flex/grid/display ve Tailwind'in default spacing skalası (`p-4`,
   `gap-2`, `w-full`...) serbesttir — bunlar token gerektirmez.
7. **Interaction state'leri merkezidir.** Hover/active/disabled/loading
   efektleri `--opacity-hover`, `--scale-active`, `--ring-width`,
   `--ring-offset`, `--opacity-disabled`, `--opacity-loading` token'larından
   gelir. Component koduna `hover:opacity-90` gibi hardcoded değer yazılmaz;
   interaktif davranış component class'ında (`.btn`, `.link`, `.input`,
   `.card-hover`...) token referansıyla tanımlıdır. Tıklanabilir kartlarda
   `hover:shadow-*`/`hover:-translate-*` yerine `.card-hover` kullanılır.
8. **Focus her zaman `:focus-visible` ile**, asla `:focus` değil (mouse
   tıklamasında gereksiz ring çıkmasın). Ring her yerde aynı formül:
   `var(--ring-width)` solid `var(--ring-color)` + `var(--ring-offset)` offset.
9. **Disabled'da hover/active iptal edilir.** Component CSS'inde
   `:hover:not(:disabled)` / `:active:not(:disabled)` guard'ı zorunlu; templ
   tarafında utility ile efekt yazılmak zorunda kalınırsa
   `disabled:active:scale-100` gibi override eklenir. htmx isteği sırasında
   `.btn.htmx-request` otomatik `--opacity-loading` + `pointer-events:none` alır.

### styles/theme.css token envanteri

### Tema kimliği (mevcut değerler)

- Palet: şişe yeşili primary (light `#146c43`, dark nane `#56c894`), kampanya
  turuncusu accent (`#c2410c` / `#fb923c`), sıcak kağıt zeminler (sayfa
  `#f6f3ec`, kart beyaz); dark mode yeşile çalan koyu tonlar (nötr gri değil).
- Fontlar: **Fraunces** (display — `h1,h2,h3` base'de otomatik alır) +
  **Manrope** (gövde). Google Fonts `<link>`'i `ThemeHead()` içinde.
- İmza dokunuş: `.btn` pill (tam yuvarlak) formdadır; input'lar
  `radius-control`'de kalır. Rakamlar `tnum` ile hizalıdır.

**Mod'dan bağımsız (@theme):**
- Font: `--font-sans`(Manrope), `--font-display`(Fraunces), `--font-mono` → `font-sans|display|mono`
- Text: `--text-xs`(12px) `--text-sm` `--text-base` `--text-lg` `--text-xl`
  `--text-2xl` `--text-3xl`(32px) → `text-xs`…`text-3xl` (4xl+ YOK)
- Weight: `--font-weight-normal`(400) `-medium`(500) `-semibold`(600) →
  `font-normal|medium|semibold` (font-bold YOK)
- Line-height: `--leading-tight|normal|relaxed`
- Tracking: `--tracking-tight|normal|wide` (uppercase label'da `tracking-wide`)
- Radius: `--radius-control`(input/buton) `--radius-card` `--radius-pill` →
  `rounded-control|card|pill` (başka rounded YOK)
- Border width: `--border-width-thin`(1px) `-thick`(2px) → `border-thin|thick`
- Transition: `--duration-fast`(120ms) `--duration-base`(200ms) →
  `duration-fast|base`; `--ease-default` → `ease-default`
- Z-index: `--z-sticky`(30) `--z-dropdown`(40) `--z-modal`(50) `--z-toast`(60)
  `--z-tooltip`(70) → `z-sticky|dropdown|modal|toast|tooltip`
- Interaction: `--opacity-hover`(0.9) `--opacity-disabled`(0.5)
  `--opacity-loading`(0.6) `--scale-active`(0.97) `--ring-width`(2px)
  `--ring-offset`(2px) — component CSS'inde var() ile, utility üretmez
- Spacing: `--spacing-control-y|control-x|card` (component CSS'inde var() ile)
- Icon: `--spacing-icon-sm`(16px) `-base`(20px) `-lg`(24px) → `size-icon-sm|base|lg`
- Container: `--container-content`(72rem) `--container-narrow`(36rem) →
  `max-w-content|narrow`

**Mod'a bağımlı (:root / .dark → @theme inline ile utility):**
- Renkler: `--primary/--primary-content`, `--accent/--accent-content`,
  `--success/…`, `--danger/…`, `--warning/…` → `bg-primary`, `text-primary-content`…
- Zemin: `--surface`(kart) `--surface-alt`(sayfa) → `bg-surface|surface-alt`
- Metin: `--ink` `--ink-muted` → `text-ink|ink-muted`
- Kenarlık: `--border-c` → `border-border`, `divide-border`
- Gölge: `--elev-card|popover|modal` → `shadow-card|popover|modal`
  (dark'ta border-ağırlıklı/hafif)
- Overlay: `--overlay` → `bg-overlay`
- Ring: `--ring-color` → `ring-ring`, component'lerde focus outline
- Skeleton: `--skeleton-base|highlight` → `.skeleton` shimmer

**Component class envanteri (@layer components, hepsi token türevi):**
`btn` (+ `btn-primary|accent|success|danger|ghost|outline|active`,
`btn-xs|sm|lg`), `badge` (+ `badge-primary|accent|success|warning|danger|ghost`,
`badge-sm|lg`), `input`, `select`, `textarea`, `input-sm`, `select-sm`,
`checkbox`(+`checkbox-sm`), `toggle`(+`toggle-sm`), `form-control`, `label`,
`label-text`, `card`, `card-hover`(tıklanabilir kart), `card-body`,
`card-title`, `table`, `alert`
(+`alert-danger|success`), `modal`, `modal-open`, `modal-box`, `link`,
`link-primary`, `join`, `join-item`, `stats`, `stat`, `stat-title`,
`stat-value`, `divider`, `loading`(+`loading-sm`), `skeleton`, `navbar`.

Eski daisyUI isimlerinden farklılar: `*-error` → `*-danger`,
`*-secondary` → `*-accent`, `input-bordered` yok (base'e gömülü),
`bg-base-100/200/300` → `bg-surface`/`bg-surface-alt`,
`text-base-content` → `text-ink`, `rounded-box|field` → `rounded-card|control`.
