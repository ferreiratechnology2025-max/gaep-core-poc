# Problem Definition, Constraints & Validation — GAEP Core

> **O nível mais alto da cadeia.** "Quem definiu as Needs?" — a resposta
> honesta é que as Needs devem provir das Constraints, e as Constraints do
> Problema. Mas **Constraints também são hipóteses**: nada impede que um
> experimento mostre que uma constraint é excessiva, insuficiente ou mal
> formulada. Elas não viram axiomas só porque aparecem antes das Needs.
> **E o próprio Problema (P0) também é hipótese** — regido pelo **Problem
> Definition Protocol** (`problem-definition-protocol.md`): ciclo de vida,
> classificação de mudanças (editorial / semântica / redefinição), cascata de
> invalidação (P0 muda → constraints voltam a Proposed) e versionamento
> semântico da cadeia inteira.
> Este documento formaliza o problema, deriva constraints **candidatas**
> com teste de falsificação explícito, e separa **mecanismo** de
> **contrato observável**.

## Registro de versão (estado atual)

| Artefato | Versão | Status | Derivado de |
|---|---|---|---|
| **P0** | v1.0.0 | Proposed (rascunho — a validar com o domínio) | — |
| C1–C7 | v1.0 | Proposed (validação pendente) | P0 v1.0.0 |
| Needs N1–N9 | v1.0 | derivadas | P0 v1.0.0 |
| Properties P1–P8 | v1.0 | hipóteses | Needs v1.0 |
| Contracts (vocabulário + P1–P8) | v1.0 | congelados na OCS | Properties v1.0 |

Regras do versionamento em `problem-definition-protocol.md` §7. Qualquer
decisão arquitetural carimba `decided against P0 vX.Y.Z`. O índice mestre das
conexões (origem · protocolo · evidência · versão · auditoria) é o
**Traceability Registry** (`traceability-registry.md`).

## A cadeia completa

```
Realidade do domínio (fonte última — o mundo real)
    ↓  o protocolo valida a representação, não a realidade
Problem Definition Protocol (PDP)
    ↓
Problem Definition (P0, versionado)     ← hipótese raiz
    ↓
Candidate Constraints        ← hipóteses: cada uma tem teste de falsificação
    ↓
Validated Constraints        ← sobrevivem à validação (ou são reformuladas/removidas)
    ↓
Architectural Needs          ← lint de higiene: resultado puro, sem mecanismo
    ↓
Observable Properties        ← critérios operacionais PASS/FAIL/UNKNOWN
    ↓
Observable Contracts         ← interfaces que todo mecanismo deve honrar
    ↓
Candidate Mechanisms
    ↓
Experiments
    ↓
Evidence
    ↓
Architectural Decision
```

**Cada transformação tem origem · protocolo · evidência · versão · auditoria**
— o índice mestre dessas conexões é o **Traceability Registry**
(`traceability-registry.md`), percorrível em qualquer direção. O sistema é
**fechado operacionalmente** (a cadeia percorre de ponta a ponta) e **aberto
epistemicamente**: o ciclo Experimentos → evidência empírica → Q7 → novo PDR
→ novo P0 pode reabrir qualquer nó — a realidade nunca é selada.

**Propagação de mudança (computável):** a cascata segue o **DAG de
proveniência** (`problem-definition-protocol.md` §1) — se P0 muda (revisão ou
redefinição via PDP), apenas o conjunto alcançável da mudança é reavaliado:
constraints alcançáveis voltam a Proposed, demais passam por revisão de
proveniência, e nada derivado do problema antigo permanece automaticamente
válido.

## 1. Problem Definition (P0 v1.0.0 — rascunho, regido pelo PDP)

- **P0 v1.0.0**: "Executar ações produzidas por IA com risco de efeitos
  incorretos." O sistema precisa governar a execução de ações cujo efeito
  pode ser incorreto — antes, durante e depois da execução.
- **Regra de independência**: nenhuma frase das seções 1–3 pode nomear
  mecanismo (registrador, kernel, célula, transação, token, plano, pipeline,
  ator, log, watchdog).
