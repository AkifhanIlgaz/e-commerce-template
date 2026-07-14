// Package styles, tasarım sistemi token dosyasını (theme.css) binary'ye gömer.
// Tailwind v4 browser build sadece <style type="text/tailwindcss"> bloklarını
// işleyebildiği için CSS, ayrı bir HTTP dosyası yerine layout'lara inline
// gömülür (bkz. internal/shared/ui/theme.templ).
package styles

import _ "embed"

//go:embed theme.css
var ThemeCSS string
