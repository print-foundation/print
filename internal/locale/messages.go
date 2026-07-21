package locale

import "fmt"

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

var en = map[string]string{
	"app.name":           "Pr!nt",
	"action.install":     "Install",
	"action.cancel":      "Cancel",
	"action.back":        "Back",
	"action.confirm":     "I understand, write to %s",
	"wizard.select_os":   "Select an operating system",
	"wizard.select_disk": "Select a target disk",
	"wizard.confirm":     "Confirm write to %s",
	"wizard.progress":    "Progress",
	"warn.destructive":   "This will erase all data on %s.",
	"error.verification": "Verification failed: %v",
	"error.no_target":    "No target disk selected.",
	"status.downloading": "Downloading %s",
	"status.verifying":   "Verifying image",
	"status.writing":     "Writing to disk",
	"status.done":        "Installation complete",
}

var de = map[string]string{
	"app.name":           "Pr!nt",
	"action.install":     "Installieren",
	"action.cancel":      "Abbrechen",
	"action.back":        "ZurÃ¼ck",
	"action.confirm":     "Ich verstehe, schreibe auf %s",
	"wizard.select_os":   "Betriebssystem auswÃ¤hlen",
	"wizard.select_disk": "Zielplatte auswÃ¤hlen",
	"wizard.confirm":     "Schreiben auf %s bestÃ¤tigen",
	"wizard.progress":    "Fortschritt",
	"warn.destructive":   "Dies lÃ¶scht alle Daten auf %s.",
	"error.verification": "ÃœberprÃ¼fung fehlgeschlagen: %v",
	"error.no_target":    "Keine Zielplatte ausgewÃ¤hlt.",
	"status.downloading": "Lade %s herunter",
	"status.verifying":   "Image wird geprÃ¼ft",
	"status.writing":     "Schreiben auf Platte",
	"status.done":        "Installation abgeschlossen",
}

var es = map[string]string{
	"app.name":           "Pr!nt",
	"action.install":     "Instalar",
	"action.cancel":      "Cancelar",
	"action.back":        "AtrÃ¡s",
	"action.confirm":     "Entiendo, escribir en %s",
	"wizard.select_os":   "Elige un sistema operativo",
	"wizard.select_disk": "Elige un disco de destino",
	"wizard.confirm":     "Confirmar escritura en %s",
	"wizard.progress":    "Progreso",
	"warn.destructive":   "Esto borrarÃ¡ todos los datos en %s.",
	"error.verification": "VerificaciÃ³n fallida: %v",
	"error.no_target":    "NingÃºn disco de destino seleccionado.",
	"status.downloading": "Descargando %s",
	"status.verifying":   "Verificando imagen",
	"status.writing":     "Escribiendo en disco",
	"status.done":        "InstalaciÃ³n completa",
}

var ja = map[string]string{
	"app.name":           "Pr!nt",
	"action.install":     "ã‚¤ãƒ³ã‚¹ãƒˆãƒ¼ãƒ«",
	"action.cancel":      "ã‚­ãƒ£ãƒ³ã‚»ãƒ«",
	"action.back":        "æˆ»ã‚‹",
	"action.confirm":     "%s ã«æ›¸ãè¾¼ã‚€ã“ã¨ã‚’ç†è§£ã—ã¾ã—ãŸ",
	"wizard.select_os":   "OSã‚’é¸æŠž",
	"wizard.select_disk": "æ›¸ãè¾¼ã¿å…ˆã®ãƒ‡ã‚£ã‚¹ã‚¯ã‚’é¸æŠž",
	"wizard.confirm":     "%s ã¸ã®æ›¸ãè¾¼ã¿ã‚’ç¢ºèª",
	"wizard.progress":    "é€²è¡ŒçŠ¶æ³",
	"warn.destructive":   "%s ä¸Šã®ã™ã¹ã¦ã®ãƒ‡ãƒ¼ã‚¿ãŒæ¶ˆåŽ»ã•ã‚Œã¾ã™ã€‚",
	"error.verification": "æ¤œè¨¼ã«å¤±æ•—ã—ã¾ã—ãŸ: %v",
	"error.no_target":    "æ›¸ãè¾¼ã¿å…ˆã®ãƒ‡ã‚£ã‚¹ã‚¯ãŒé¸æŠžã•ã‚Œã¦ã„ã¾ã›ã‚“ã€‚",
	"status.downloading": "%s ã‚’ãƒ€ã‚¦ãƒ³ãƒ­ãƒ¼ãƒ‰ä¸­",
	"status.verifying":   "ã‚¤ãƒ¡ãƒ¼ã‚¸ã‚’æ¤œè¨¼ä¸­",
	"status.writing":     "ãƒ‡ã‚£ã‚¹ã‚¯ã«æ›¸ãè¾¼ã¿ä¸­",
	"status.done":        "ã‚¤ãƒ³ã‚¹ãƒˆãƒ¼ãƒ«å®Œäº†",
}

var fr = map[string]string{
	"app.name":           "Pr!nt",
	"action.install":     "Installer",
	"action.cancel":      "Annuler",
	"action.back":        "Retour",
	"action.confirm":     "J'ai compris, Ã©crire sur %s",
	"wizard.select_os":   "Choisir un systÃ¨me d'exploitation",
	"wizard.select_disk": "Choisir un disque cible",
	"wizard.confirm":     "Confirmer l'Ã©criture sur %s",
	"wizard.progress":    "Progression",
	"warn.destructive":   "Cela effacera toutes les donnÃ©es sur %s.",
	"error.verification": "Ã‰chec de la vÃ©rification : %v",
	"error.no_target":    "Aucun disque cible sÃ©lectionnÃ©.",
	"status.downloading": "TÃ©lÃ©chargement de %s",
	"status.verifying":   "VÃ©rification de l'image",
	"status.writing":     "Ã‰criture sur le disque",
	"status.done":        "Installation terminÃ©e",
}
