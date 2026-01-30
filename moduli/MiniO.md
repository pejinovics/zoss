# MiniO

## Uvod

MinIO je open-source rešenje za  skladištenje objekata, dizajnirano kao cloud-native sistem i u potpunosti kompatibilno sa Amazon S3 API-jem. Za razliku od tradicionalnih fajl ili blok sistema, MinIO skladišti podatke u obliku objekata, što omogućava jednostavno skaliranje, visoku dostupnost i efikasno rukovanje velikim količinama nestruktuiranih podataka. Sistem je implementiran kao lagan server koji se može pokretati u kontejnerskim okruženjima i lako integrisati u moderne mikroservisne arhitekture.

MinIO je posebno optimizovan za rad u distribuiranim okruženjima, gde više čvorova zajednički obezbeđuje skladišni prostor, toleranciju na otkaze i visok nivo paralelizma. Korišćenjem mehanizama poput *erasure coding-a,* replikacije i pametne distribucije podataka, MinIO postiže balans između pouzdanosti, performansi i efikasnog korišćenja resursa. Zbog toga se često koristi kao skladište za sisteme koji generišu velike količine podataka, poput analitičkih platformi ili sistema za logovanje.

U okviru našeg sistema MinIO ima ulogu trajnog skladišta snapshot-ova dokumenata. Podaci koje generiše Document servis se asinhrono prosleđuju putem RabbitMQ-a i zatim perzistiraju u MinIO skladištu koristeći S3 API. Ovakav pristup omogućava da se skladištenje podataka obavlja u pozadini, bez negativnog uticaja na performanse sistema, uz istovremeno obezbeđivanje pouzdanog i skalabilnog čuvanja dokumenata.

## Arhitektura

<img width="1092" height="684" alt="architect" src="https://github.com/user-attachments/assets/40bd35e7-9ce3-4987-8a91-6d2dd5bc5345" />

*Slika 1. Arhitektura MiniO-a.*

## Izvori:
- https://docs.daocloud.io/en/storage/hwameistor/application/minio/minio/
- https://xenonstack.medium.com/minio-distributed-object-storage-architecture-performance-xenonstack-fe9dfc125c17
- https://dev.to/ashokan/object-storage-as-primary-storage-the-minio-story-3g39
- https://e-whisper.com/posts/9462/
