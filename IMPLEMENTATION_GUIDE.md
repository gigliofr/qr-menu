📋 GUIDA IMPLEMENTAZIONE - DESIGN SYSTEM QR MENU
==============================================

## 📁 File Creati

### CSS Design System
1. **static/css/design-system.css** (2.5KB)
   - CSS custom properties (variabili colore, spacing, typography)
   - Reset e base styles
   - Utility classes
   - Dark mode support
   - Responsive breakpoints

2. **static/css/components.css** (4KB)
   - Buttons (4 variants × 3 sizes)
   - Cards
   - Forms (inputs, labels, validation states)
   - Navigation bar
   - Tab bar (mobile)
   - Modal/Dialog
   - Alerts & Badges
   - Loading states

3. **static/css/icons.css** (1KB)
   - Icon sizing utilities
   - Icon colors
   - Icon animations (spin, pulse)
   - Icon buttons
   - Icon badge (notification dot)

4. **static/icons-library.html**
   - Library SVG icons (Apple-style minimal)
   - 30+ icone comuni
   - Pronte per essere copiate nei template

### Template Esempi
5. **templates/login_redesigned.html**
   - Esempio di redesign completo con nuovo design system
   - Form validation di base
   - Animazioni fluide
   - Mobile-responsive

6. **templates/public_menu_redesigned.html**
   - Esempio menu pubblico completo
   - Grid responsive (1-3 colonne)
   - Sticky header
   - Section filtering
   - Item cards con hover effect

### Documentazione
7. **DESIGN_PLAN.md**
   - Piano completo di miglioramenti grafici
   - Analisi dello stato attuale
   - Paletta colori definita
   - Componenti descritti
   - Roadmap implementazione


## 🎨 COME USARE IL DESIGN SYSTEM

### 1. Includere i CSS nei Template

```html
<!-- Base CSS (ordine importante) -->
<link rel="stylesheet" href="/static/css/design-system.css">
<link rel="stylesheet" href="/static/css/components.css">
<link rel="stylesheet" href="/static/css/icons.css">
<link rel="stylesheet" href="/static/css/platform.css">

<!-- Font Google (già configurato) -->
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
```

### 2. Usare Custom Properties per Colori

```css
/* Invece di hardcoded colors, usa le variabili */
color: var(--color-text-primary);
background-color: var(--color-bg-primary);
border-color: var(--color-border-light);
```

### 3. Usare Utility Classes

```html
<!-- Spacing -->
<div class="p-lg">Contenuto</div>
<div class="m-md">Contenuto</div>
<div class="py-lg">Vertical padding</div>

<!-- Display & Flexbox -->
<div class="flex flex-center gap-md">
    <!-- Centra contenuto con gap -->
</div>

<div class="flex-between">
    <!-- Space between -->
</div>

<!-- Text -->
<p class="text-sm text-muted">Testo piccolo muted</p>
<h3 class="text-primary font-bold">Titolo colorato</h3>

<!-- Hide/Show Mobile -->
<div class="hide-mobile">Solo desktop</div>
<div class="hide-desktop">Solo mobile</div>
```

### 4. Componenti Comuni

#### Buttons
```html
<!-- Primary -->
<button class="btn btn-primary btn-lg">Accedi</button>

<!-- Secondary -->
<button class="btn btn-secondary btn-md">Annulla</button>

<!-- Icon Button -->
<button class="btn-icon icon-md">
    <svg>...</svg>
</button>

<!-- Full Width (mobile) -->
<button class="btn btn-primary btn-full">Submit</button>
```

#### Forms
```html
<form>
    <div class="form-group">
        <label for="email">Email <span class="required">*</span></label>
        <input type="email" id="email" placeholder="nome@esempio.it" required/>
        <small class="form-text">Usiamo la tua email per login</small>
    </div>
    
    <div class="form-group has-error">
        <label for="password">Password</label>
        <input type="password" id="password" required/>
        <small class="form-error">Password non valida</small>
    </div>
    
    <div class="form-check">
        <input type="checkbox" id="remember"/>
        <label for="remember">Ricordami</label>
    </div>
</form>
```

#### Cards
```html
<div class="card">
    <div class="card-header">
        <h3>Titolo Card</h3>
        <span class="badge badge-primary">Label</span>
    </div>
    
    <div class="card-body">
        Contenuto
    </div>
    
    <div class="card-footer">
        <button class="btn btn-secondary">Annulla</button>
        <button class="btn btn-primary">OK</button>
    </div>
</div>
```

#### Alerts
```html
<div class="alert alert-success">
    ✓ Operazione completata con successo!
</div>

<div class="alert alert-danger">
    ✗ Si è verificato un errore.
</div>

<div class="alert alert-warning">
    ⚠ Attenzione: azione irreversibile.
</div>
```

#### Grid Responsive
```html
<div class="grid grid-2">
    <div>Colonna 1</div>
    <div>Colonna 2</div>
</div>

<!-- Breakpoint responsive -->
<div class="grid grid-lg-3">
    <div>Su desktop: 3 colonne</div>
    <div>Su tablet: 2 colonne (via media query)</div>
    <div>Su mobile: 1 colonna</div>
</div>
```

#### Icons
```html
<!-- Inline SVG con classe icon -->
<span class="icon icon-md icon-primary">
    <svg viewBox="0 0 24 24">...</svg>
</span>

<!-- Icon con testo -->
<span class="icon-text">
    <svg class="icon icon-sm">...</svg>
    <span>Testo accanto</span>
</span>

<!-- Icon animato (loading) -->
<span class="icon icon-spin">
    <svg>...</svg>
</span>
```

### 5. Inclusione Icone nei Template

Nel file `static/icons-library.html` sono disponibili tutti gli SVG.
Per inserirli nei template Go:

