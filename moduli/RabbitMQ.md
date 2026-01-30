# RabbitMQ

## Uvod

U savremenom razvoju softvera mikroservisna arhitektura je postala standard, ali sa sobom donosi i izazov upravljanja velikim brojem nezavisnih servisa i njihovih međusobnih veza, što sistem lako može učiniti složenim i teškim za održavanje. U takvom okruženju oslanjanje isključivo na komunikaciju tipa zahtev–odgovor (*request/response*) često nije praktično, jer servisi mogu da blokiraju jedni druge, što se negativno odražava na performanse sistema. Zbog toga se kao efikasno rešenje uvodi **Message Broker**, koji omogućava asinhronu komunikaciju i razdvajanje aplikacija. Jedno od najčešće korišćenih rešenja u tu svrhu je **RabbitMQ**.

RabbitMQ predstavlja jedno od najpopularnijih open-source rešenja za sisteme zasnovane na porukama. On se oslanja na **AMQP 0–9–1** (*Advanced Message Queuing Protocol*), protokol aplikativnog sloja koji podatke prenosi u binarnom formatu putem okvira (frames) čiji format je prikazan na *Slici 1*, čime se postiže visoka efikasnost prenosa. Razvijen je u programskom jeziku Erlang. RabbitMQ takođe nudi napredne mehanizme poput potvrde o prijemu poruka i perzistencije, koji garantuju da podaci neće biti izgubljeni čak ni u slučaju kvara na mreži ili padova servera.

<img width="841" height="198" alt="image" src="https://github.com/user-attachments/assets/5df857a2-799f-4869-b861-472c440554de" />

*Slika 1. Format AMQP okvira (frame).*

U okviru našeg sistema za kolaborativno uređivanje dokumenata, RabbitMQ modul ima ulogu u protoku podataka: **Go (Document servis) -> RabbitMQ  -> MinIO (S3)**. Umesto da Document servis direktno vrši upis snapshot-ova na MinIO skladište, što može biti spora operacija, on te podatke delegira RabbitMQ-u i  oslobađa se za dalju interakciju. Ovaj pristup omogućava da se svaki snapshot bezbedno i pouzdano sačuva u pozadini bez ikakvog uticaja na performanse samog editora.

## Arhitektura

Arhitektura prikazana na *Slici 2* ilustruje RabbitMQ klaster sa podrškom za replikaciju redova, u kojem više brokera zajedno obezbeđuje pouzdanu i distribuiranu obradu poruka. Producer-i šalju poruke ka exchange komponentama unutar brokera, koje ih dalje prosleđuju u odgovarajuće redove (queue). Redovi su sinhronizovani između brokera unutar klastera kako bi se obezbedila dostupnost podataka i tolerancija na otkaz. Consumer-i preuzimaju poruke iz redova i obrađuju ih nezavisno, čime se postiže asinhrona i skalabilna komunikacija između sistema.

<img width="1019" height="712" alt="Screenshot 2026-01-30 011050" src="https://github.com/user-attachments/assets/af06bd7c-f5d7-4cdc-b3f9-ee44c12113ea" />

*Slika 2. Arhitektura RabbitMQ-a.*

### Producer

Producer je aplikacija ili servis koji kreira i šalje poruke u RabbitMQ sistem. Poruke se ne šalju direktno u red, već se prosleđuju exchange komponenti zajedno sa routing informacijama. Na ovaj način producer ostaje potpuno nezavisan od načina na koji se poruke kasnije distribuiraju i obrađuju. Njegova odgovornost se svodi isključivo na slanje poruka u sistem.

### Broker

Broker predstavlja jedan RabbitMQ server (node) koji upravlja prijemom, rutiranjem i skladištenjem poruka. U klaster okruženju više brokera radi zajedno, razmenjujući podatke i stanje sistema. Kao što je prikazano na slici, redovi se mogu replicirati između brokera kako bi poruke ostale dostupne i u slučaju otkaza pojedinačnog čvora. Ovakav pristup povećava pouzdanost i dostupnost sistema.

### Exchange

Exchange je komponenta zadužena za rutiranje poruka unutar brokera. Ona prima poruke od producer-a i, na osnovu definisanih pravila rutiranja, prosleđuje ih jednom ili više redova. Exchange ne skladišti poruke, već služi isključivo kao mehanizam za njihovu distribuciju. Time se omogućava fleksibilno usmeravanje poruka bez direktne veze između producer-a i redova.

### Queue

Queue predstavlja strukturu za privremeno skladištenje poruka koje čekaju na obradu. Poruke ostaju u redu sve dok ih neki consumer ne preuzme i potvrdi obradu. Redovi omogućavaju da se sistem nosi sa neujednačenim opterećenjem, jer poruke mogu biti akumulirane dok potrošači ne postanu dostupni. U klasteru, redovi mogu biti sinhronizovani između više brokera radi veće pouzdanosti.

### Binding

