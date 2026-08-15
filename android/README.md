# BLX para Android

APK que embute o servidor BLX dentro do app: um WebView carrega `http://127.0.0.1:8080/`
gerado localmente. Funciona 100% offline, sem PC e sem internet.

## Estrutura

- `app/src/main/AndroidManifest.xml` — manifesto do app (INTERNET, cleartext p/ localhost, `extractNativeLibs`)
- `app/src/main/java/com/brunlx/blx/MainActivity.java` — WebView + gerenciamento do processo do servidor
- `app/src/main/res/` — ícones do launcher (PNGs + adaptive icon)
- `build-apk.sh` — build completo sem Gradle (aapt2, d8, zipalign, apksigner)
- `keystore.jks` — keystore de assinatura (gerado no 1º build; **não commitar**)

## Requisitos

- Go (para compilar o servidor `GOOS=android GOARCH=arm64`)
- JDK (javac + keytool)
- Python 3 (empacotamento)
- Internet na 1ª execução (baixa o Android SDK se ausente)

## Como construir

```bash
make apk          # ou: VERSION=0.2.4 ./android/build-apk.sh
```

O SDK é baixado automaticamente para `$HOME/Android/Sdk` na primeira execução.
O APK sai em `android/BLX-<versão>.apk` (padrão: tag git atual).

## Como funciona

O binário `cmd/server` é compilado com `GOOS=android GOARCH=arm64 CGO_ENABLED=0`
(PIE, com o linker bionic `/system/bin/linker64`) e empacotado como **lib nativa**
em `lib/arm64-v8a/libblxserver.so` (vem de `jniLibs/`).

Isso é obrigatório: desde o **Android 10**, o SELinux (política W^X) **bloqueia
`exec()` em arquivos de `getFilesDir()`** (`chmod +x` não resolve). O instalador
extrai a lib para `nativeLibraryDir`, que tem rótulo `exec_type` e é executável.
Por isso o binário vai em `jniLibs/` e não em `assets/`.

1. No app, o caminho da lib é resolvido via `getApplicationInfo().nativeLibraryDir`.
2. O binário é iniciado via `ProcessBuilder` com `HOST=127.0.0.1 PORT=8080 -no-browser`.
3. O WebView espera o servidor responder em `127.0.0.1:8080` e carrega a interface.
4. Ao fechar o app, o processo é encerrado. Logs em `filesDir/blx-server.log`.

O servidor é 100% em memória (não grava arquivos), então roda de qualquer cwd.

## Arquiteturas

Só **arm64-v8a** (Android 7+). O Go sem cgo/NDK não compila `android/arm` nem
`android/amd64`; a esmagadora maioria dos celulares modernos é arm64.

## Assinatura

Um keystore auto-assinado é gerado na primeira execução (`android/keystore.jks`)
com uma senha **aleatória** persistida em `android/.keystore-pass` (ambos
gitignored, não commitá-los). Para builds reproduzíveis/CI, gere o keystore uma
vez e defina a variável `KEYSTORE_PASS` com a senha usada.

Serve para instalação manual (sideload). Para publicar na Play Store é preciso
gerar um keystore de produção e guardá-lo com segurança.
