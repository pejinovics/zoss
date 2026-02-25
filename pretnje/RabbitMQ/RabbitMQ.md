# RabbitMQ

## Uvod

RabbitMQ u okviru sistema za kolaborativno uređivanje dokumenata ima ulogu posrednika u asinhronom toku podataka između Document servisa (*Go*) i skladišta MinIO. Snapshot verzije dokumenata se ne upisuju direktno u skladište, već se prvo šalju na RabbitMQ red (*snapshots*), odakle ih consumer servis preuzima i čuva u MinIO. Ovakav obrazac omogućava neometan rad aplikacije i bolju skalabilnost, ali uvodi dodatnu napadnu površinu u vidu message broker-a.

U ovom izveštaju analizirana je pretnja narušavanja dostupnosti (*Availability*) RabbitMQ komponente putem napada tipa Resource Exhaustion (*CWE-400*). Napad se zasniva na masovnom slanju velikih, persistent poruka u red bez odgovarajućih ograničenja, čime dolazi do iscrpljivanja memorije i diska brokera. Kao posledica, RabbitMQ aktivira zaštitne mehanizme, zatvara konekcije i prestaje da prihvata nove poruke, što dovodi do prekida snapshot pipeline-a i degradacije funkcionalnosti sistema.

## Stablo napada

<img width="6698" height="4470" alt="RabbitMQ Resource-2026-02-25-005025" src="https://github.com/user-attachments/assets/31861470-86f4-4b54-86fa-c89d85728ac1" />

## Realizovan napad – Resource Exhaustion (CWE-400)

Napad se zasniva na preopterećenju RabbitMQ brokera slanjem velikog broja velikih, persistent poruka u red `snapshots`. Producer generiše poruke unapred definisane veličine (npr. 1 MB) i šalje ih u velikom broju (npr. 10.000). Pošto consumer paralelno upisuje poruke u MinIO, brzina obrade je manja od brzine slanja, što dovodi do nagomilavanja poruka u redu.

U ranjivoj konfiguraciji red je deklarisan bez ikakvih ograničenja:

```go
q, err := ch.QueueDeclare(  
    "snapshots",  
    true,  // durable  
    false, // autoDelete  
    false, // exclusive  
    false, // noWait  
    nil,   // bez limita  
)
```

Producer šalje persistent poruke:
```go
err = ch.Publish(  
    "",  
    "snapshots",  
    false,  
    false,  
    amqp.Publishing{  
        DeliveryMode: amqp.Persistent,  
        ContentType:  "application/octet-stream",  
        Body:         body,  
    })
```

Kako broj poruka raste, dubina reda se povećava, memorijska potrošnja brokera naglo raste, a nakon dostizanja internog praga zaštite (memory high watermark), RabbitMQ zatvara konekcije. Producer tada dobija grešku:

```
Exception (504) Reason: "channel/connection is not open"
```

Ovim trenutkom snapshot pipeline se prekida – nove verzije dokumenata se više ne mogu sačuvati.

### Efekat na sistem

Napad ima sledeće posledice:

- RabbitMQ odbija nove poruke (narušena dostupnost).
- Producer servis ulazi u grešku zbog zatvorene konekcije.
- Snapshotovi koji nisu uspešno publishovani nikada ne stižu do MinIO.
- Sistem gubi funkcionalnu mogućnost verzionisanja dokumenata.
    
Važno je napomenuti da integritet postojećih podataka nije direktno kompromitovan, ali dolazi do funkcionalnog gubitka novih verzija, što predstavlja ozbiljan operativni problem.

## Mitigacije

Nakon analize primenjene su tehničke mere ograničavanja rasta reda i uvođenja kontrolisanog backpressure mehanizma.

### Ograničenje veličine reda

Red je deklarisan sa limitom ukupne veličine poruka:

```go
args := amqp.Table{  
    "x-max-length-bytes": 500 * 1024 * 1024, // 500 MB  
    "x-message-ttl":      300000,            // 5 minuta  
}  
```

```go
q, err := ch.QueueDeclare(  
    "snapshots",  
    true,  
    false,  
    false,  
    false,  
    args,  
)
```

Ovim se sprečava neograničeno povećanje reda. Poruke koje prelaze limit automatski se odbacuju.

### Prefetch i backpressure

Consumer je konfigurisan sa ograničenjem broja poruka koje istovremeno preuzima:

```go
err = ch.Qos(1, 0, false)
```

Na ovaj način broker ne isporučuje novu poruku dok prethodna nije potvrđena (_ACK_), čime se sprečava akumulacija velikog broja neobrađenih poruka.

### Rezultat nakon mitigacija

