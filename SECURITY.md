VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

### Befund 1: Unbekannte Routen und nicht registrierte Methoden liefern text/plain statt JSON (AC-13, AC-14)

- **Schweregrad:** mittel
- **Betroffene Stelle:** `internal/api/router.go` (ServeMux ohne JSON-Fallback) in Verbindung mit `main.go`
- **Beschreibung:** Der Router registriert ausschließlich die in der Spec definierten Methoden- und Pfadmuster. Ein Request auf einen unbekannten Pfad wie `/unbekannt` oder eine nicht registrierte Methode wie `POST /healthz` wird vom `http.ServeMux` selbst mit Status 404 bzw. 405 im Klartextformat (`text/plain`) beantwortet. AC-13 verlangt für alle JSON-Antworten – auch Fehlerantworten – den Header `Content-Type: application/json`; AC-14 verlangt für Fehlerantworten ausschließlich das definierte JSON-Fehlerobjekt `{"error": "..."}`. Diese Vorgaben werden für solche Router-Fehlerfälle verletzt. Es werden keine internen Details preisgegeben, die Fehlerbehandlung ist jedoch lückenhaft und nicht konform zu den Security-Kriterien.
- **Konkrete Lösung:** Einen zentralen JSON-Fallback für unbekannte Pfade in `NewRouter` registrieren:
    ```go
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        writeError(w, http.StatusNotFound, "not found")
    })
    ```
    Für Methodenverstöße (`405 Method Not Allowed`) genügt der reine ServeMux-Fallback nicht. Hier kann ein schlanker Wrapper um den Mux die Antwort eines gepufferten `ResponseWriter` prüfen und bei `404`/`405` stattdessen `writeError` mit dem passenden Statuscode ausgeben. Alternativ lassen sich für jede Ressource alle nicht erlaubten Methoden explizit registrieren und mit `writeError(w, http.StatusMethodNotAllowed, "method not allowed")` beantworten.

## Weitere Hinweise (nicht Bestandteil des Verdicts)

- **Fehlende Authentifizierung/Autorisierung:** Die API ist vollständig offen. Die Spec enthält dazu kein Kriterium; für einen produktiven Einsatz wären Zugriffsschutz, TLS und Rate-Limiting erforderlich.
- **Fehlende HTTP-Server-Timeouts:** `main.go` nutzt `http.ListenAndServe` ohne explizite `ReadTimeout`, `WriteTimeout` und `IdleTimeout`. Dadurch ist der Dienst anfällig für langsame Verbindungen bzw. Ressourcenbindung (z. B. Slowloris).
- **Unbegrenztes Speicherwachstum:** Der In-Memory-Store begrenzt weder die Anzahl der Flags noch die Länge einzelner Felder. Ein anonymer Client kann durch wiederholtes Anlegen großer Flags Speicher erschöpfen. Eine Begrenzung der Flag-Anzahl und/oder max. Feldlängen wäre sinnvoll.