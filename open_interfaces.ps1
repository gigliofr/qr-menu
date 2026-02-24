# Script per aprire rapidamente tutte le interfacce QR Menu
# Esegui: .\open_interfaces.ps1

$baseUrl = "http://localhost:8080"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    Apertura Interfacce QR Menu" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Controlla se il server è attivo
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/menus" -Method GET -TimeoutSec 3
    Write-Host "✅ Server attivo su $baseUrl" -ForegroundColor Green
} catch {
    Write-Host "❌ Server non raggiungibile!" -ForegroundColor Red
    Write-Host "   Avvia il server con: .\start.bat" -ForegroundColor Yellow
    Write-Host "   Oppure: .\qr-menu.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "🌐 Apertura interfacce web..." -ForegroundColor Yellow

# Ottiene la lista dei menu per aprire un esempio
$menus = Invoke-RestMethod -Uri "$baseUrl/api/menus" -Method GET
$sampleMenuId = $null
if ($menus.PSObject.Properties.Count -gt 0) {
    $sampleMenuId = ($menus.PSObject.Properties | Select-Object -First 1).Value.id
}

# Apri le interfacce principali
Write-Host "   📋 Admin Interface..." -ForegroundColor Cyan
Start-Process "http://localhost:8080/admin"

Start-Sleep -Seconds 2

Write-Host "   ➕ Create Menu Form..." -ForegroundColor Cyan  
Start-Process "http://localhost:8080/admin/menu/create"

if ($sampleMenuId) {
    Start-Sleep -Seconds 2
    Write-Host "   👁️  Sample Public Menu..." -ForegroundColor Cyan
    Start-Process "http://localhost:8080/menu/$sampleMenuId"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "           Interfacce Aperte!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📱 URLs principali:" -ForegroundColor White
Write-Host "   Admin: $baseUrl/admin" -ForegroundColor Cyan
Write-Host "   Create Menu: $baseUrl/admin/menu/create" -ForegroundColor Cyan
if ($sampleMenuId) {
    Write-Host "   Sample Menu: $baseUrl/menu/$sampleMenuId" -ForegroundColor Cyan
}
Write-Host ""
Write-Host "🔗 API Endpoints:" -ForegroundColor White  
Write-Host "   GET  /api/menus" -ForegroundColor Cyan
Write-Host "   POST /api/menu" -ForegroundColor Cyan
Write-Host "   GET  /api/menu/{id}" -ForegroundColor Cyan
Write-Host "   POST /api/menu/{id}/generate-qr" -ForegroundColor Cyan
Write-Host ""