#### Uvod u Redis

Redis (Remote Dictionary Server) je open-source sistem za skladištenje podataka u memoriji koji se koristi kao baza podataka, keš i message broker. Za razliku od tradicionalnih baza koje podatke čuvaju na disku, Redis drži sve podatke u RAM memoriji, što mu omogućava performanse sa latencijom ispod jedne milisekunde.

Osnovne karakteristike
- In-memory data structures: Podržava kompleksne tipove podataka kao što su stringovi, liste, setovi, hash mape i bitmape.
- Single-threaded core: Glavni proces obrade komandi je jednonitan, što eliminiše potrebu za lock-ovima i obezbeđuje atomičnost operacija.
- Perzistencija: Iako radi u memoriji, može periodično snimati podatke na disk (RDB) ili beležiti svaku operaciju u log (AOF).    
- Visoka dostupnost: Podržava replikaciju i automatski failover kroz Sentinel ili Cluster režime.

#### Redis unutrašnja arhitektura

Osnova Redisa počiva na jednostavnosti i efikasnosti. Iako se često opisuje kao ključ-vrednost skladište, njegova arhitektura je slojevita i optimizovana za maksimalnu propusnu moć.

##### Event Loop i Single-Threaded model

Glavna karakteristika Redisa je korišćenje jedne niti (single-thread) za obradu svih klijentskih komandi. Razlozi za ovakav dizajn su izbegavanje lock-ova, sprečavanje race condition-a i eliminacija procesorskog overhead-a koji nastaje pri promeni konteksta između niti (context switching).

Za rukovanje hiljadama istovremenih konekcija koristi se I/O Multiplexing. Redis koristi event loop mehanizme specifične za operativni sistem (epoll na Linuxu, kqueue na BSD-u). Ovi mehanizmi omogućavaju jednoj niti da nadgleda ogroman broj soketa i reaguje samo na one koji su spremni za čitanje ili pisanje, čime se postiže izuzetna skalabilnost.

##### Memorijska organizacija i strukture podataka

Redis ne čuva podatke kao obične stringove na disku, već ih organizuje kroz Redis Object System. Svaki objekat u Redisu ima tip (npr. string, list, hash) i unutrašnju enkoding šemu.

Interni enkoding (Internal Encoding) Redis dinamički menja način na koji čuva podatke u memoriji kako bi uštedeo prostor. Na primer, mali hash-evi se čuvaju kao ziplist (kompaktna lista), a kada pređu određenu veličinu, automatski se konvertuju u pravu hash tabelu (hashtable).

Hashtable implementacija Redis koristi dve hash tabele istovremeno za svaki skup podataka kako bi omogućio incremental resizing. Kada tabela postane preopterećena, Redis postepeno prebacuje elemente u novu, veću tabelu u pozadini, bez blokiranja glavnog procesa.

##### Mehanizam perzistencije

Redis nudi balans između brzine i sigurnosti podataka kroz dva arhitektonska rešenja:

1. RDB (Snapshotting): Proces se račva (fork) i kreira dete-proces koji kopira trenutno stanje memorije u binarni fajl. Ovaj proces koristi copy-on-write mehanizam operativnog sistema, što znači da ne zauzima duplo više memorije osim u delovima gde se podaci menjaju tokom pisanja.
    
2. AOF (Append Only File): Svaka komanda koja menja podatke beleži se u log fajl. Kako log raste, Redis vrši AOF Rewrite – kreira novi, optimizovani fajl koji sadrži minimalan broj komandi potrebnih za rekonstrukciju trenutnog stanja.
    

###### Replikacija i sinhronizacija

Arhitektura replikacije se zasniva na asinhronom slanju komandi od mastera ka replikama.

Handshake i Initial Sync: Kada se replika poveže, master šalje ceo RDB fajl. Replication ID i Offset: Svaki master ima jedinstveni ID i offset (brojač bajtova). Ako se veza prekine, replika pokušava delimičnu sinhronizaciju (partial resync) šaljući svoj offset, čime izbegava ponovno slanje celog seta podataka.

##### Redis Cluster i Sharding

U klaster arhitekturi podaci se ne nalaze na jednom serveru, već su horizontalno podeljeni.

