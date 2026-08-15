# BLX para Android

APK que embute o servidor BLX dentro do app: um WebView carrega `http://127.0.0.1:8080/`
gerado localmente. Funciona 100% offline, sem PC e sem internet.

## Estrutura

- `app/src/main/AndroidManifest.xml` — manifesto do app (permissão INTERNET, cleartext para localhost)
- `app/src/main/java/com/brunlx/blx/MainActivity.java` — WebView + gerenciamento do processo do servidor
- `build-apk.sh` — build completo sem Gradle (aapt2, d8, zipalign, apksigner)
- `keystore.jks` — keystore de assinatura (gerado no 1º build; **não commitar**)

## Requisitos

- Go (para compilar o servidor `GOOS=android GOARCH=arm64`)
- JDK (javac + keytool)
- Python 3 (empacotamento)
- Internet na 1ª execução (baixa o Android SDK se ausente)

## Como construir

```bash
make apk          # ou: VERSION=0.2.3 ./android/build-apk.sh
```

O SDK é baixado automaticamente para `$HOME/Android/Sdk` na primeira execução.
O APK sai em `android/BLX-<versão>.apk` (padrão: tag git atual).

## Como funciona

1. O binário `cmd/server` é compilado para `android/arm64` e embutido em `assets/`.
2. No app, o binário é extraído para `filesDir`, chmod 755 e iniciado via
   `ProcessBuilder` com `HOST=127.0.0.1 PORT=8080 -no-browser`.
3. O WebView espera o servidor responder em `127.0.0.1:8080` e carrega a interface.
4. Ao fechar o app, o processo é encerrado. Logs em `filesDir/blx-server.log`.

## Assinatura

Um keystore auto-assinado é gerado na primeira execução (`android/keystore.jks`).
Serve para instalação manual (sideload). Para publicar na Play Store é preciso
gerar um keystore de produção e guardá-lo com segurança.
