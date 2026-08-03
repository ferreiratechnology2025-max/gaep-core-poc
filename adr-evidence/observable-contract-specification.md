# Observable Contract Specification (OCS)

> **O que permanece arquitetural quando há vários mecanismos equivalentes.**
> A regra "mecanismo equivalente → não pertence à Architecture Specification"
> vale para mecanismos, **não para contratos observáveis**. Um contrato é o
> ponto de acoplamento entre arquiteturas concorrentes: define
> interoperabilidade, não implementação. Este documento congela o critério de
> o que *é* um contrato observável e o inventário v1 (o que o harness já
> declara como contrato). **O harness é um verificador de contratos
> observáveis**: arquiteturas A–E podem mudar completamente por dentro; o que
> devem preservar é eventos, propriedades e contratos.

## 1. Definição e critérios

Um **contrato observável** é uma especificação que:

1. **É solução-livre** — não nomeia mecanismo (nada de registrador, token,
   plano, log, mailbox, célula, WAIT_OUTCOME, ExecutionPlan).
2. **É observável** — pode ser verificado a partir de eventos/fatos emitidos
   pela arquitetura, sem inspeção interna ("espionar por baixo").
3. **É o ponto de acoplamento** — arquiteturas concorrentes interagem com ele;
   é o que permite trocar uma implementação por outra sem trocar o observador.
4. **Tem critério operacional** — PASS/FAIL/UNKNOWN pré-definidos (§2,
   `architectural-properties.md`).
5. **Sobrevive à substituição de mecanismo** — o contrato permanece válido
   quando o mecanismo interno muda (é *isso* que o torna arquitetural).

> **Teste da bala:** se trocar o mecanismo interno por outro (capability token
> → lease → ACL) mantém o mesmo vocabulário observável e os mesmos vereditos,
> então o vocabulário é contrato; o mecanismo, não. Se o vocabulário muda com
> o mecanismo, ele não é contrato — é detalhe de implementação vazado.

## 2. Critério operacional de aprovação (formato)

Para cada propriedade, o critério é declarado em três ramos:

```
P2 — Authorization Before Effect

PASS   Nenhum evento ExternalEffect ocorre antes de AuthorizationGranted
       da mesma capability.
FAIL   Existe pelo menos um ExternalEffect anterior à autorização.
UNKNOWN Não houve efeito externo no cenário.
```

Contrato **não exercitado** não conta como passou (UNKNOWN ≠ PASS) — ver
regra anti-benchmark em `architectural-properties.md`.

## 3. Inventário v1 — o que o harness já congela como contrato

### 3.1 Vocabulário de eventos observáveis (solução-livre)

Implementado em `harness/events.go`. Doze fatos observáveis, sem linguagem de
solução:

| Evento | Significado observável |
|---|---|
| `AdmissionRequested` | uma ação foi apresentada para admissão |
| `AdmissionAccepted` / `AdmissionRejected` | decisão de admissão (a entrada da P7) |
| `AuthorizationGranted` / `AuthorizationDenied` | autorização emitida/negada por capability |
| `CapabilityStarted` / `CapabilitySucceeded` / `CapabilityFailed` | ciclo de vida da capability |
| `RollbackExecuted` | estado restaurado após falha |
| `ExternalEffect` | efeito externo observável produzido |
| `AuditWritten` | registro de auditoria |
| `TimedOut` | execução pendente expirou |

**Regra de evolução:** novos eventos exigem emenda deste documento **antes**
de entrarem no código — o vocabulário é contrato, não detalhe de implementação.

### 3.2 Propriedades observáveis P1–P8

| Property | Resultado observável |
|---|---|
| P1 | execução é auditável (todo evento de estado tem auditoria posterior) |
| P2 | autorização precede qualquer efeito externo |
| P3 | falha bloqueia efeitos posteriores até nova autorização |
| P4 | estado parcial é recuperável (falha → rollback antes de recomeço) |
| P5 | execução assíncrona é tolerada (sem efeito após timeout) |
| P6 | governança independe da semântica de domínio (estrutural) |
| P7 | decisão de admissão é determinística (mesma entrada, mesma decisão) |
| P8 | execuções são isoladas (eventos de uma capability são contíguos) |

Critérios operacionais completos em `architectural-properties.md`;
implementação em `harness/oracle.go`.

### 3.3 O Property Oracle como verificador de contratos

`harness/oracle.go` implementa o contrato: recebe apenas traces de eventos,
não conhece nenhuma arquitetura, e devolve vereditos. **É o ponto de
acoplamento** — A–E só precisam honrar o vocabulário; o Oracle não muda
quando uma arquitetura muda.

## 4. Mecanismo vs contrato — o exemplo canônico

```
Capability Token ─┐
Lease             ├─ mecanismos diferentes
ACL               ─┘
        │ todos exigem o mesmo contrato observável
        ▼
      Authorize(...)
```

- **Mecanismo** → desce para ADR / Reference Implementation.
- **Contrato** (`Authorize(...)`, o vocabulário de eventos) → permanece
  arquitetural: é o ponto de acoplamento entre arquiteturas concorrentes.

No GAEP: se E1 mostrar que A (registradores) e B (context object + policy
pipeline) preservam P1–P8, o resultado **não** é "registradores são
arquitetura" nem "context objects são arquitetura" — é "o vocabulário de
eventos + P1–P8 são o contrato arquitetural; R0–R7 desce para ADR".

> **Nota de escopo:** `Authorize(...)` acima é o exemplo **ilustrativo** do
> critério — contratos de *interface* (assinaturas de função) ainda não fazem
> parte do inventário congelado v1 (que hoje cobre apenas vocabulário de
> eventos + P1–P8). Se o E1 mostrar que um contrato de interface como
> `Authorize(...)` é o ponto de acoplamento real entre rivais, ele entra como
> emenda (novo item do inventário), seguindo a mesma regra de evolução do §3.1.

## 5. Fronteira com a Structural Cost Matrix

Dois mecanismos equivalentes que preservam os mesmos contratos podem ainda
diferir em **custo estrutural** (complexidade, estados, acoplamento,
verificabilidade, superfície de manutenção). A decisão de desempate não
muda o contrato — muda *qual* mecanismo vira Reference Implementation. Ver
`structural-cost-matrix.md`.
