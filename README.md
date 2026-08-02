# BLX

Gerador de comandos de pentest: o profissional **escolhe a ferramenta**, o app
**pergunta como e com o que** ela será usada e, ao final, entrega **comandos
prontos para copiar** (ou baixar como script `.sh`).

![stack](https://img.shields.io/badge/stack-Go_1.26-brightgreen) ![ui](https://img.shields.io/badge/UI-Vanilla_JS_SPA-22d3ee) ![test](https://img.shields.io/badge/tests-go_test-2ea043)

> Ferramentas de teste de intrusão **devem** ser usadas somente em sistemas e
> redes para os quais você possui **autorização formal e escrita**. Uso não
> autorizado constitui crime (no Brasil, art. 154-A do Código Penal e Lei
> 12.737/2012). Veja o [Aviso legal](#aviso-legal).

---

## Funcionalidades

- **Catálogo com 24 ferramentas** e fluxos de perguntas específicos:
  Nmap, Gobuster, Nikto, SQLMap, Metasploit, Hydra, Hashcat, Aircrack-ng,
  tcpdump, Impacket/NetExec, **Mimikatz**, **Caido**, **BloodHound**, **Fluxion**,
  **ffuf**, **msfvenom**, **masscan**, **beEF**, **Kismet**, **XSSer**,
  **Commix**, **Sherlock**, **Holehe** e **Reverse Genie (rDNS)**.
- **Wizard em 3 passos**: escolher ferramenta → responder configuração →
  receber comandos.
- **Comandos reais e prontos**, com hints, notas, alertas de risco e
  formatação de sintaxe.
- **Copiar** um comando, **copiar todos** ou **baixar um script `.sh`**
  executável (scripts do Metasploit são embutidos via heredoc em `.rc`).
- **Validação dupla**: no front e no back (campos obrigatórios, valores de
  opções, parsing seguro com *shell-quoting* de valores especiais e rejeição
  de metacaracteres em campos sensíveis como LHOST/interface).
- **API REST documentada**, testes automatizados e frontend embutido no
  binário (um único binário, sem Node nem build tooling).
- **Servidor endurecido por padrão**: bind em `127.0.0.1`, CORS desabilitado
  (opt-in via env) e cabeçalhos de segurança (CSP, X-Frame-Options, etc.).

## Requisitos

| Necessário para | Requisito |
| --------------- | --------- |
| Compilar e rodar | [Go](https://go.dev/dl/) 1.26+ e `make` |
| Build do instalador Windows | Wine + [Inno Setup](https://jrsoftware.org/isinfo.php) (ISCC) — ver [Instalação no Windows](#instalação-no-windows-distribuição) |
| Binário do Windows | apenas `make build-windows` (cross-compile, sem Wine) |

Não há dependências externas de Go (`go.mod` sem `require`). O frontend é
embutido no binário via `go:embed` — **nenhum** Node/npm é necessário.

## Instalação e execução

### 1. Clone

```bash
git clone https://github.com/Brunlx/BLX.git
cd BLX
```

### 2. Compilar

```bash
make build          # gera bin/blx
```

### 3. Executar

```bash
make run            # build + sobe em http://localhost:8080
```

Ou, sem `make`:

```bash
go build -o bin/blx ./cmd/server
PORT=8080 bin/blx
```

Abra **http://localhost:8080** no navegador. Para servir para outras máquinas
na rede (ex.: notebooks da equipe), rode com `HOST=0.0.0.0`.

### Variáveis de ambiente

| Variável      | Padrão      | Descrição                                        |
| ------------- | ----------- | ------------------------------------------------ |
| `HOST`        | `127.0.0.1` | Endereço de bind. Use `0.0.0.0` para acesso LAN. |
| `PORT`        | `8080`      | Porta HTTP.                                      |
| `CORS_ORIGIN` | (vazio)     | Habilita CORS apenas para esta origem (dev).     |

## Como usar

1. Na página inicial, **escolha uma ferramenta** (ex.: Nmap).
2. Responda às perguntas do wizard (alvos, tipo de varredura, opções…).
3. Receba os comandos gerados com sintaxe, hints e avisos de risco.
4. **Copie** o comando individual, **copie todos** ou **baixe o script `.sh`**
   e execute no seu terminal (não esqueça o `chmod +x` ao baixar scripts).

## Testes e qualidade

```bash
make test     # go test ./...
make vet      # go vet ./...
make fmt      # gofmt -w .
```

O repositório inclui CI no GitHub Actions (`go vet`, `go test` e build para
Linux e Windows a cada push/PR).

## Instalação no Windows (distribuição)

Gere o instalador `BLX-Setup.exe` para distribuir aos notebooks Windows da
equipe:

```bash
# Linux (uma única máquina de build):
make build-icon      # regenera o ícone embutido no .exe (opcional)
make build-windows   # gera bin/blx.exe com o ícone
make build-installer # gera output/BLX-Setup.exe via Inno Setup (wine)
```

O instalador:

- instala em `Program Files\BLX`;
- cria atalhos no Menu Iniciar e Área de Trabalho;
- exibe os termos de uso (`LICENSE.txt`) antes de instalar;
- é assinado? **Não** — por não possuir certificado de assinatura de código,
  o **Windows Defender / SmartScreen exibe alerta** ao executar. O usuário
  clica em *Mais informações > Executar assim mesmo*. Para eliminar o aviso,
  a organização precisa assinar o `.exe` com um certificado de código
  (ex.: via Azure Trusted Signing) ou criar exceção em políticas de TI.

O app sobe em **http://localhost:8080** (bind `127.0.0.1`). Para acesso a
partir de outras máquinas da rede, rode com `HOST=0.0.0.0` e crie uma regra
de firewall no Windows (há instrução comentada em `installer.iss`).

## Estrutura

```
BLX/
├── cmd/
│   ├── server/main.go            # entrypoint + graceful shutdown
│   └── genicon/main.go           # gerador do ícone (assets/icon.ico)
├── internal/
│   ├── api/                      # handlers HTTP, middleware, testes
│   │   ├── server.go
│   │   └── handlers.go
│   ├── static/                   # frontend embutido via go:embed
│   │   └── web/
│   │       ├── index.html
│   │       └── assets/{styles.css, app.js}
│   └── tools/                    # catálogo + motor de geração
│       ├── catalog.go            # modelos, validação, registro
│       ├── tools.go              # definições das 24 ferramentas
│       ├── generators.go         # geração de comandos por ferramenta
│       └── helpers.go            # shell-quoting e utilitários
├── assets/                       # ícone + imagens do wizard
│   ├── icon.ico / icon.png
│   ├── wizard-large.png          # painel lateral do instalador (164x314)
│   └── wizard-small.png          # ícone pequeno do instalador (55x58)
├── .github/workflows/ci.yml      # CI (vet, test, build Linux/Windows)
├── installer.iss                 # script do instalador Windows (Inno Setup)
├── LICENSE.txt                   # termos exibidos na instalação
├── Makefile                      # build, run, test, vet, fmt
└── go.mod
```

## API REST

| Método | Rota                 | Descrição                                       |
| ------ | -------------------- | ----------------------------------------------- |
| GET    | `/api/health`        | healthcheck                                     |
| GET    | `/api/tools`         | categorias + catálogo completo (com perguntas)  |
| GET    | `/api/tools/{id}`    | detalhe de uma ferramenta                       |
| POST   | `/api/generate`      | gera comandos a partir das respostas            |
| GET    | `/`                  | frontend SPA                                    |

Exemplo de geração:

```bash
curl -s -X POST http://localhost:8080/api/generate \
  -H 'Content-Type: application/json' \
  -d '{
    "toolId": "nmap",
    "answers": {
      "targets": "10.10.10.0/24",
      "scanType": "aggr",
      "verbose": "true"
    }
  }'
```

Resposta:

```json
{
  "toolId": "nmap",
  "toolName": "Nmap",
  "risk": "medium",
  "commands": [
    { "title": "Varredura principal", "language": "shell",
      "code": "nmap -A -vv 10.10.10.0/24", "hint": "…" }
  ],
  "notes": ["…"],
  "warnings": ["Varredura de redes sem autorização é ilegal…"]
}
```

## Adicionar uma nova ferramenta

1. Defina o `Tool` (perguntas) em `internal/tools/tools.go`.
2. Escreva a função geradora em `internal/tools/generators.go`.
3. Registre no `registerAll()` em `internal/tools/tools.go`.
4. Adicione um teste em `internal/tools/tools_test.go`.

Tipos de pergunta suportados: `text`, `number`, `select`, `multi`, `boolean`.

## Licença

Distribuído sob a **Licença MIT**. Veja os termos completos em
[LICENSE.txt](LICENSE.txt) (exibidos também durante a instalação no Windows).

> Ferramentas de segurança ofensiva **devem** ser usadas somente em sistemas e
> redes para os quais você possui autorização formal e escrita. O autor não se
> responsabiliza pelo uso indevido desta ferramenta.

## Aviso legal

Ferramentas de teste de intrusão **devem** ser usadas somente em sistemas e
redes para os quais você possui **autorização formal e escrita**. Uso não
autorizado constitui crime (no Brasil, art. 154-A do Código Penal e Lei
12.737/2012). A responsabilidade pelo uso é sempre de quem executa.