- **Sinais de invalidação de P0 (exemplos)** — o que dispararia o PDP:
  domínio declara novo requisito fora do escopo atual (ex.: execução offline
  com autorização posterior); constraint reformulada/rejeitada pela segunda
  vez sugere problema mal formulado; N9 sem discriminabilidade apesar de
  cenários válidos. Ver `problem-definition-protocol.md` §4.

## 2. Candidate Constraints (hipóteses — status: *candidata, validação pendente*)

Cada constraint tem um **teste de falsificação observável**: o cenário que
mostraria que ela é excessiva, insuficiente ou mal formulada. A validação é
feita pelo domínio + harness — nunca assumida.

| Constraint | Resultado esperado (solução-livre) | Teste de falsificação | Status |
|---|---|---|---|
| C1 | Efeitos incorretos não podem ocorrer sem autorização | Cenário mostra efeito legítimo sem autorização por-ação (ex.: autorização em lote cobre N efeitos) e o problema segue satisfeito → C1 excessiva em granularidade → reformular | Candidata — validação pendente (domínio) |
| C2 | Os efeitos de uma execução são reconstruíveis a posteriori do estado persistido | Cenário mostra que reconstruir efeitos exige informação que o estado persistido não retém, e a governança não exige essa reconstrução → C2 excessiva | Candidata |
| C3 | Uma falha não pode deixar efeitos derivados aplicados | Cenário de falha total mostra efeitos parciais *não observáveis por construção* (P3 emergente, sem mecanismo) → C3 satisfeita sem mecanismo | Candidata |
| C4 | O sistema permanece consistente sob execução assíncrona | Domínio determina que toda execução é síncrona → C4 excessiva (decisão do domínio) | Candidata |
| C5 | A mesma entrada produz a mesma decisão de admissão | Domínio aceita decisões dependentes da ordem de chegada → C5 reformulável | Candidata |
| C6 | A execução de uma ação não interfere no resultado de outra | Domínio exige compartilhamento de estado entre ações → C6 excessiva | Candidata |
| C7 | Decidir sobre uma ação não exige conhecer a semântica de domínio | Domínio exige que decisões conheçam semântica das ações → C7 reformulável | Candidata |

**Validação:** segue o **Constraint Validation Protocol**
(`constraint-validation-protocol.md`): ciclo de vida
Proposed→Validated/Reformulated/Rejected, registro em CVR (objeto rastreável
com evidências e data), separação de papéis (proponente ≠ validador; domínio
+ harness, nunca um só), e a regra anti-axioma (validada ≠ imutável). Nota de
terminologia: "Candidata" abaixo = "Proposed" no CVP. **Revalidação
obrigatória:** se P0 mudar (PDP), constraints afetadas voltam a Proposed —
nenhuma permanece Validated sob um problema diferente. Uma
constraint é **Validada** quando (a) o domínio a confirma como obrigatória e
(b) o harness não a viola sob os cenários. É **Reformulada** quando o teste
de falsificação a mostra excessiva/insuficiente mas o resultado subjacente
permanece obrigatório. É **Rejeitada** quando o domínio a declara
 dispensável.

## 3. Constraints → Needs → Properties (mapeamento com proveniência)

| Constraint | Need (resultado puro) | Property | Cenário |
|---|---|---|---|
| C1 | N1 — nenhum efeito observável ocorre sem autorização válida | P2 | S2 |
| C4 | N2 — o sistema consegue interromper efeitos de uma execução pendente | P5 | S3 |
| C2 | N3 — toda execução é reconstruível do estado persistido | P1 | S1 |
| C3 | N4 — após uma falha, o sistema volta a um estado consistente conhecido | P4 | S1 |
| C3 | N8 — uma falha não deixa efeitos derivados aplicados | P3 | S1 |
| C5 | N5 — a decisão de admissão é a mesma para a mesma entrada | P7 | S5 |
| C6 | N6 — uma execução não interfere no resultado de outra | P8 | S4 |
| C7 | N7 — decisões de governança independem da semântica de domínio | P6 | dual-run |
| — | N9 (meta) — o conjunto P1–P8 distingue arquiteturas? | — | E1 |

## 4. Higiene das Necessidades (anti-contaminação)

**Palavras proibidas em Needs:** registrador, kernel, célula, transação,
token, plano, pipeline, ator, log, watchdog, R0–R7, WAIT_OUTCOME,
ExecutionPlan.

| Frasagem | Veredito |
|---|---|
| "o sistema precisa manter um estado transacional isolado" | ✗ nomeia mecanismo |
| "o sistema precisa impedir efeitos observáveis inconsistentes" | ✓ resultado puro |
| "o núcleo não depende de semântica de domínio" | ✗ nomeia mecanismo ("núcleo") |
| "decisões de governança independem da semântica de domínio das ações" | ✓ resultado puro |

**Lint:** se uma Need nomeia palavra proibida, reescreva como resultado e
mova o mecanismo para "candidate mechanisms".

## 5. Mecanismo vs Contrato Observável (distinção conservadora)

A regra "se existe mecanismo equivalente, não pertence à Architecture
Specification" vale **para mecanismos**. **Não vale automaticamente para
interfaces ou contratos.**

- **Mecanismo** — só permanece na arquitetura se for inevitável.
- **Contrato observável** — pode permanecer mesmo com vários mecanismos
  equivalentes, **se for o ponto de acoplamento entre arquiteturas
  concorrentes** (define interoperabilidade, não implementação).

Exemplo: Capability Token, Lease e ACL são mecanismos diferentes, mas podem
exigir exatamente o mesmo contrato observável `Authorize(...)`. Nesse caso:
o mecanismo desce para ADR; **o contrato permanece arquitetural**.

No harness: o **vocabulário de eventos observáveis** (`harness/events.go`) e
os **critérios P1–P8** (`architectural-properties.md`) são contratos
observáveis — o ponto de acoplamento entre A–E. Eles permanecem na
arquitetura independentemente de quais mecanismos sobreviverem. O inventário
v1 desses contratos está congelado na **Observable Contract Specification**
(`observable-contract-specification.md`), que também define o critério do que
conta como contrato (solução-livre, observável, ponto de acoplamento,
critério operacional, sobrevive à substituição de mecanismo).

**Consequência na regra dos três resultados:** "vários mecanismos
equivalentes" → o mecanismo desce para ADR/Protocol; **o contrato fica na
spec** se for o acoplamento entre arquiteturas.

## 6. Status epistemológico

1. **Hipótese de arquitetura** — proposto, não testado.
2. **Mecanismo candidato aguardando falsificação** — evidência mostrou outro
   mecanismo preservando as mesmas propriedades (R0–R7 após E1v2;
   WAIT_OUTCOME pendente da bateria S3).
3. **Necessidade arquitetural** — apenas quando *um único* mecanismo preserva
   a propriedade; ou quando o que permanece é o **contrato observável** (ponto
   de acoplamento), não o mecanismo.

## 7. As duas perguntas obrigatórias por decisão

1. **Qual problema inevitável esta decisão resolve?**
2. **Existe outro mecanismo que resolve o mesmo problema preservando as
   mesmas propriedades?** Se "sim" → o mecanismo não pertence à Architecture
   Specification (o contrato observável pode permanecer).

## 8. Matriz de custo estrutural (congelada — vazia até o E1)

Para arquiteturas que preservam as mesmas propriedades: complexidade
estrutural, número de estados, acoplamento, verificabilidade formal, esforço
de implementação, superfície de manutenção. A matriz
Property → arquiteturas → custo → trade-offs → decisão **complementa** a de
necessidade quando há múltiplos mecanismos equivalentes. O protocolo
(dimensões, formato, regras de desempate, limitações) está congelado em
`structural-cost-matrix.md`; a matriz só é preenchida quando duas ou mais
arquiteturas empatarem em P1–P8 — é desempate, nunca critério de correção
(ver `experiment-plan.md`, E6).
