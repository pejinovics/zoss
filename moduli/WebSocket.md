## Uvod

WebSocket je komunikacioni protokol koji omogućava full-duplex komunikaciju preko jedne TCP konekcije. Standardizovan je 2011. godine kroz RFC 6455 kao odgovor na potrebu za real-time komunikacijom između web klijenata i servera.

Klijent uvek inicira komunikaciju, WebSocket omogućava serveru da šalje podatke klijentu u bilo kom trenutku bez eksplicitnog zahteva. Ovo čini WebSocket idealnim za aplikacije sa stalnim ažuriranja poput chat-ova, live notifikacija, gaming-a i collaborative editing-a.

#### Prednosti WebSocket-a

- Persistentna konekcija: Jedna konekcija ostaje otvorena tokom celog životnog ciklusa sesije.
- Niska latencija: Nema overhead-a HTTP ciklusa za svaku poruku.
- Bidirekcioni protok: Server i klijent šalju poruke nezavisno.
- Efikasnost: Manji overhead u odnosu na polling tehnike.
- Real-time komunikacija: Instant isporuka poruka.

#### Kako WebSocket radi

WebSocket konekcija počinje kao HTTP zahtev sa Upgrade header-om. Server odgovara sa 101 Switching Protocols statusom, nakon čega se TCP konekcija prebacuje na WebSocket protokol. Poruke se razmenjuju kroz frames koji mogu biti tekstualni ili binarni.


## Arhitektura WebSocket-a sa Django integracijom

#### Jezgro sistema

Django Channels Channels omogućava rad sa protokolima koji zahtevaju long-running konekcije. Uvodi ASGI podršku koja zamenjuje WSGI za asinkrone protokole. Arhitektura je bazirana na događajima (event-driven) i zadržava kompatibilnost sa standardnim Django HTTP zahtevima.

Daphne Daphne je ASGI application server optimizovan za Channels. On upravlja životnim ciklusom konekcija, rutira protokole ka odgovarajućim handlerima i podržava hiljade konkurentnih konekcija.

#### Routing podsistem

ProtocolTypeRouter Ovo je top-level router koji rutira zahteve na osnovu protokola. Mapira http ili websocket na odgovarajuću ASGI aplikaciju.

URLRouter Mapira WebSocket putanje na Consumer klase koristeći path convertere slične onima u standardnom Django rutiranju.

AuthMiddlewareStack Povezuje WebSocket konekciju sa Django session i user objektom. Automatski vrši autentifikaciju korisnika na osnovu cookie-ja iz handshake zahteva.

#### Consumer podsistem

Consumer klase sadrže event handlere za događaje poput connect, receive i disconnect. Postoje tri glavna tipa:

- WebsocketConsumer: Sinhroni consumer koji blokira thread tokom obrade.
    
-  AsyncWebsocketConsumer: Koristi async/await za neblokirajuću obradu. Za pristup ORM-u koristi se database_sync_to_async.
    
- JsonWebsocketConsumer: Automatski vrši serijalizaciju i deserijalizaciju JSON poruka.
    

#### Channel layers podsistem

Channel layer omogućava komunikaciju između različitih Consumer instanci. Najčešće se koristi Redis backend.

Groups Konekcije se mogu grupisati radi broadcast-ovanja poruka svim članovima grupe. Grupe su dinamičke i kreiraju se u letu.

Channel names Svaka consumer instanca ima jedinstveni, automatski generisani channel_name koji služi za direktno rutiranje poruka toj specifičnoj konekciji.

#### Integracija sa Django ekosistemom

ORM i Baza podataka Pristup bazi iz asinkronih consumera vrši se preko wrapper-a poput sync_to_async kako bi se osigurala thread safety i izbeglo blokiranje event loop-a.

Authentication i Sessions Channels koristi iste session backende kao i standardni Django, omogućavajući pristup self.scope['user'] i self.scope['session'].

Signals Django signals rade normalno unutar consumera, što omogućava da post_save signal iz modela trigeruje WebSocket notifikaciju.

#### Deployment i Security

Deployment arhitektura U produkciji se koristi Daphne u kombinaciji sa Nginx-om kao reverse proxy-jem. Nginx mora biti konfigurisan da podržava WebSocket upgrade. Za skalabilnost se koristi Redis klaster.

Security considerations Obavezno je korišćenje wss protokola (TLS). Server treba da vrši Origin checking radi zaštite od CSRF napada. Preporučuje se implementacija rate limiting-a i validacija veličine poruka radi zaštite od DoS napada.

Performance optimizations Korisno je koristiti connection pooling za Redis i bazu podataka. Poruke se mogu komprimovati (Gzip), a neaktivne konekcije treba automatski zatvarati.

Monitoring i Testing Monitoring se vrši kroz praćenje broja aktivnih konekcija i protoka poruka u sekundi. Za testiranje se koristi ChannelsLiveServerTestCase i Communicator helper za simulaciju WebSocket saobraćaja bez potrebe za pravim browserom.