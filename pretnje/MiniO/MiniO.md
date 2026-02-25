# MiniO

## Uvod

MinIO predstavlja trajni skladišni sloj sistema i zadužen je za čuvanje snapshot verzija dokumenata. Svaki snapshot predstavlja istorijsku verziju dokumenta u određenom trenutku i mora biti zaštićen od neovlašćenih izmena kako bi se očuvala konzistentnost i revizibilnost podataka.

Za razliku od RabbitMQ komponente, gde je dominantna pretnja usmerena na dostupnost, kod MinIO-a ključni bezbednosni aspekt je integritet podataka (*Integrity*). Analizirana je ranjivost koja omogućava neprimetno prepisivanje postojećih objekata u bucket-u bez mogućnosti povratka, u slučaju da versioning nije uključen.

Napad spada u kategorije:
- *CWE-284 – Improper Access Control*
- *CWE-732 – Incorrect Permission Assignment for Critical Resource*

## Stablo napada

<img width="5670" height="4310" alt="RabbitMQ Resource-2026-02-25-020557" src="https://github.com/user-attachments/assets/ca1f7466-423e-4610-a3f8-161c150cdfe9" />

## Realizovan napad – Overwrite objekata bez versioninga

Napad se zasniva na činjenici da MinIO, ukoliko versioning nije uključen, dozvoljava da se objekat sa istim ključem prepiše novim sadržajem bez zadržavanja prethodne verzije.

U ranjivoj konfiguraciji bucket `snapshots` je kreiran bez uključenog versioninga:

```go
exists, _ := client.BucketExists(ctx, bucketName)  
if !exists {  
    client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})  
}
```

Napadač (ili kompromitovan servis) šalje novi `PUT` zahtev sa istim `objectName`:

```go
client.PutObject(ctx, bucketName, objectName,  
    bytes.NewReader([]byte(maliciousContent)),  
    int64(len(maliciousContent)),  
    minio.PutObjectOptions{ContentType: "application/json"})
```

Pošto versioning nije aktivan, novi sadržaj automatski zamenjuje prethodni objekat.

### Efekat na sistem

- Originalni snapshot je nepovratno izgubljen.
- Istorija verzija dokumenta više nije pouzdana.
- Ne postoji mehanizam detekcije izmene bez dodatnog audit sistema.
- Korisnici mogu raditi nad kompromitovanim podacima bez saznanja.

Za razliku od DoS napada, ovde sistem ostaje funkcionalan, ali podaci više nisu verodostojni – što predstavlja direktno narušavanje integriteta.

Izgled ranjive verzije:

<img width="1299" height="201" alt="image" src="https://github.com/user-attachments/assets/d1db6218-631f-4fce-9579-078153f030ab" />

## Mitigacije

### Uključivanje versioninga

Aktiviranjem versioninga, svaka izmena objekta generiše novu verziju:

```go
err = client.EnableVersioning(ctx, bucketName)
```

U tom slučaju, prepisivanje objekta ne briše prethodnu verziju već kreira novu, čime se omogućava povratak na originalni snapshot.

### Minimalne privilegije (IAM politike)

Odvajanje korisnika po nameni:

- `snapshot-writer` → samo `PutObject`
- `snapshot-reader` → samo `GetObject`
- zabrana `DeleteObject` operacija

Ovim se smanjuje površina napada i ograničava mogućnost zloupotrebe.

Izgled mitigovane verzije:

<img width="1266" height="200" alt="image" src="https://github.com/user-attachments/assets/ea2d6a73-5016-4e84-b425-fa2e2761d5e8" />

### Object Lock (WORM)

Object Lock omogućava zabranu brisanja ili izmene objekata u definisanom vremenskom periodu. Ova mera je naročito značajna u sistemima gde je potrebna regulatorna usklađenost ili forenzička pouzdanost podataka.
## Ostali napadi

### Neovlašćeno brisanje objekata

**Površina napada i resursi:**

- `DeleteObject` operacija
- Bucket `snapshots`
- IAM politika

**Kako izgleda napad:**  
Napadač sa preširokim privilegijama briše postojeće snapshotove. Bez versioninga i Object Lock-a, objekti se trajno uklanjaju.

**Posledice:**

- Gubitak istorijskih verzija
- Nemogućnost rekonstrukcije podataka
- Direktan uticaj na integritet i pouzdanost sistema
    

**Mitigacije:**

- Onemogućavanje `DeleteObject` privilegije
- Uključivanje versioninga
- Primena Object Lock mehanizma

### Curenje pristupnih ključeva (Credential Exposure)

**Površina napada i resursi:**

- Access key / secret key
- API endpoint MinIO-a
- Bucket `snapshots`

**Kako izgleda napad:**  
Ukoliko se kredencijali nalaze u kodu, logovima ili `.env` fajlovima bez zaštite, napadač može dobiti potpuni pristup bucket-u i izvršiti izmene ili brisanja objekata.

**Posledice:**

- Potpuna kompromitacija integriteta podataka
- Masovno prepisivanje ili brisanje snapshotova
- Gubitak poverenja u sistem

**Mitigacije:**

- Korišćenje tajni (Docker secrets / Vault)
- Rotacija ključeva
- Ograničenje privilegija po korisniku
- Audit pristupa

## Zaključak

Za razliku od RabbitMQ komponente gde je primarna pretnja usmerena na dostupnost, MinIO predstavlja kritičnu tačku integriteta podataka. Demonstrirani napad pokazuje da odsustvo versioninga omogućava neprimetno prepisivanje snapshotova i gubitak istorije verzija.

Primena versioninga, minimalnih privilegija i Object Lock mehanizma značajno povećava otpornost sistema na manipulaciju podacima i osigurava pouzdanost istorijskih zapisa.
