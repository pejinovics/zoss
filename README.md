# Real-time Collaborative Document Editor

## Domen problema

### Kratak opis domena

Sistem je platforma za kolaborativno uređivanje dokumenata u realnom vremenu, namenjena pojedincima i timovima u okviru organizacija. Omogućava istovremeni rad više korisnika nad istim sadržajem uz deljenje pristupa, komentarisanje i pregled istorije izmena. Pored toga, podržava pretragu i uvoz/izvoz dokumenata, kao i organizaciju rada kroz radne prostore (workspaces) i administraciju korisnika i prava pristupa.

### Akteri i njihove uloge

- **Korisnik (Individual/Team Member)**: kreira i uređuje dokumente, deli sadržaj, sarađuje u realnom vremenu, koristi komentare i istoriju izmena.
- **Gost (Guest User)**: pristupa deljenom dokumentu sa ograničenim pravima (view/comment) preko linka ili pozivnice.
- **Administrator radnog prostora (Workspace Admin)**: upravlja članovima i ulogama, pravilima pristupa (npr. obavezna MFA), politikama zadržavanja podataka i audit logovima.
- **Eksterni servisi (Integrations)**: SSO provajderi (OIDC/SAML), skladišta fajlova (npr. Drive/Dropbox sync), email/push provajderi.
### Poslovni procesi

1. **Autentifikacija i upravljanje sesijama** (email/SSO, MFA, session management).
2. **Životni ciklus dokumenta** (kreiranje, čuvanje, organizacija, arhiviranje/brisanje).
3. **Real-time kolaboracija** (simultano uređivanje, prisustvo korisnika, rešavanje konflikata).
4. **Deljenje i prava pristupa** (owner/editor/commenter/viewer, share link sa istekom/lozinkom, revokacija).
5. **Komentari i diskusije** (thread-ovi, mention, notifikacije).
6. **Istorija i verzije** (snapshot, diff/restore, audit trail).
7. **Pretraga i otkrivanje sadržaja** (full-text search, filteri).
8. **Uvoz/izvoz i integracije** (PDF/DOCX/Markdown export, sync sa storage servisima).
9. **Upravljanje podacima** (retention, audit, export/brisanje podataka).
## 2. Arhitektura sistema

### Karakteristike

Sistem je zamišljen kao mikroservisna, event-driven arhitektura fokusirana na real-time komunikaciju i skaliranje. Ključne karakteristike su:

- **WebSocket-first** komunikacija za kolaboraciju u realnom vremenu.
- **CRDT** pristup za rešavanje konflikata pri istovremenom uređivanju i podršku za offline→online sinhronizaciju.
- **Event log / audit trail** za praćenje izmena i vraćanje na prethodne verzije (snapshot + istorija operacija).
- **Kombinacija više tipova skladišta**: sadržaj dokumenata, metapodaci, sesije i pretraga čuvaju se u različitim sistemima prilagođenim toj nameni.
- **Asinhroni poslovi** (export, indeksiranje, notifikacije) preko message broker-a.
- **API Gateway** kao jedinstvena ulazna tačka (TLS, rate limiting, JWT validacija, WebSocket proxy).
### Tehnologije

- **Web klijent (React + TypeScript + CRDT biblioteka npr. Yjs/Automerge)** - UI i lokalni editor state, CRDT omogućava merge bez konflikata i rad uz povremene prekide veze.
- **Collaboration servis (Elixir + Phoenix Channels)** - upravljanje WebSocket konekcijama, distribucija edit operacija i presence informacija
- **Document/Metadata servis (Go + gRPC)** - rad sa dokumentima, čuvanje snapshot-a i metapodataka, inter-service komunikacija, Go je dobar za performanse i konkurentnost.
- **NoSQL skladište sadržaja (Cassandra)** - skladištenje snapshot-a/istorije dokumenta i write-heavy obrazaca; pogodna za horizontalno skaliranje i replikaciju.
- **Relaciona baza (PostgreSQL)** - korisnici, radni prostori, prava pristupa i strukture (npr. folderi); ACID i integritet su bitni za autorizaciju.
- **Redis** - sesije, rate limiting brojači i keš
- **Elasticsearch/OpenSearch**  - full-text pretraga i indeksiranje sadržaja.
- **RabbitMQ** - pozadinski poslovi (export, indeksiranje, notifikacije) i pouzdana isporuka zadataka.
- **MinIO (S3-compatible storage)** - fajlovi i prilozi (slike/attachment), kao i generisani export-i (PDF/DOCX).
- **Keycloak (OIDC/SAML)** - autentifikacija, SSO i MFA

## 3. Grupe slučajeva korišćenja

1. **Upravljanje dokumentima**: kreiranje, organizacija (folder/tag), arhiviranje/brisanje, omiljeni/recent.
2. **Real-time uređivanje**: simultano editovanje, presence, live cursors, offline→online sinhronizacija.
3. **Deljenje i pristup**: pozivnice, share link (istek/lozinka), role-based prava, revokacija pristupa.
4. **Komentari i saradnja**: komentari u thread-ovima, mention, notifikacije, resolve/reopen.
5. **Verzije i istorija**: pregled istorije, snapshot, diff i restore na prethodnu verziju.
6. **Pretraga**: full-text search, filteri (autor/datum/tag), autocomplete i isticanje rezultata.
7. **Uvoz/izvoz**: import postojećih fajlova, export u PDF/DOCX/Markdown, batch export.
8. **Administracija workspace-a**: članovi, role, SSO konfiguracija, audit logovi, retention politike.

## 4. Osetljivi resursi i bezbednosni ciljevi

| #   | Resurs                                             | Opis                                                       | Bezbednosni cilj                                      |
| --- | -------------------------------------------------- | ---------------------------------------------------------- | ----------------------------------------------------- |
| 1   | **Sadržaj dokumenata**                             | Tekst, formatiranje, prilozi, ugrađeni mediji              | **C/I/A** (neovlašćen pristup, izmene, dostupnost)    |
| 2   | **Autentifikacioni podaci i sesije**               | Hash lozinki, JWT/opaque tokeni, refresh tokeni, MFA tajne | **C/I/A** (krađa tokena, session hijack, brute force) |
| 3   | **Matrica prava pristupa**                         | Uloge, prava, share-link tokeni i pravila isteka           | **I/A** (sprečiti privilege escalation i IDOR)        |
| 4   | **Istorija izmena / event log**                    | Operacije, snapshot-i, autorstvo izmena, audit trail       | **I/A** + neporicanje (tampering, replay)             |
| 5   | **Lični podaci i podaci o aktivnostima korisnika** | Email, ime, IP, uređaji, zapisi o pristupu dokumentima     | **C** + minimizacija + brisanje (**GDPR**)            |
| 6   | **Kriptografski ključevi i secrets**               | JWT ključevi, enkripcioni ključevi, OAuth i API secrets    | **C/I/A** (curenje, zamena, rotation bez prekida)     |
| 7   | **Audit logovi sistema**                           | Zapisi o prijavama, deljenju dokumenata i admin akcijama   | **C/I/A** (forenzika, integritet zapisa)              |
| 8   | **Tokeni za deljenje dokumenata**                  | Javni linkovi sa pravima, istekom i opcionalnom lozinkom   | **C/I/A** (neovlašćen pristup, curenje linka)         |

