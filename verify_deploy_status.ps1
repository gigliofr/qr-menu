# Verifica stato deploy Railway e confronto con commit

Write-Host "🔍 Verifica Deploy Railway" -ForegroundColor Cyan
Write-Host "=" * 50 -ForegroundColor Gray

# 1. Ultimo commit locale
Write-Host ""
Write-Host "📦 Ultimo commit locale:" -ForegroundColor Yellow
git log -1 --pretty=format:"%h - %s (%cr)" main

# 2. Verifica che Railway abbia l'ultimo codice
Write-Host ""
Write-Host ""
Write-Host "🌐 Versione su Railway:" -ForegroundColor Yellow
$health = Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/health"
Write-Host "   Version: $($health.data.version)"
Write-Host "   Timestamp: $($health.data.timestamp)"

# 3. Controlla build logs su Railway
Write-Host ""
Write-Host "⚠️  IMPORTANTE:" -ForegroundColor Red
Write-Host "   Railway deve aver completato il rebuild dopo il commit 84295b3" -ForegroundColor Yellow
Write-Host "   Se il deploy è ancora in corso o non è partito:" -ForegroundColor Yellow
Write-Host ""
Write-Host "   1. Vai su: https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab" -ForegroundColor White
Write-Host "   2. Verifica che l'ultimo deploy sia del commit 84295b3" -ForegroundColor White
Write-Host "   3. Se il deploy è fallito o non è partito, clicca 'Redeploy'" -ForegroundColor White
Write-Host ""

# 4. Test empirico: verifica se i nuovi tag bson sono attivi
Write-Host "🧪 Test empirico: controllo se il bug è ancora presente..." -ForegroundColor Cyan
Write-Host "   Se il menu appena creato non è visibile in GET = deploy vecchio" -ForegroundColor Gray
Write-Host ""

$confirm = Read-Host "Vuoi forzare un redeploy su Railway? (s/n)"
if ($confirm -eq 's' -or $confirm -eq 'S') {
    Write-Host ""
    Write-Host "✅ Per forzare il redeploy:" -ForegroundColor Green
    Write-Host "   1. Apri: https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab" -ForegroundColor White
    Write-Host "   2. Seleziona il servizio 'qr-menu'" -ForegroundColor White
    Write-Host "   3. Tab 'Deployments'" -ForegroundColor White
    Write-Host "   4. Clicca sui tre puntini (...) del deploy più recente" -ForegroundColor White
    Write-Host "   5. Seleziona 'Redeploy'" -ForegroundColor White
    Write-Host ""
    Write-Host "   Oppure crea un commit vuoto per triggherare il deploy:" -ForegroundColor White
    Write-Host "   git commit --allow-empty -m 'Trigger Railway redeploy'" -ForegroundColor Gray
    Write-Host "   git push" -ForegroundColor Gray
}