Nakon primene navedenih mera:
- Dubina reda ostaje unutar definisanog limita.
- Memorijska potrošnja brokera ostaje stabilna.
- Producer ne gubi konekciju.
- Sistem ostaje dostupan i pod opterećenjem.

Iako se deo poruka odbacuje kada se dostigne limit, sistem zadržava operativnu stabilnost i sposobnost obrade novih zahteva, što predstavlja prihvatljiv kompromis između pouzdanosti i dostupnosti.

## Ostali napadi

### Unacked Messages / Consumer Stalling

**Napadana površina i resursi:**
- Memorija RabbitMQ brokera
- Interni mehanizam pouzdane isporuke (ACK/NACK)
- Red `snapshots` i njegove _in-flight_ poruke

**Površina napada i resursi:**

Napadač ili kompromitovan consumer servis prestaje da šalje _ACK_ potvrde za preuzete poruke. Poruke ostaju u stanju _unacked_, broker ih zadržava u memoriji i ne redistribuira drugim potrošačima. Ukoliko se takvo stanje produži, broj neobrađenih poruka u memoriji kontinuirano raste.

**Posledice po sistem:**
- Postepeno povećanje memorijske potrošnje brokera
- Povećanje latencije u isporuci novih poruka
- Potencijalno aktiviranje memory watermark mehanizma
- Zastoj snapshot pipeline-a (poruke su “u obradi”, ali se ne izvršavaju)

Ovaj napad je posebno opasan jer ne zahteva masovno slanje poruka – dovoljan je logički zastoj potrošača.

**Mitigacije:**
- Ograničenje `prefetch` vrednosti (QoS)
- Monitoring dužine _unacked_ poruka
- Timeout i automatski restart consumer servisa
- Implementacija dead-letter mehanizma za neuspešne poruke

### Connection / Channel Exhaustion

**Površina napada i resursi:**

- TCP konekcije ka RabbitMQ
- Logički AMQP kanali
- Sistemski resursi (file descriptor-i, memorija)

**Kako izgleda napad:**

Napadač ili kompromitovan servis otvara veliki broj AMQP konekcija ili kanala bez njihovog zatvaranja. Svaka konekcija troši memoriju i sistemske resurse. Kada se dostigne maksimalan broj dozvoljenih konekcija, broker počinje da odbija nove zahteve.

**Posledice po sistem:**
- Producer i consumer servisi gube mogućnost povezivanja
- Snapshot pipeline prestaje sa radom
- Potencijalna degradacija drugih redova i servisa u istom brokeru

Za razliku od flood napada, ovde nije potrebno slanje poruka – dovoljno je iscrpeti komunikacione resurse.

**Mitigacije:**
- Ograničenje maksimalnog broja konekcija po korisniku
- Postavljanje limita nad brojem kanala
- Mrežna segmentacija i firewall pravila
- Korišćenje posebnih kredencijala po servisu

### Mass Durable Queue Creation

**Napadana površina i resursi:**
- Mehanizam deklaracije redova
- Memorija i metadata strukture brokera
- Disk (kod durable redova)

**Kako izgleda napad:**  

Napadač sa dozvolom deklaracije redova masovno kreira durable queue-ove. Svaki red zahteva memoriju za upravljanje i perzistentne operacije nad diskom. Veliki broj redova povećava kompleksnost upravljanja brokerom i opterećenje sistema.

**Posledice po sistem:**
- Usporavanje rada brokera
- Povećano vreme inicijalizacije i upravljanja redovima
- Indirektno usporenje ili nestabilnost reda `snapshots`
- Potencijalni pad performansi celog sistema

Ovaj napad ne zahteva flood poruka, već zloupotrebu administrativnih prava.

**Mitigacije:**
- Restriktivne permission politike (zabrana `queue.declare`)
- Izolacija putem virtual host-ova
- Policy pravila za ograničenje resursa po korisniku
- Redovan audit privilegija

## Zaključak

RabbitMQ predstavlja centralnu tačku asinhrone komunikacije i ključnu komponentu dostupnosti sistema. Analiza pokazuje da neadekvatna kontrola resursa, konekcija i potrošača može dovesti do ozbiljne degradacije ili potpunog prekida rada snapshot pipeline-a.

Implementirani napad demonstrira praktičnu realizaciju pretnje tipa _Resource Exhaustion (CWE-400)_, dok teorijski scenariji potvrđuju da dostupnost message brokera zavisi od pravilne konfiguracije limita, nadzora i restriktivnih politika pristupa. Otpornost sistema u event-driven arhitekturama ne zavisi isključivo od poslovne logike, već i od pažljivo definisanih infrastrukturnih kontrola.