Binding definiše vezu između exchange-a i queue-a. On određuje pod kojim uslovima će poruke primljene na exchange biti prosleđene određenom redu. Putem binding pravila precizno se kontroliše tok poruka kroz sistem. Ovaj mehanizam omogućava fino podešavanje rutiranja bez izmene producer ili consumer logike.

### Consumer

Consumer je aplikacija ili servis koji preuzima poruke iz reda i vrši njihovu obradu. Nakon uspešne obrade, consumer šalje potvrdu brokeru, čime se poruka trajno uklanja iz reda. Više consumer-a može paralelno konzumirati poruke iz istog reda, što omogućava horizontalno skaliranje. Na ovaj način se povećava propusnost sistema i ubrzava obrada podataka.

### Connection i Channel

Connection predstavlja fizičku mrežnu vezu između aplikacije i RabbitMQ brokera, najčešće ostvarenu putem TCP protokola. Channel-i su logički kanali unutar jedne konekcije koji omogućavaju istovremenu komunikaciju bez otvaranja dodatnih mrežnih veza. Ovakav model značajno smanjuje potrošnju resursa i povećava efikasnost sistema. Većina operacija u RabbitMQ-u se zapravo odvija na nivou kanala, a ne konekcije.

### Virtual Host (vhost)

Virtual host predstavlja logičku izolaciju unutar jednog RabbitMQ brokera. On omogućava da više aplikacija ili sistema koristi isti broker bez međusobnog uticaja i mešanja resursa. Svaki vhost poseduje sopstvene exchange-e, queue-e i binding-e, kao i posebna pravila pristupa. Ovaj mehanizam je posebno koristan u multi-tenant okruženjima, iako nije obavezan deo osnovne arhitekture.

### Tipovi Exchange-a

RabbitMQ podržava više tipova exchange-a, uključujući direct, topic, fanout i headers. Svaki tip definiše način na koji se poruke rutiraju ka redovima. Izbor odgovarajućeg tipa exchange-a omogućava prilagođavanje sistema različitim obrascima komunikacije. Ovi tipovi predstavljaju detalj implementacije rutiranja, a ne posebne arhitekturne komponente.

### Acknowledgments (ACK / NACK)

Acknowledgment mehanizam omogućava potrošačima da potvrde uspešnu ili neuspešnu obradu poruke. Korišćenjem ACK i NACK signala broker dobija informaciju da li poruku treba ukloniti iz reda ili ponovo poslati. Na ovaj način se sprečava gubitak poruka u slučaju grešaka ili prekida u radu consumer-a. Ovaj mehanizam predstavlja logički aspekt pouzdanosti sistema, a ne fizičku komponentu arhitekture.

### Persistence i Durability

Persistence i durability mehanizmi obezbeđuju trajno čuvanje poruka i redova u RabbitMQ sistemu. Durable redovi i persistent poruke omogućavaju oporavak sistema nakon restarta ili pada brokera. U zavisnosti od konfiguracije, podaci se mogu čuvati u memoriji ili na disku, što direktno utiče na performanse i pouzdanost. Ovi mehanizmi predstavljaju operativni sloj arhitekture i obično se opisuju tekstualno.

## Bezbednost

RabbitMQ nudi niz bezbednosnih mehanizama koji su ključni za zaštitu podataka i integritet sistema. Među najvažnijim aspektima su autentifikacija i autorizacija, koje određuju ko može da šalje i prima poruke u okviru određenih virtualnih host-ova, kao i enkripcija komunikacije putem TLS/SSL protokola kako bi se sprečilo prisluškivanje i napadi u mreži. Pored toga, RabbitMQ omogućava finu kontrolu nad pravima korisnika, audite događaja i politike pristupa na nivou virtualnih host-ova. Takođe, pažnja mora da se posveti upravljanju lozinkama, konfiguraciji firewall-a i ograničavanju pristupa brokeru samo na poverene servise i mreže, kako bi se umanjio rizik od neautorizovanih pristupa.

Ipak, više o bezbednosnim aspektima će biti razmatrano u sledećim koracima analize sistema i njegovih komponenti.

<img width="679" height="300" alt="image" src="https://github.com/user-attachments/assets/1feb69e1-3ed6-410e-afd3-adbfda099b89" />

*Slika 3. Security komponente RabbitMQ-a.*

## Izvori:
- https://medium.com/cwan-engineering/rabbitmq-concepts-and-best-practices-aa3c699d6f08
- https://www.researchgate.net/publication/347866161_A_Fair_Comparison_of_Message_Queuing_Systems
- https://medium.com/@okanyildiz1994/rabbitmq-security-complete-guide-for-enterprise-message-broker-systems-cea5318be941
- https://medium.com/p/bbfa0a6b0eff
- https://seventhstate.io/rabbitmq-architecture/
- https://www.arcjournals.org/pdfs/ijrscse/v6-i2/4.pdf
- https://www.brianstorti.com/speaking-rabbit-amqps-frame-structure/#:~:text=20%20Apr%202016,have%20the%20same%20basic%20structure:
