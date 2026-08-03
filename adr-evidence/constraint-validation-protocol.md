# Constraint Validation Protocol (CVP)

> **O fechamento do ciclo.** "Quem definiu as Needs?" levou a
> "Constraints também são hipóteses". Mas **quem valida, e com base em quê?**
> Sem uma resposta operacional, a validação é subjetiva e o viés apenas sobe
> um nível: sai dos mecanismos e passa para as constraints. Este documento é
> o protocolo que transforma cada constraint em um **objeto científico
> rastreável** — com ciclo de vida, critérios de evidência, separação de
> papéis e registro auditável. Com ele, toda a cadeia (problema → decisão
> arquitetural) é um processo de hipóteses sucessivamente testadas, não uma
> sequência de premissas assumidas.

## 1. Ciclo de vida de uma constraint

```
                    ┌──────────────┐
                    │  Proposed    │  candidata proposta (com teste de falsificação)
                    └──────┬───────┘
                           │ validação (ver §3)
              ┌────────────┼──────────────┐
              ▼            ▼              ▼
        ┌──────────┐  ┌───────────┐  ┌──────────┐
        │ Validated│  │Reformulated│  │ Rejected │
        └──────────┘  └───────────┘  └──────────┘
              │              │              │
              │              └──► nova proposta (v2) com teste corrigido
              ▼
     entra na cadeia: Need → Property → …
```

- **Proposed (candidata)** — uma afirmação de resultado esperado, solução-livre,
  com **teste de falsificação observável** obrigatório (o cenário que mostraria
  que ela é excessiva, insuficiente ou mal formulada). Sem teste de
  falsificação, não entra no protocolo.
- **Validated** — confirmada pelo domínio **e** não violada pelo harness.
- **Reformulated** — o teste de falsificação a mostrou excessiva/insuficiente,
  mas o *resultado* que ela expressava segue obrigatório → reescrever como v2
  e reiniciar o ciclo.
- **Rejected** — o domínio ou um experimento mostra que a restrição é
  dispensável (o resultado que ela exigia não é exigido pelo problema).

**Anti-axioma (regra de ouro):** *nada é axiomático porque veio primeiro.*
Uma constraint Validated pode ser reaberta se um experimento posterior a
violar, se o domínio mudar, **ou se o Problema (P0) mudar** — os **galhos
alcançáveis no DAG de proveniência** (PDP §1) voltam a Proposed; uma
redefinição (MAJOR) reavalia o DAG inteiro (ver `problem-definition-protocol.md`,
§1 e §5). Validada ≠ imutável; significa apenas "até aqui, resistiu à
falsificação".

## 2. As seis perguntas operacionais

### P1 — Como uma constraint candidata é proposta?

Qualquer pessoa do projeto (domínio ou harness) propõe, preenchendo um
**CVR** (§5) com: origem (qual elemento do Problema P0 ela deriva), a
afirmação em **resultado puro** (lint de higiene de `problem-definition.md`),
e o **teste de falsificação** observável. A proposta entra como
`Status: Proposed` e não tem autoridade antes da validação.

### P2 — Que evidência pode validá-la?

Duas evidências independentes, **ambas** necessárias:

1. **Evidência de domínio** — confirmação externa à arquitetura de que o
   resultado é obrigatório para o problema (não derivado de nenhuma solução).
2. **Evidência de harness** — os cenários relevantes (S1–S5) não violam a
   restrição sob nenhuma arquitetura candidata que a preserve; um contraexemplo
   (arquitetura que passa no harness violando a constraint) **reformula** ou
   **rejeita**.

Sem a evidência de domínio, uma constraint validada só pelo harness pode ser
um artefato das próprias propriedades P1–P8 (circularidade). Sem a evidência
de harness, é só opinião de domínio.

### P3 — Que evidência pode refutá-la?

- **Contraexemplo de domínio** — o problema, declarado pelo domínio, não
  exige aquele resultado (→ Rejected).
- **Contraexemplo de harness** — uma arquitetura conceitualmente independente
  preserva P1–P8 **sem** satisfazer a constraint como formulada. Isso mostra
  que a formulação é excessiva (→ Reformulated) ou que a restrição é
  dispensável (→ Rejected), dependendo de o *resultado* subjacente continuar
  obrigatório.

A refutação **não precisa** esperar a validação: um CVR pode ser aberto e
fechado como Rejected no mesmo ciclo.

### P4 — Quando é reformulada em vez de rejeitada?

Reformular quando o **resultado** que a constraint expressa permanece
obrigatório para o problema, mas a **formulação** foi mostrada excessiva ou
insuficiente (granularidade, escopo, condição). Rejeitar quando o próprio
resultado deixa de ser exigido. Regra prática:

- a formulação está errada, o resultado está certo → **Reformulated** (v2);
- o resultado está errado (não é exigido) → **Rejected**.