Hash Slots: Redis koristi 16384 hash slota. Svaki ključ se mapira na slot pomoću CRC16 algoritma. Slotovi su raspoređeni među čvorovima u klasteru. Gossip Protocol: Čvorovi u klasteru neprestano komuniciraju koristeći gossip protokol kako bi znali stanje ostalih čvorova, identifikovali kvarove i ažurirali mapu slotova. Client Redirection: Ako klijent pošalje upit čvoru koji ne drži traženi slot, čvor mu šalje MOVED odgovor sa adresom ispravnog čvora.

##### Upravljanje memorijom i Eviction polise

Pošto je memorija ograničen resurs, Redis arhitektura uključuje algoritme za oslobađanje prostora. Kada se dostigne maxmemory limit, aktiviraju se eviction polise:

- LRU (Least Recently Used): Uklanja ključeve koji nisu korišćeni najduže vreme.
    
- LFU (Least Frequently Used): Uklanja ključeve koji se najređe koriste.
    
- TTL: Uklanja ključeve kojima najskorije ističe vreme trajanja.
    

Ovi algoritmi su aproksimativni – Redis nasumično bira mali uzorak ključeva i na njima primenjuje logiku kako bi uštedeo CPU resurse.

#### Režimi rada i Skalabilnost

Standalone Najjednostavniji oblik gde jedan Redis proces služi sve zahteve. Pogodan za razvoj i manje aplikacije, ali predstavlja single point of failure.

Replikacija (Master-Slave) Jedan Master čvor prima zapise, dok se podaci asinhrono kopiraju na više Slave čvorova. Slave čvorovi služe za čitanje, čime se postiže horizontalno skaliranje čitanja.

Redis Sentinel Sistem za monitoring i visoku dostupnost. Sentinel procesi prate rad Master čvora i, u slučaju kvara, automatski promovišu jedan od Slave čvorova u novi Master, obaveštavajući klijente o promeni adrese.

Redis Cluster Distribuirana implementacija koja automatski deli podatke na više čvorova (sharding). Cluster koristi koncept od 16384 hash slotova koji se raspoređuju po čvorovima, omogućavajući linearno skaliranje i kapaciteta i performansi.

#### Redis kao Message Broker

Pub/Sub (Publish/Subscribe) Mehanizam gde izdavači šalju poruke na kanale, a pretplatnici ih primaju u realnom vremenu. Poruke su vatrene i zaboravi tipa (fire and forget), što znači da se ne čuvaju ako klijent nije povezan u tom trenutku.

Redis Streams Moderniji pristup messaging-u koji omogućava trajno čuvanje poruka, slično kao Apache Kafka. Podržava grupe potrošača (consumer groups) i praćenje potvrda o prijemu poruka.

#### Uloga Redisa u Django Channels arhitekturi

U kontekstu Django Channels-a, Redis služi kao Channel Layer (sloj za komunikaciju).

- Message Routing: Kada jedan korisnik pošalje poruku u grupu, Django je šalje u Redis.
- Cross-process communication: Pošto Django procesi (Daphne workeri) ne dele memoriju, Redis služi kao zajednički prostor gde svi procesi mogu da vide ko je u kojoj grupi.
- Backpressure handling: Redis privremeno skladišti poruke koje čekaju na obradu ako je sistem preopterećen.
    

#### Perzistencija podataka

Redis nudi dva glavna mehanizma za očuvanje podataka nakon restarta:

1. RDB (Redis Database): Kreira snapshot podataka u određenim intervalima. Veoma je kompaktan i brz za oporavak, ali može dovesti do gubitka podataka između dva snapshot-a.
2. AOF (Append Only File): Beleži svaku komandu pisanja u fajl. Bezbedniji je jer osigurava minimalan gubitak podataka, ali fajlovi postaju veliki i sporije se učitavaju.

#### Security i Optimizacija

Bezbednost Redis po defaultu nema omogućenu autentifikaciju i očekuje se da bude u privatnoj mreži. Sigurnost se postiže preko AUTH komande (lozinka), ACL (Access Control Lists) za finu kontrolu dozvola i TLS enkripcije za saobraćaj.

Optimizacija performansi

- Maxmemory policy: Definiše šta se dešava kada se RAM napuni (npr. LRU - izbacivanje najstarijih podataka).
- Pipelining: Omogućava klijentu da pošalje više komandi odjednom bez čekanja na pojedinačne odgovore.
- LUA skripte: Omogućavaju izvršavanje kompleksne logike direktno na serveru, čime se smanjuje mrežni saobraćaj.