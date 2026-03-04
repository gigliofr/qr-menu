# Pre-Commit Security Check
# Esegui questo script prima di committare per verificare potenziali segreti

$ErrorActionPreference = "Continue"

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           SECURITY CHECK - Pre-Commit Verification           ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$issues = @()
$warnings = @()

# Pattern da cercare (regex)
$patterns = @{
    "MongoDB URI" = "mongodb(\+srv)?://[^\s]+"
    "Password in plain" = "(password|passwd|pwd)\s*[:=]\s*['\`"][^\s'`"]{8,}"
    "API Key" = "(api[_-]?key|apikey)\s*[:=]\s*['\`"][^\s'`"]{16,}"
    "AWS Key" = "AKIA[0-9A-Z]{16}"
    "GitHub Token" = "gh[ps]_[A-Za-z0-9_]{36,}"
    "Private Key" = "-----BEGIN (RSA |EC |DSA )?PRIVATE KEY-----"
    "Bearer Token" = "Bearer\s+[A-Za-z0-9\-._~+\/]+=*"
    "JWT Token" = "eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*"
}

# File da scansionare (staged changes)
Write-Host "🔍 Scansionando file staged..." -ForegroundColor Yellow

$stagedFiles = git diff --cached --name-only --diff-filter=ACM
if ($stagedFiles) {
    foreach ($file in $stagedFiles) {
        # Salta file binari, immagini, etc
        if ($file -match '\.(exe|dll|png|jpg|jpeg|gif|ico|pdf|zip|tar|gz)$') {
            continue
        }
        
        # Salta file in cartelle ignorate
        if ($file -match '^(vendor|node_modules|.git)/') {
            continue
        }
        
        Write-Host "   📄 $file" -ForegroundColor Gray
        
        # Leggi contenuto
        try {
            $content = git diff --cached $file | Out-String
            
            # Cerca pattern sospetti
            foreach ($patternName in $patterns.Keys) {
                $pattern = $patterns[$patternName]
                if ($content -match $pattern) {
                    $match = $Matches[0]
                    # Nascondi parte del match per sicurezza
                    $maskedMatch = $match.Substring(0, [Math]::Min(20, $match.Length)) + "..."
                    
                    $issues += @{
                        File = $file
                        Type = $patternName
                        Match = $maskedMatch
                    }
                }
            }
            
            # Controlli aggiuntivi
            if ($content -match '(?i)(mongodb.*password|db.*password)') {
                $warnings += @{
                    File = $file
                    Warning = "Possibile riferimento a password database"
                }
            }
            
        } catch {
            Write-Host "      ⚠️ Impossibile leggere: $($_.Exception.Message)" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "   ℹ️ Nessun file staged" -ForegroundColor Gray
}

Write-Host ""
Write-Host "══════════════════════════════════════════════════════════════" -ForegroundColor Gray
Write-Host ""

# Report
if ($issues.Count -gt 0) {
    Write-Host "❌ POTENZIALI SEGRETI TROVATI!" -ForegroundColor Red
    Write-Host ""
    foreach ($issue in $issues) {
        Write-Host "   File: $($issue.File)" -ForegroundColor Yellow
        Write-Host "   Tipo: $($issue.Type)" -ForegroundColor Yellow
        Write-Host "   Match: $($issue.Match)" -ForegroundColor Gray
        Write-Host ""
    }
    
    Write-Host "⚠️ AZIONI RICHIESTE:" -ForegroundColor Yellow
    Write-Host "   1. Rimuovi le credenziali dai file" -ForegroundColor White
    Write-Host "   2. Usa variabili d'ambiente invece" -ForegroundColor White
    Write-Host "   3. Aggiungi file a .gitignore se necessario" -ForegroundColor White
    Write-Host ""
    Write-Host "❌ COMMIT BLOCCATO per sicurezza" -ForegroundColor Red
    Write-Host ""
    exit 1
    
} elseif ($warnings.Count -gt 0) {
    Write-Host "⚠️ AVVISI TROVATI" -ForegroundColor Yellow
    Write-Host ""
    foreach ($warning in $warnings) {
        Write-Host "   File: $($warning.File)" -ForegroundColor Yellow
        Write-Host "   Avviso: $($warning.Warning)" -ForegroundColor Gray
        Write-Host ""
    }
    
    Write-Host "💡 Verifica che non ci siano credenziali reali" -ForegroundColor Cyan
    Write-Host ""
    
    # Chiedi conferma
    $response = Read-Host "Procedere con il commit? (s/n)"
    if ($response -ne 's' -and $response -ne 'S') {
        Write-Host "❌ Commit annullato" -ForegroundColor Red
        exit 1
    }
    
} else {
    Write-Host "✅ NESSUN SEGRETO RILEVATO" -ForegroundColor Green
    Write-Host ""
    Write-Host "   Puoi procedere con il commit in sicurezza" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "══════════════════════════════════════════════════════════════" -ForegroundColor Gray
Write-Host ""

exit 0
