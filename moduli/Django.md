## Uvod

Django je web framework visoke produktivnosti napisan u Python-u, koji je prvi put objavljen 2005. godine. Razvijen je u novinskoj kući Lawrence Journal-World kako bi se ubrzao razvoj veb aplikacija. Framework prati MTV (Model-Template-View) arhitekturu, što je varijacija poznatog MVC obrazca.

Osnovna filozofija Django-a je "batteries included" pristup. Dolazi sa svim potrebnim komponentama za izgradnju kompletnih web aplikacija. Administratorski panel, ORM sistem, autentifikacija, sesije i druge funkcionalnosti su ugrađene i spremne za upotrebu odmah nakon instalacije.

U Djanga-u je naglašava princip DRY (Don't Repeat Yourself) i omogućava fokusiranje na jedinstvene aspekte aplikacija umesto da iznova pišu već rešene probleme.

Jedna od ključnih prednosti Django-a je njegova skalabilnost. Koriste ga kompanije poput Instagram-a, Pinterest-a, Mozilla-e i NASA-e za aplikacije koje opslužuju milione korisnika dnevno.

## Arhitektura Django framework-a

### Jezgro sistema

**WSGIHandler / ASGIHandler**
- Centralni koordinator - WSGIHandler/ASGIHandler je glavna struktura koja upravlja celokupnim životnim ciklusom HTTP zahteva 
- Konfiguracija - Učitava settings modul i inicijalizuje sve komponente framework-a 
- Request factory - Kreira HttpRequest objekat iz WSGI/ASGI okruženja Funkcionalnosti:

	- Inicijalizacija i pokretanje HTTP servera
	- Učitavanje middleware stack-a
	- Signaliziranje pre i posle obrade zahteva
	- Rukovanje exception-ima i error responses
### Routing podsistem

URLResolver Implementacija URL routing sistema koji mapira URL pattern-e na view funkcije/klase Karakteristike:

- Podrška za path converters
- Named URL patterns za reverse lookup
- Namespace podrška za organizaciju ruta po aplikacijama
- Lazy kompilacija regex pattern-a
- Caching rezolucije za bolje performanse

**URLconf** 
- Konfiguracija ruta - urlpatterns lista definiše sve rute u aplikaciji
- Hijerarhijska struktura - include() omogućava ugneždavanje URL konfiguracija
- Pattern matching - Koristi path() za jednostavne pattern-e i re_path() za regex
- Resolver chain - URLResolver rekurzivno pretražuje ugneždene URLconf-e dok ne pronađe match

### Middleware podsistem

**Middleware stack**
- Funkcionalni wrapper - Svaki middleware je callable koji prima get_response funkciju 
- Onion architecture - Zahtev prolazi "unutra" kroz middleware-e i odgovor "nazad" kroz iste 

**Ugrađeni middleware-i**
 - SecurityMiddleware - Postavlja sigurnosne HTTP headers (HSTS, X-Content-Type-Options, X-Frame-Options) 
 - SessionMiddleware - Upravlja session podacima kroz cookie-based ili database-backed storage 
 - CommonMiddleware - URL rewriting, dodavanje trailing slash, APPEND_SLASH logika 
 - CsrfViewMiddleware - CSRF token validacija za POST/PUT/DELETE zahteve 
 - AuthenticationMiddleware - Vezuje user objekat za request 
 - MessageMiddleware - Privremeno skladištenje poruka za korisnika između zahteva 
 - XFrameOptionsMiddleware - Clickjacking zaštita

### ORM podsistem

**QuerySet** 
- Method chaining - Metode poput filter(), exclude(), order_by() vraćaju novi QuerySet 
- Caching - Rezultati se kešira nakon prvog izvršenja 
- Slicing - Podrška za Python slice sintaksu za LIMIT/OFFSET upite

*Model layer*
- Meta klasa - Definiše database tabelu, ordering, indexes, constraints
- Field types - CharField, IntegerField, ForeignKey, ManyToManyField, itd.
- Model methods - Custom metode za biznis logiku 
- Model Manager - Default manager (objects) za database upite
- Custom managers - Omogućava kreiranje custom query metoda

Database router Multi-database omogućava rad sa više baza podataka istovremeno 

**SQL Compiler**
- Query compilation - Prevodi QuerySet operacije u SQL upite specifične za database backend
- Database abstraction - Različiti baze (PostgreSQL, MySQL, SQLite, Oracle) imaju različite compiler-e 
- Query optimization - Automatsko korišćenje JOIN-ova, SELECT_RELATED, PREFETCH_RELATED 
- Parameter escaping - Automatski escaping parametara za zaštitu od SQL injection-a
### Template podsistem

**Template Engine**
- Loader - Učitava template fajlove iz fajl sistema
- Parser - Parsira sintaksu u Node tree strukturu i generiše finalni output 
- Context - Dictionary-like objekat sa varijablama dostupnim u template-u 
- RequestContext - Specijalna verzija koja automatski uključuje request objekat i context processors

**Auto-escaping**
- Sve varijable se automatski escape-uju za HTML kao zaštita od XSS napada 
- Safe označavanje kroz safe filter i mark_safe() za provereno bezbedan HTML JavaScript i CSS konteksti zahtevaju različit escaping
### Static files podsistem

**Static files handling** 
- STATIC_ROOT - Direktorijum gde collectstatic sakuplja sve statičke fajlove 
- STATIC_URL - URL prefix za pristup statičkim fajlovima 
- STATICFILES_DIRS - Dodatni direktorijumi za statičke fajlove 
- STATICFILES_FINDERS - Klase koje pronalaze statičke fajlove

**collectstatic command** 
- File collection - Skuplja fajlove iz svih aplikacija i STATICFILES_DIRS 
- Post-processing - Omogućava minifikaciju, kompresiju, hashing naziva fajlova 
- Storage backend - Podrška za local filesystem, S3, CDN, itd.

**Media files handling**
- MEDIA_ROOT - Direktorijum za user-uploaded fajlove 
- MEDIA_URL - URL prefix za pristup media fajlovima 
- FileField storage - Automatsko rukovanje upload-om i skladištenjem

### Caching podsistem

**Cache backends** 
- Memcached - Preporučen distribuirani cache
- Redis - Kroz django-redis paket 
- Database cache - Koristi bazu kao cache storage 
- File-based cache - Koristi fajl sistem 
- Local-memory cache - In-process cache za development 
- Dummy cache - No-op cache za testiranje

**Cache granularity** 
- Per-site cache - CacheMiddleware kešira ceo sajt
- Per-view cache - cache_page dekorator kešira view output
- Template fragment cache - tag kešira delove template-a
- Low-level API - Manuelno keširanje bilo kojih podataka
## Bezbednost i ranjivosti

### Bezbednost u dizajnu Django framework-a (Security by design)

#### SQL Injection zaštita

- Django ORM automatski escaping-uje upite
- Parametrizovani upiti sprečavaju injection napade
- Raw SQL i extra() metode zahtevaju dodatnu pažnju
- Nikada ne treba koristiti string interpolaciju za SQL upite
- Korišćenje ORM-a je najsigurniji pristup
#### Cross-Site Scripting (XSS) zaštita

- Template engine automatski escapuje sve varijable
- Auto-escaping je podrazumevano uključen
- Filter 'safe' i mark_safe() omogućavaju eksplicitno poveravanje HTML-a
- Opasnosti:
    - korišćenje safe filtera bez validacije
    - renderovanje JSON-a direktno u template-u
- Dobra praksa je validacija i sanitizacija korisničkog unosa pre označavanja kao safe

#### Cross-Site Request Forgery (CSRF) zaštita

- Middleware automatski proverava CSRF token
- csrf_token tag mora biti prisutan u svakoj POST formi
- AJAX zahtevi zahtevaju X-CSRFToken header
- Izuzeci:
    - csrf_exempt dekorator isključuje zaštitu
    - upotreba csrf_exempt povećava rizik
- CSRF zaštita radi na principu provere tokena koji se generiše za svaku sesiju

### Autentifikacija i autorizacija

#### Password handling

- PBKDF2 je podrazumevani hasher
- Podržani su Argon2, bcrypt i scrypt
- PASSWORD_HASHERS setting kontroliše algoritme
- Opasnosti:
    - korišćenje slabih hashera
    - hardkodovane lozinke u kodu

#### Session hijacking

- SessionMiddleware upravlja sesijama
- Podržano je skladištenje u bazi, cache-u, fajlovima
- Opasnosti:
    - session fixation napadi
    - nedostatak HTTPS-a
- Zaštita:
    - SESSION_COOKIE_SECURE postavka
    - SESSION_COOKIE_HTTPONLY sprečava JavaScript pristup
    - redovno generisanje novih session ključeva

#### Brute-force napadi

- Django ne pruža rate limiting po defaultu
- Potrebna je implementacija kroz:
    - django-axes paket
    - django-ratelimit
    - custom middleware
### Ranjivosti u Django aplikacijama

#### Insecure Direct Object References (IDOR)

- Nastaje kada aplikacija direktno koristi user input za pristup objektima
- Primer: /user/123/profile gde 123 dolazi od korisnika
- Rešenje:
    - provera dozvola pre pristupa objektu
    - korišćenje get_object_or_404 sa filterima
    - implementacija permission_required dekoratora

#### Mass Assignment ranjivost

- Nastaje kod ModelForm-i koje dozvoljavaju sve polja
- Napadač može postaviti polja koja ne bi trebalo
- Zaštita:
    - eksplicitno definisanje fields u Meta klasi
    - korišćenje exclude za opasna polja
    - custom validacija

#### Server-Side Request Forgery (SSRF)

- Nastaje kada aplikacija pravi HTTP zahteve bazirane na korisničkom unosu
- Django ne pruža automatsku zaštitu
- Opasnosti:
    - pristup internim servisima
    - skeniranje unutrašnje mreže
- Mitigacija:
    - whitelist dozvoljenih domena
    - izbegavanje direktnog prosljeđivanja URL-ova
    - korišćenje timeout-a

#### Insecure Deserialization

- pickle, yaml i XML parseri mogu biti opasni
- Deserijalizacija nepoverljivih podataka može dovesti do remote code execution
- Zaštita:
    - korišćenje JSON umesto pickle-a
    - validacija podataka pre deserijalizacije
    - ograničenje tipova koje se deserijalizuju

### Bezbednost konfiguracije

#### DEBUG mod u produkciji

- DEBUG = True otkriva:
    - stack trace-ove sa osetljivim podacima
    - strukturu baze podataka
    - instaliranih paketa
- Uvek koristiti DEBUG = False u produkciji
- Korišćenje environment varijabli za konfiguraciju

#### SECRET_KEY izloženost

- SECRET_KEY koristi se za kriptografske operacije
- Kompromitovan ključ omogućava:
    - falsifikovanje sesija
    - narušavanje CSRF zaštite
- Dobra praksa:
    - generisanje jakog ključa
    - čuvanje u environment varijablama
    - rotacija ključeva

#### ALLOWED_HOSTS

- Kontroliše koje host/domain imena može servirati Django
- Prazna lista u produkciji dozvoljava HTTP Host header napade
- Mora sadržati samo legitimne domene
- Dodatno koristiti SECURE_SSL_REDIRECT za HTTPS enforcing