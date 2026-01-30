# MiniO

## Uvod

MinIO je open-source rešenje za  skladištenje objekata, dizajnirano kao cloud-native sistem i u potpunosti kompatibilno sa Amazon S3 API-jem. Za razliku od tradicionalnih fajl ili blok sistema, MinIO skladišti podatke u obliku objekata, što omogućava jednostavno skaliranje, visoku dostupnost i efikasno rukovanje velikim količinama nestruktuiranih podataka. Sistem je implementiran kao lagan server koji se može pokretati u kontejnerskim okruženjima i lako integrisati u moderne mikroservisne arhitekture.

MinIO je posebno optimizovan za rad u distribuiranim okruženjima, gde više čvorova zajednički obezbeđuje skladišni prostor, toleranciju na otkaze i visok nivo paralelizma. Korišćenjem mehanizama poput *erasure coding-a,* replikacije i pametne distribucije podataka, MinIO postiže balans između pouzdanosti, performansi i efikasnog korišćenja resursa. Zbog toga se često koristi kao skladište za sisteme koji generišu velike količine podataka, poput analitičkih platformi ili sistema za logovanje.

U okviru našeg sistema MinIO ima ulogu trajnog skladišta snapshot-ova dokumenata. Podaci koje generiše Document servis se asinhrono prosleđuju putem RabbitMQ-a i zatim perzistiraju u MinIO skladištu koristeći S3 API. Ovakav pristup omogućava da se skladištenje podataka obavlja u pozadini, bez negativnog uticaja na performanse sistema, uz istovremeno obezbeđivanje pouzdanog i skalabilnog čuvanja dokumenata.

## Arhitektura

Arhitektura prikazana na *Slici 1* predstavlja distribuirani MinIO klaster u kome aplikacije komuniciraju sa skladištem preko S3 API-ja, dok se podaci ispod toga obrađuju kroz slojeve servera, objektne logike i fizičkog skladištenja. Svaki čvor (node) ima MinIO server koji implementira S3 kompatibilan interfejs i lokalno upravlja diskovima, a čvorovi međusobno komuniciraju internim protokolima radi koordinacije, raspodele podataka i oporavka u slučaju otkaza.

<img width="1092" height="684" alt="architect" src="https://github.com/user-attachments/assets/40bd35e7-9ce3-4987-8a91-6d2dd5bc5345" />

*Slika 1. Arhitektura MiniO-a.*

### Client / Application sloj (S3 API)

Aplikacije pristupaju MinIO-u preko S3 API-ja, što znači da se sistem ponaša kao standardni objektni storage (bucket-i, objekti, operacije PUT/GET/LIST…). Ovo pojednostavljuje integraciju, jer servisi ne zavise od internog načina skladištenja već samo od stabilnog API-ja. U praksi, isti interfejs olakšava kasniju migraciju ili povezivanje sa drugim S3-kompatibilnim alatima i bibliotekama. Time se MinIO prirodno uklapa u cloud-native okruženja i mikroservisnu arhitekturu.

### MinIO Server sloj

MinIO server je front-end skladišta - prima S3 zahteve, autentifikuje ih i sprovodi logiku upisa/čitanja objekata. Tipično se pokreće kao lagan servis i može raditi na VM-ovima ili Kubernetes-u. U distribuiranom režimu, više MinIO instanci zajedno čini jedan klaster, pa klijenti dobijaju jedinstveno skladište iako je fizički raspoređeno. Ovaj sloj je odgovoran i za koordinaciju sa drugim čvorovima tokom operacija koje zahtevaju distribuciju ili oporavak podataka.

### Object Layer

Object layer predstavlja logiku nad objektima - upravlja metapodacima i načinom na koji se jedan objekat raspakuje na više delova pri upisu u distribuiranom režimu. U ovom sloju se obično nalaze mehanizmi poput kompresije, keširanja i enkripcije (u zavisnosti od konfiguracije), jer se tu podaci obrađuju pre nego što odu na disk. Ključna stvar je da se pouzdanost često ostvaruje *erasure coding-om*, gde se objekat deli na data/parity delove raspoređene preko više diskova/čvorova. Tako MinIO može da nastavi rad i ako deo diskova ili čak čvorova otkaže, uz mogućnost *healing* oporavka.

### Storage Layer

Storage layer je najniži sloj koji radi sa fizičkim diskovima i fajl sistemom na svakom čvoru. MinIO je zamišljen da koristi lokalne diskove (umesto skupih shared-storage sistema) i da skalira dodavanjem novih diskova/čvorova. U distribuiranom režimu, podaci su raspodeljeni po diskovima tako da se postiže paralelizam čitanja/pisanja. Pouzdanost na ovom nivou se dodatno čuva kroz zaštitu integriteta podataka, kako bi se detektovala i sanirala korupcija na disku.

### Komunikacija između čvorova

U MinIO klasteru čvorovi sarađuju kako bi se održala konzistentnost, izvršila raspodela shard-ova i omogućio oporavak u slučaju otkaza. Interna komunikacija služi za koordinaciju operacija (npr. upis objekta preko više nodova), kao i za periodične *healing* procese kada neki disk/čvor privremeno nestane pa se kasnije vrati. Ovakav dizajn daje kombinaciju skalabilnosti (dodaješ node/disk) i otpornosti (klaster nastavlja rad i pri otkazima), što je i glavna vrednost distribuiranog objekt storage-a.

## Bezbednost

MinIO obezbeđuje niz bezbednosnih mehanizama koji su od značaja za zaštitu podataka. Među najvažnijim aspektima su autentifikacija i autorizacija zasnovane na IAM politikama, enkripcija podataka, kao i izolacija pristupa putem korisničkih naloga i pristupnih ključeva kompatibilnih sa S3 modelom. Posebnu pažnju treba posvetiti pravilnoj konfiguraciji mrežnog pristupa, upravljanju ključevima i zaštiti internih komunikacija između čvorova klastera.

<img width="947" height="399" alt="image" src="https://github.com/user-attachments/assets/863ba288-7905-44df-b557-a181388f4329" />

*Slika 2. Bezbednosni aspekti MiniO-a*

Ipak, više o bezbednosnim aspektima će biti razmatrano u sledećim koracima analize sistema i njegovih komponenti.

## Izvori:
- https://docs.daocloud.io/en/storage/hwameistor/application/minio/minio/
- https://xenonstack.medium.com/minio-distributed-object-storage-architecture-performance-xenonstack-fe9dfc125c17
- https://dev.to/ashokan/object-storage-as-primary-storage-the-minio-story-3g39
- https://e-whisper.com/posts/9462/
