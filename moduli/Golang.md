# Golang

## Uvod

Go (Golang) je programski jezik osmišljen da odgovori na izazove razvoja modernih distribuiranih sistema. Njegove osnovne karakteristike uključuju jednostavnu sintaksu, efikasne mehanizme za upravljanje memorijom i ugrađenu podršku za konkurentnost. Umesto pristupa baziranog na nitima operativnog sistema i primitivama za zaključavanje, Go nudi lakše konkurentne konstrukcije koje ubrzavaju pisanje paralelnog koda sa manjom sklonošću greškama.

Značajnu ulogu u ekosistemu Go-a ima njegov runtime, koji obezbeđuje automatsko upravljanje kritičnim aspektima izvršavanja programa. Runtime preuzima zadatke raspoređivanja gorutina, alokacije i oslobađanja memorije, kao i  komunikacije sa operativnim sistemom putem sistemskih poziva. Ova abstrakcija kompleksnosti omogućava programerima da se fokusiraju na poslovnu logiku aplikacije, dok runtime sistem efikasno upravlja resursima u pozadini. 

U okviru našeg sistema, Go se koristi kao primarni jezik za implementaciju serverskih servisa, uključujući Document servis koji se bavi obradom dokumenata ikomunikacijom sa RabbitMQ-om. Zahvaljujući efikasnom modelu konkurentnosti i malom runtime overhead-u, Go omogućava da se veliki broj istovremenih zahteva obrađuje stabilno i sa predvidivim performansama.

## Arhitektura

Arhitektura prikazana na *Slici 1* ilustruje način na koji se Go program izvršava kroz saradnju korisničkog koda, Go runtime-a i operativnog sistema. Na vrhu se nalazi Go program koji koristi goroutine, channel-e i sinhronizacione primitive, dok runtime ispod toga upravlja raspoređivanjem izvršavanja, memorijom, mrežnim operacijama i sistemskim pozivima. Go runtime predstavlja posredni sloj između aplikacije i operativnog sistema, čime se postiže balans između apstrakcije i performansi. 

<img width="623" height="593" alt="image" src="https://github.com/user-attachments/assets/2a151969-cbb8-434a-9e69-520d4c67d06a" />

*Slika 1. Arhitektura Go-a.*

### Go Program

Go program predstavlja korisnički napisanu logiku aplikacije i sastoji se od funkcija, paketa i konkurentnih konstrukata. Programer eksplicitno koristi goroutine, channel-e i sinhronizacione mehanizme, dok se detalji izvršavanja delegiraju runtime-u. Ovaj sloj je fokusiran isključivo na poslovnu logiku, bez potrebe za upravljanjem nitima ili memorijom na niskom nivou. Time se postiže čitljiv i održiv kod čak i u kompleksnim sistemima.

### Goroutines

Goroutine su lagane jedinice izvršavanja koje Go runtime mapira na manji broj operativnih niti. Njihovo kreiranje je izuzetno jeftino, što omogućava pokretanje velikog broja konkurentnih zadataka. Runtime dinamički raspoređuje goroutine na dostupne procesorske resurse. Ovakav model omogućava visoku skalabilnost bez opterećenja sistema velikim brojem OS niti.

### Channels

Channel-i predstavljaju osnovni mehanizam za komunikaciju i sinhronizaciju između goroutine-a. Umesto deljenja memorije uz zaključavanja, Go promoviše razmenu podataka kroz channel-e. Ovaj pristup smanjuje rizik od race condition-a i olakšava razumevanje toka podataka. Channel-i su duboko integrisani u runtime i efikasno podržani od strane scheduler-a.

### Sync Package

Sync paket pruža klasične sinhronizacione primitive kao što su mutex-i, wait-group-e i atomic operacije. Koristi se u situacijama gde channel-i nisu najprikladnije rešenje. Ovi mehanizmi omogućavaju finiju kontrolu nad deljenim resursima. Runtime obezbeđuje da se oni efikasno integrišu u konkurentni model izvršavanja.

### Scheduler

Scheduler je ključna komponenta Go runtime-a zadužena za raspoređivanje goroutine-a na procesorske niti. On implementira M:N model, gde se veliki broj goroutine-a mapira na manji broj OS niti. Scheduler dinamički balansira opterećenje kako bi se maksimalno iskoristili dostupni CPU resursi. Ovaj mehanizam je transparentan za programera i ne zahteva manuelno upravljanje.

### Memory Management i Garbage Collector

Go runtime sadrži automatski garbage collector koji upravlja alokacijom i dealokacijom memorije. Cilj mu je da minimizuje pauze u radu aplikacije, što je posebno važno za serverske sisteme. Programer nije u obavezi da ručno oslobađa memoriju, čime se smanjuje rizik od curenja memorije. Upravljanje memorijom je optimizovano za dugotrajne i konkurentne aplikacije.

### Netpoller

Netpoller je komponenta runtime-a koja efikasno upravlja mrežnim I/O operacijama. Umesto blokirajućih poziva, Go koristi događajno-orijentisan model zasnovan na epoll/kqueue mehanizmima operativnog sistema. Ovo omogućava da veliki broj mrežnih konekcija bude opslužen sa malim brojem niti. Netpoller je ključan za visoke performanse mrežnih servisa pisanih u Go-u.

### System Calls Interface

Ovaj sloj predstavlja interfejs između Go runtime-a i operativnog sistema. Runtime grupiše i optimizuje sistemske pozive kako bi smanjio overhead. Time se postiže efikasna interakcija sa OS-om bez izlaganja programera kompleksnosti sistemskih poziva. Ovaj sloj omogućava prenosivost Go aplikacija između različitih platformi.

## Bezbednost

Go pruža solidnu osnovu za izgradnju bezbednih aplikacija kroz strogu tipizaciju, automatsko upravljanje memorijom i odsustvo uobičajenih klasa grešaka poput buffer overflow-a. Dodatno, standardna biblioteka nudi ugrađenu podršku za kriptografiju, TLS i bezbednu mrežnu komunikaciju. Posebnu pažnju treba posvetiti pravilnom rukovanju konkurentnim strukturama i validaciji ulaznih podataka.

Ipak, više o bezbednosnim aspektima će biti razmatrano u sledećim koracima analize sistema i njegovih komponenti.

## Izvori
- https://medium.com/hprog99/concurrency-in-go-a-deep-dive-2abbb4838984
- https://www.linkedin.com/pulse/internal-workings-gos-concurrency-model-david-solis/
- https://medium.com/@omidahn/concurrency-in-go-understanding-goroutines-and-channels-f094f79fcc10
