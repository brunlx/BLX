# Política de Segurança

## Escopo suportado

Versões com correções de segurança recebem suporte na **última release** e no
branch `main`. Versões antigas não recebem correções.

## Reportando uma vulnerabilidade

Este projeto é uma ferramenta ofensiva legítima: ele **gera** comandos de
pentest. Vulnerabilidades relevantes aqui são as que afetam **quem usa o BLX**
(ex.: execução de código remoto no servidor, bypass de validação de entrada,
exposição de dados) — não a natureza das ferramentas que o BLX invoca.

Para reportar:

- **Não** abra uma issue pública para falhas que permitam exploração remota.
- Envie um e-mail para **hshshehs407@gmail.com** com:
  - descrição clara da falha;
  - passos para reproduzir (ou PoC);
  - versão afetada;
  - impacto estimado.

Respondemos em até **7 dias úteis**. Após a correção ser publicada, uma
advisory (GHSA ou release note) será emitida creditando o reportante.

## Divulgação responsável

- Reporte **antes** de divulgar publicamente.
- Conceda **90 dias** para a correção antes de publicar detalhes.
- Se a correção exigir mais tempo, coordene conosco antes da divulgação.

## Recomendações para usuários

- Use o BLX somente em sistemas e redes para os quais você possui **autorização
  formal e escrita**.
- Mantenha o `HOST` em `127.0.0.1` (padrão) — só exponha na rede com regras de
  firewall e CORS configurados.
- Não rode o BLX como root/administrador quando não necessário.
