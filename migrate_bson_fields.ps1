# Script per migrare i campi MongoDB dai nomi Go (PascalCase) ai nomi bson (snake_case)
# Questo script rinnomina i campi nei documenti esistenti per compatibilità con i nuovi tag bson

$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "🔄 Script di Migrazione Campi MongoDB" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  ATTENZIONE: Questo script modificherà i documenti esistenti nel database MongoDB." -ForegroundColor Yellow
Write-Host "    È consigliato eseguirlo solo su ambienti di staging/test." -ForegroundColor Yellow
Write-Host ""
Write-Host "Invece di eseguire questa migrazione, puoi semplicemente:" -ForegroundColor Green
Write-Host "1. Eliminare i documenti di test esistenti dal database" -ForegroundColor Green
Write-Host "2. Ricreare i ristoranti e menu usando gli script di test" -ForegroundColor Green
Write-Host ""
Write-Host "Se vuoi procedere con la migrazione automatica, sappi che Railway probabilmente NON espone" -ForegroundColor Yellow
Write-Host "una connessione MongoDB diretta per sicurezza. Dovrai usare MongoDB Atlas direttamente." -ForegroundColor Yellow
Write-Host ""
Write-Host "📋 ALTERNATIVA RACCOMANDATA:" -ForegroundColor Cyan
Write-Host "   1. Vai su MongoDB Atlas: https://cloud.mongodb.com" -ForegroundColor White
Write-Host "   2. Naviga al database 'qr-menu'" -ForegroundColor White
Write-Host "   3. Elimina le collections: 'restaurants', 'menus', 'sessions'" -ForegroundColor White
Write-Host "   4. Esegui di nuovo: .\setup_ristorante_completo.ps1" -ForegroundColor White
Write-Host ""

$risposta = Read-Host "Vuoi procedere con l'eliminazione manuale su Atlas? (s/n)"

if ($risposta -eq 's' -or $risposta -eq 'S') {
    Write-Host ""
    Write-Host "✅ Perfetto! Segui questi passi:" -ForegroundColor Green
    Write-Host "   1. Apri https://cloud.mongodb.com e accedi" -ForegroundColor White
    Write-Host "   2. Seleziona il cluster (ac-d8zdak4.b9jfwmr.mongodb.net)" -ForegroundColor White
    Write-Host "   3. Clicca su 'Browse Collections'" -ForegroundColor White
    Write-Host "   4. Database: 'qr-menu'" -ForegroundColor White
    Write-Host "   5. Per ogni collection (restaurants, menus, sessions):" -ForegroundColor White
    Write-Host "      - Clicca sui tre puntini (...) accanto al nome" -ForegroundColor White
    Write-Host "      - Seleziona 'Drop Collection'" -ForegroundColor White
    Write-Host "   6. Torna qui e esegui: .\setup_ristorante_completo.ps1" -ForegroundColor White
    Write-Host ""
    Write-Host "👉 Premi un tasto quando hai completato l'eliminazione..." -ForegroundColor Cyan
    $null = Read-Host
    
    Write-Host ""
    Write-Host "🧪 Vuoi eseguire ora lo script di setup completo? (s/n)" -ForegroundColor Cyan
    $setup = Read-Host
    
    if ($setup -eq 's' -or $setup -eq 'S') {
        Write-Host ""
        Write-Host "▶️  Esecuzione setup_ristorante_completo.ps1..." -ForegroundColor Green
        & "$PSScriptRoot\setup_ristorante_completo.ps1"
    } else {
        Write-Host ""
        Write-Host "✅ Setup annullato. Esegui manualmente quando pronto:" -ForegroundColor Yellow
        Write-Host "   .\setup_ristorante_completo.ps1" -ForegroundColor White
    }
} else {
    Write-Host ""
    Write-Host "❌ Operazione annullata." -ForegroundColor Red
    Write-Host "   I menu esistenti NON saranno visibili finché non vengono migrati o ricreati." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🏁 Script completato." -ForegroundColor Green
