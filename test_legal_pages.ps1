# Test Pagine Legali - QR Menu
# Esegui questo script per verificare che tutte le pagine legali siano accessibili

Write-Host "`n🧪 TEST CONFORMITÀ LEGALE ITALIANA`n" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray

$baseUrl = "http://localhost:8080"

# Se Railway è attivo, usa quello
$railwayUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "📍 Testing su: $baseUrl" -ForegroundColor Yellow
Write-Host "(Assicurati che il server sia avviato localmente)`n" -ForegroundColor DarkGray

$pages = @(
    @{ Name = "Privacy Policy"; Path = "/privacy" },
    @{ Name = "Cookie Policy"; Path = "/cookie-policy" },
    @{ Name = "Termini e Condizioni"; Path = "/terms" },
    @{ Name = "Note Legali"; Path = "/legal" }
)

$results = @()

foreach ($page in $pages) {
    Write-Host "Testing: $($page.Name)..." -NoNewline
    
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl$($page.Path)" -Method GET -TimeoutSec 5 -ErrorAction Stop
        
        if ($response.StatusCode -eq 200) {
            Write-Host " ✅ OK" -ForegroundColor Green
            
            # Verifica presenza segnaposto non compilati
            $content = $response.Content
            if ($content -match "\[INSERIRE") {
                Write-Host "   ⚠️  ATTENZIONE: Contiene segnaposto [INSERIRE ...] da compilare!" -ForegroundColor Yellow
                $results += @{ Page = $page.Name; Status = "OK con segnaposto"; Color = "Yellow" }
            } else {
                $results += @{ Page = $page.Name; Status = "✅ Completo"; Color = "Green" }
            }
            
            # Verifica lunghezza minima (pagina non vuota)
            if ($content.Length -lt 1000) {
                Write-Host "   ⚠️  ATTENZIONE: Pagina molto corta ($($content.Length) caratteri)" -ForegroundColor Yellow
            }
        }
    }
    catch {
        Write-Host " ❌ ERRORE" -ForegroundColor Red
        Write-Host "   Errore: $($_.Exception.Message)" -ForegroundColor Red
        $results += @{ Page = $page.Name; Status = "❌ Non accessibile"; Color = "Red" }
    }
}

Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray

Write-Host "📊 RIEPILOGO TEST`n" -ForegroundColor Cyan

foreach ($result in $results) {
    Write-Host "  $($result.Page): " -NoNewline
    Write-Host $result.Status -ForegroundColor $result.Color
}

Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray

# Test footer links
Write-Host "🔗 Testing Footer Links`n" -ForegroundColor Cyan

try {
    $adminResponse = Invoke-WebRequest -Uri "$baseUrl/login" -Method GET -TimeoutSec 5
    
    $footerChecks = @(
        @{ Link = 'href="/privacy"'; Name = "Link Privacy" },
        @{ Link = 'href="/cookie-policy"'; Name = "Link Cookie Policy" },
        @{ Link = 'href="/terms"'; Name = "Link Termini" },
        @{ Link = 'href="/legal"'; Name = "Link Note Legali" }
    )
    
    foreach ($check in $footerChecks) {
        if ($adminResponse.Content -match [regex]::Escape($check.Link)) {
            Write-Host "  ✅ $($check.Name) presente" -ForegroundColor Green
        } else {
            Write-Host "  ❌ $($check.Name) MANCANTE" -ForegroundColor Red
        }
    }
}
catch {
    Write-Host "  ⚠️  Impossibile verificare footer (server non risponde)" -ForegroundColor Yellow
}

Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray

Write-Host "📋 PROSSIMI PASSI:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. ✏️  Compila tutti i segnaposto [INSERIRE ...] nei 4 file HTML" -ForegroundColor Yellow
Write-Host "2. 📧 Configura email reali: privacy@, support@, info@" -ForegroundColor Yellow
Write-Host "3. 🏢 Inserisci P.IVA, indirizzo sede, REA (se società)" -ForegroundColor Yellow
Write-Host "4. 🍪 Implementa Cookie Banner (se usi analytics)" -ForegroundColor Yellow
Write-Host "5. 🚀 Deploy in produzione" -ForegroundColor Green
Write-Host ""
Write-Host "📖 Leggi LEGAL_COMPLIANCE_IT.md per guida completa`n" -ForegroundColor Cyan

# Test Railway (se disponibile)
$testRailway = Read-Host "`nVuoi testare anche su Railway? (s/n)"

if ($testRailway -eq "s" -or $testRailway -eq "S") {
    Write-Host "`n🌐 Testing su Railway: $railwayUrl`n" -ForegroundColor Cyan
    
    foreach ($page in $pages) {
        Write-Host "Testing: $($page.Name)..." -NoNewline
        
        try {
            $response = Invoke-WebRequest -Uri "$railwayUrl$($page.Path)" -Method GET -TimeoutSec 10 -ErrorAction Stop
            
            if ($response.StatusCode -eq 200) {
                Write-Host " ✅ OK" -ForegroundColor Green
            }
        }
        catch {
            Write-Host " ❌ ERRORE: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host "`n✅ Test completati!`n" -ForegroundColor Green
