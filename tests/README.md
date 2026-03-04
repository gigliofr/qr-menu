# Test Scripts

Questa cartella contiene gli script di test e diagnostica per il progetto QR Menu.

## Script Principali (Da mantenere)

### Test Completo
- **`test_completo_fix.ps1`** - Test end-to-end completo
  - Registrazione nuovo account
  - Login
  - Creazione menu
  - Verifica visibilità menu
  - Test debug endpoint
  - Analytics

### Setup & Diagnostica
- **`setup_ristorante_completo.ps1`** - Crea account di test con menu completi
- **`verifica_api_menu.ps1`** - Debug approfondito API menu
  
## Script Legacy (Da eliminare)

I seguenti script sono stati creati durante il debugging del bug dei tag bson e non sono più necessari:

- `check_debug.ps1` - Sostituito da `verifica_api_menu.ps1`
- `diagnostic_test.ps1` - Obsoleto
- `migrate_bson_fields.ps1` - Migrazione già eseguita
- `monitor_deploy.ps1` - Non più necessario
- `test_api_direct.ps1` - Duplicato
- `test_app_railway.ps1` - Duplicato
- `test_completo_funzionalita.ps1` - Sostituito da `test_completo_fix.ps1`
- `test_completo_railway.ps1` - Duplicato
- `test_con_account_esistente.ps1` - Obsoleto
- `test_create_menu.ps1` - Funzionalità inclusa in test completo
- `test_debug_endpoint.ps1` - Sostituito da `verifica_api_menu.ps1`
- `test_immediate_fix.ps1` - Obsoleto (fix completato)
- `test_menu_visibility.ps1` - Obsoleto (bug risolto)
- `verify_deploy_status.ps1` - Non più necessario
- `verify_fix.ps1` - Obsoleto (fix completato)
- `wait_and_check_logs.ps1` - Non più necessario
- `wait_for_deploy.ps1` - Non più necessario

## Uso Raccomandato

Per testare l'applicazione dopo un deploy:

```powershell
# Test completo con nuovo account
.\test_completo_fix.ps1

# Oppure con account esistente
.\setup_ristorante_completo.ps1

# Debug specifico menu
.\verifica_api_menu.ps1
```

## Pulizia

Per eliminare i file obsoleti:

```powershell
# Dalla root del progetto
Remove-Item tests\check_debug.ps1, tests\diagnostic_test.ps1, tests\migrate_bson_fields.ps1, tests\monitor_deploy.ps1, tests\test_api_direct.ps1, tests\test_app_railway.ps1, tests\test_completo_funzionalita.ps1, tests\test_completo_railway.ps1, tests\test_con_account_esistente.ps1, tests\test_create_menu.ps1, tests\test_debug_endpoint.ps1, tests\test_immediate_fix.ps1, tests\test_menu_visibility.ps1, tests\verify_deploy_status.ps1, tests\verify_fix.ps1, tests\wait_and_check_logs.ps1, tests\wait_for_deploy.ps1
```