```html
<!-- Opzione 1: Copiare l'SVG direttamente -->
<button class="btn-icon">
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
        <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
        <circle cx="12" cy="10" r="3"/>
    </svg>
</button>

<!-- Opzione 2: Usare data URL (per icone statiche) -->
<!-- Preprocessare SVG in data URL e salvare in variabili -->
```


## 🔧 MODIFICA TEMPLATE ESISTENTI

### Passo 1: Aggiornare HTML Head
```html
<!-- Rimuovere -->
<link rel="stylesheet" href="/static/css/style.css">

<!-- Aggiungere -->
<link rel="stylesheet" href="/static/css/design-system.css">
<link rel="stylesheet" href="/static/css/components.css">
<link rel="stylesheet" href="/static/css/icons.css">
```

### Passo 2: Aggiornare Classi CSS
Cercare e sostituire:
- `.header` → `.navbar` + `.menu-header`
- `.btn-primary` → `.btn btn-primary`
- `.container` → `.container` (già supporta responsive)
- Inline `<style>` → Consolidare in CSS files

### Passo 3: Aggiungere Icone
Dove necessario, sostituire emoji/testo con SVG icons:
```html
<!-- Prima -->
<button>📞 Chiama</button>

<!-- Dopo -->
<button class="btn btn-primary">
    <svg class="icon icon-md"><!-- SVG phone icon --></svg>
    <span>Chiama</span>
</button>
```

### Passo 4: Testare su Mobile
- Verificare touch targets ≥ 44px ✓
- Verificare viewport scale corretta ✓
- Testare su iOS + Android ✓


## 📱 TESTING CHECKLIST

### Desktop (1920px+)
- [ ] Layout multi-colonna funziona
- [ ] Hover effects visibili
- [ ] Spacing coerente

### Tablet (1024px - 640px)
- [ ] Layout responsive 2-colonne
- [ ] Tab bar nascosto
- [ ] Touch targets funzionali

### Mobile (< 640px)
- [ ] Layout single-colonna
- [ ] Touch targets ≥ 44px
- [ ] Bottom tab bar visibile
- [ ] Nessun overflow orizzontale
- [ ] Fontsize ≥ 16px (prevent iOS zoom)

### Accessibility
- [ ] Color contrast ≥ 4.5:1
- [ ] Keyboard navigation funziona
- [ ] Focus outline visibile
- [ ] Screen reader test
- [ ] Reduced motion rispettato


## 🌙 DARK MODE IMPLEMENTATION

Il design system supporta dark mode via `prefers-color-scheme`.

Per testare in browser:
```javascript
// Chrome DevTools → Rendering → Emulate CSS media feature
// prefers-color-scheme: dark
```

Le variabili CSS cambiano automaticamente:
```css
@media (prefers-color-scheme: dark) {
  :root {
    --color-text-primary: #F3F4F6;
    --color-bg-primary: #111827;
    /* ... */
  }
}
```


## 🚀 PROSSIMI PASSI

### Fase 1: Foundation (1 settimana)
- [x] Creare design-system.css
- [x] Creare components.css
- [x] Creare icons.css
- [ ] Consolidare style.css con design-system.css

### Fase 2: Template Redesign (2 settimane)
- [ ] Redesignare admin.html
- [ ] Redesignare create_menu.html
- [ ] Redesignare edit_menu.html
- [ ] Redesignare select_restaurant.html
- [ ] Redesignare analytics_dashboard.html

### Fase 3: Optimization (1 settimana)
- [ ] Mobile testing completo
- [ ] Accessibility audit
- [ ] Performance optimization
- [ ] Dark mode fine-tuning

### Fase 4: Deployment
- [ ] Staging testing
- [ ] Browser compatibility testing
- [ ] Production rollout


## 📊 METRICHE PERFORMANCE

Target:
- CSS file size: < 50KB gzipped ✓ (attualmente ~7KB)
- Lighthouse Performance: > 90
- Mobile FCP: < 1s
- Mobile LCP: < 2.5s
- CLS: < 0.1


## 🎓 RIFERIMENTI & RISORSE

Design Principles:
- https://www.apple.com/it/
- https://developer.apple.com/design/human-interface-guidelines/
- https://getbootstrap.com/docs/5.3/

Performance:
- https://web.dev/responsive-web-design-basics/
- https://web.dev/web-vitals/

Accessibility:
- https://www.w3.org/WAI/WCAG21/quickref/
- https://www.a11y-101.com/

SVG Icons:
- https://heroicons.com/
- https://www.svgrepo.com/


## 📝 NOTE IMPORTANTI

1. **No Bootstrap Dependency**: Il design system è completamente custom, nessuna dipendenza esterna oltre a Inter font

2. **CSS Custom Properties**: Tutto usa variabili, facile da customizzare

3. **Mobile-First**: Tutti i breakpoints partono da mobile

4. **Accessibility First**: WCAG AA compliance built-in

5. **PWA Ready**: Safe area insets, manifest.json supportato

6. **Dark Mode Built-in**: Funziona automaticamente

7. **Performance**: File sizes ottimizzati per PWA


## ❓ FAQ

**D: Come cambio i colori primari?**
R: Modifica `:root { --color-primary: #XXXXXX; }` in design-system.css

**D: Come aggiungo nuove icone?**
R: Aggiungi SVG in icons-library.html e usa inline nei template

**D: Posso usare Bootstrap insieme?**
R: Sconsigliato - va in conflitto. Usa solo il design system custom.

**D: Come testo dark mode?**
R: Chrome DevTools → Rendering → prefers-color-scheme: dark

**D: I CSS sono retrocompatibili con old browsers?**
R: CSS Custom Properties richiedono ES6 browsers (IE11 not supported)


---
Documento: 2026-05-10
Versione: 1.0