Exemplo registrado (C1): "efeitos incorretos não podem ocorrer sem
autorização" — se autorização em lote cobre N efeitos e o problema segue
satisfeito, a granularidade é excessiva → **Reformulated** (por-ação → por-lote),
não Rejected (o resultado "nada sem autorização" permanece).

### P5 — Quem tem autoridade para promover de Candidate para Validated?

**Separação de dois papéis — o proponente não valida a própria proposta:**

| Papel | Quem | Pode fazer |
|---|---|---|
| **Proponente** | quem abre o CVR | propor, aportar evidências, defender |
| **Validador** | quem não propôs; o harness (oracle) + o domínio | decidir o status final |

**Nota de projeto solo:** em um projeto de um pesquisador, proponente e
validador podem ser a mesma pessoa. Quando a identidade se sobrepõe, o CVR
deve declarar essa sobreposição como **limitação de conflito de interesse**
explícita (campo "Nota"); a separação de papéis é a regra de autoridade, mas
a honestidade de registro é o que a torna verificável.

- A promoção a **Validated** exige: (a) concordância do domínio (externo à
  arquitetura) **e** (b) veredito do harness (não violação sob cenários).
  Um só papel não promove.
- A decisão é registrada no CVR com data e responsável (§5).
- **Nenhum conceito entra na Architecture Specification por autoridade de
  quem o propôs** — só por evidência de quem validou.

### P6 — Como a decisão fica registrada e auditável?

Todo CVR é um arquivo versionado em `adr-evidence/cvrs/` com o esquema do §5,
imutável após fechamento (novas versões geram um novo CVR ou uma emenda).
O histórico de decisões é auditável pelo diff de git. Cada CVR referenciado
pela linha correspondente da necessity matrix (`architectural-necessity-matrix.md`),
pelo `problem-definition.md` e indexado no **Traceability Registry**
(`traceability-registry.md`) com as cinco colunas: origem · protocolo ·
evidência · versão · auditoria.

## 3. Critérios de veredito

| Status | Condição operacional | Exemplo |
|---|---|---|
| **Validated** | Domínio confirma obrigatoriedade **e** harness não viola sob cenários | C3 após bateria S1 sem efeito derivado pós-falha |
| **Reformulated** | Teste de falsificação mostra formulação excessiva/insuficiente; resultado subjacente obrigatório | C1 granularidade por-ação → por-lote |
| **Rejected** | Domínio ou harness mostra resultado dispensável | C4 se o domínio declarar execução síncrona |

Veredito **UNKNOWN** é possível (evidência insuficiente): a constraint
permanece *Proposed* e não entra na cadeia até ter veredito.

## 4. Onde entra na cadeia

```
Realidade do domínio (fonte última — o mundo real)
    ↓  PDP valida a representação, não a realidade
Problem Definition (P0, versionado)     ← hipótese raiz: PDP
    ↓  (mudança em P0 ⇒ cascata de invalidação pelo DAG de proveniência)
Candidate Constraints ──► CVR (este protocolo) ──► Validated / Reformulated / Rejected
    ↓
Validated Constraints
    ↓
Architectural Needs (lint de higiene)
    ↓
Observable Properties (P1–P8)
    ↓
Observable Contracts (observable-contract-specification.md)
    ↓
Candidate Mechanisms
    ↓
Experiments → Evidence
    ↓
Architectural Decision
```

As constraints não são requisitos dados; são **hipóteses com ciclo de vida**.
A cadeia inteira herda essa propriedade: nenhum nível nasce "por autoridade"
— nem o próprio problema (PDP). **Versionamento:** cada CVR carimba a versão
de P0 contra a qual a constraint foi validada (`for P0 vX.Y.Z`), para que a
cascata de invalidação saiba exatamente quais vereditos perdem validade
quando P0 muda.

## 5. Esquema do Constraint Validation Record (CVR)

Arquivo: `adr-evidence/cvrs/CVR-<nn>-<slug>.md`

| Campo | Conteúdo |
|---|---|
| CVR | id sequencial |
| Constraint | id (C1–C7) e formulação (resultado puro) |
| P0 (versão) | `for P0 vX.Y.Z` — contra qual versão do problema o veredito vale; sem isso, a cascata de invalidação (PDP) não sabe o que revalidar |
| Origem | elemento do Problema P0 que a deriva |
| Hipótese | afirmação testável + teste de falsificação observável |
| Evidência favorável | domínio (quem confirmou) + harness (cenários que não violam) |
| Evidência contrária | contraexemplos de domínio e/ou harness |
| Cenários que a falsificam | ids S1–S5 (ou novos) |
| Status | Proposed / Validated / Reformulated (v2) / Rejected |
| Data da decisão | data + responsável validador |
| Próximo passo | se Reformulated: o que muda na v2; se Rejected: o que a substitui (se algo) |
