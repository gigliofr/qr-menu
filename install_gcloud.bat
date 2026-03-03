@echo off
REM Script di setup Google Cloud SDK per Windows

echo.
echo ===============================================
echo  Google Cloud SDK Setup for QR-Menu
echo ===============================================
echo.

REM Controlla se winget è disponibile
where /q winget
if errorlevel 1 (
    echo ❌ winget non trovato. Installazione manuale:
    echo.
    echo 1. Scarica da: https://cloud.google.com/sdk/docs/install
    echo 2. Esegui l'installer
    echo 3. Riavvia PowerShell
    echo 4. Torna qui e esegui di nuovo questo script
    echo.
    exit /b 1
)

echo ✅ winget trovato. Installando Google Cloud SDK...
winget install Google.CloudSDK -e

echo.
echo ✅ Installazione completata!
echo.
echo Prossimi step:
echo 1. Riavvia PowerShell
echo 2. Esegui: gcloud auth login
echo 3. Segui le istruzioni nel browser
echo.
pause
