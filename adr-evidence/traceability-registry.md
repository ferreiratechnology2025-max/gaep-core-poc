# Traceability Registry (TR)

> **O índice mestre.** Os protocolos definem **como** cada elemento evolui
> (PDP → problema; CVP → constraints; OCS → contratos; SCM → desempate). O
> Traceability Registry registra **como todos eles se conectam** — o elo
> PDR → P0 vX.Y.Z → CVRs → Needs → Properties → Contracts → Experiments →
> Evidence → ADR, percorrível em qualquer direção. Não é um protocolo novo:
> é governança de conexão. Com ele, toda a metodologia pode ser percorrida de
> maneira completamente auditável, e qualquer decisão responde \"contra qual
> versão do problema?\" e \"com base em qual evidência?\".

## A cadeia canônica (estrutura de referência)

```
Realidade do domínio (fonte última — o mundo real)
    ↓  PDP valida a representação, não a realidade
P0 vX.Y.Z (versionado)
    ↓  CVP
Validated Constraints
    ↓
Needs → Properties → Contracts (OCS)
    ↓
Mechanisms (SCM desempata) → Reference Implementation / ADR
    ↓
Harness (verificador de contratos observáveis)
    ↓
Evidence → Architectural Decision
```

Cada transformação possui **cinco colunas**: **origem · protocolo · evidência
· versão · auditoria**.

## Índice mestre (estado atual)

### Problema

| Id | Versão | Status | Protocolo | Registro | Evidência |
|---|---|---|---|---|---|
| P0 | v1.0.0 | Proposed (rascunho) | PDP | PDR pendente | a validar com o domínio |

### Constraints (DAG de proveniência: Constraint → Need → Property → Cenário)

| Id | Versão | Status | Protocolo | Registro | Derivada de | Property / Cenário |
|---|---|---|---|---|---|---|
| C1 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N1 → P2 (S2) |
| C2 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N3 → P1 (S1) |
| C3 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N4 → P4 (S1); N8 → P3 (S1) |
| C4 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N2 → P5 (S3) |
| C5 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N5 → P7 (S5) |
| C6 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N6 → P8 (S4) |
| C7 | v1.0 | Proposed | CVP | CVR pendente | P0 v1.0.0 | N7 → P6 (dual-run) |

### Needs

| Id | Versão | Origem | Resultado puro (resumo) |
|---|---|---|---|
| N1 | v1.0 | C1 | nenhum efeito observável sem autorização válida |
| N2 | v1.0 | C4 | consegue interromper efeitos de execução pendente |
| N3 | v1.0 | C2 | execução reconstruível do estado persistido |
| N4 | v1.0 | C3 | volta a estado consistente conhecido após falha |
| N8 | v1.0 | C3 | falha não deixa efeitos derivados aplicados |
| N5 | v1.0 | C5 | mesma entrada, mesma decisão de admissão |
| N6 | v1.0 | C6 | uma execução não interfere no resultado de outra |
| N7 | v1.0 | C7 | governança independe da semântica de domínio |
| N9 | v1.0 | (meta) | o conjunto P1–P8 distingue arquiteturas? |

### Properties (P1–P8)

| Id | Versão | Critérios | Implementação |
|---|---|---|---|
| P1–P8 | v1.0 | `architectural-properties.md` | `harness/oracle.go` |

### Contracts (inventário v1 — congelados na OCS)

| Contrato | Versão | Elementos |
|---|---|---|
| Vocabulário de eventos | v1.0 | 12 eventos observáveis (`harness/events.go`) |
| Critérios operacionais | v1.0 | PASS/FAIL/UNKNOWN por P1–P8 |

### Experiments & Evidence (todos `for P0 v1.0.0`)

| Exp | Artefato | Log de evidência | P0 |
|---|---|---|---|
| E0 | `experiments/e0_minimality` | `E0_evidence.log` | for P0 v1.0.0 |
| E1 | `experiments/e1_registers` | `E1_evidence.log` | for P0 v1.0.0 |
| E1v2 | `experiments/e1_models` | `E1v2_evidence.log` | for P0 v1.0.0 |
| E1arch | `experiments/e1_architectures` + `harness/` | `E1arch_evidence.log` | for P0 v1.0.0 |
| **Baseline** | Re-execução completa — `go test -count=1 -v ./...` (E0 + E1 + E1v2 + E1arch/harness) | `BASELINE_rerun.log` | for P0 v1.0.0 |

> **Baseline — verificação dos cinco invariantes:** (1) *unicidade ativa* —
> uma única linha para a re-execução; (2) *imutabilidade referenciada* — o
> registro aponta para `BASELINE_rerun.log` (conteúdo imutável no histórico;
> re-execuções futuras geram **novo** log com sufixo de data, nunca
> sobrescrevem este); (3) *rastreabilidade a P0* — `for P0 v1.0.0`;
> (4) *evidência mínima* — o log é a evidência verificável; (5) *aciclicidade*
> — nenhuma alteração no DAG de proveniência. Resultado: **todos os pacotes
> PASS** (exit 0) e **estabilidade 0 falhas em 10 execuções** — nenhuma falha
> ou invalidação concreta encontrada; portanto, pela regra de ouro, **nenhum
> PDR/CVR novo é aberto** neste baseline.

> **Nota de carimbo:** os logs atuais (saída bruta de `go test`) precedem a
> convenção de carimbo e são marcados `for P0 v1.0.0` **por declaração neste
> índice**. Logs de experimentos futuros carimbam `for P0 vX.Y.Z` no próprio
> arquivo (PDP §7).

### Decisões / ADRs

Nenhuma ainda. A primeira decisão carimba `decided against P0 v1.0.0` e entra
aqui com as cinco colunas.

## Como percorrer

- **Para frente** (problema → decisão): `P0 v1.0.0` → C1 → N1 → P2 → cenário
  S2 → contrato (vocabulário) → veredito do Oracle → evidência → decisão.
- **Para trás** (decisão → problema): decisão `decided against P0 v1.0.0` →
  evidência (log, `for P0 v1.0.0`) → experimento → contrato → P2 → N1 → C1 →
  P0 v1.0.0.

## O ciclo de evolução (fechado operacionalmente, aberto epistemicamente)

```
Operacionalmente fechado — a cadeia de derivação, percorrível de ponta a ponta:

Realidade → PDP → P0 vX.Y.Z → CVP → Constraints → Needs → Properties → OCS
→ Contracts → Mechanisms → Harness → Evidence → Decision

Epistemicamente aberto — o retorno que reabre o problema:

Experimentos → evidência empírica → Q7 ("P0 ainda representa o domínio?")
→ novo PDR → novo P0 → cascata (alcançabilidade no DAG) → re-derivação
→ volta à cadeia
```

O sistema é **fechado operacionalmente** (toda transformação tem origem,
protocolo, evidência, versão e auditoria — nada salta) e **aberto
epistemicamente** (a realidade pode sempre reabrir P0 — nada é selado).

## Camadas da metodologia

| Camada | Responsabilidade |
|---|---|
| Realidade do domínio | Fonte externa de verdade (o axioma externo) |
| PDP | Representação auditável do domínio |
| CVP | Validação das restrições |
| OCS + SCM | Verificação arquitetural e custos |
| TR | Índice e governança de toda a cadeia |

Nenhuma camada invade a responsabilidade da outra.

## Matriz de responsabilidades e dependências (congelada)

> **Quem pode modificar o quê, quem pode depender de quem e quem tem a
> autoridade final (ownership)** entre os cinco componentes do núcleo. Não
> cria mecanismo — congela a divisão de responsabilidades (antes espalhada
> pelos documentos) e impede deriva arquitetural futura. Toda a cadeia herda:
> axioma epistemológico explícito, DAG de proveniência, protocolos
> normativos, invariantes de governança, **responsabilidades e ownership
> congelados** e regras de evolução delimitadas.

### Matriz A — Responsabilidade (quem pode alterar o quê)

| Componente | Pode alterar | Não pode alterar |
|---|---|---|
| PDP | P0 (representação do problema) | CVR (constraints), OCS (contratos) |
| CVP | Constraints (status via CVR) | P0 |
| OCS | Contratos (vocabulário + critérios P1–P8) | Constraints |
| SCM | Custos (matriz de desempate) | Contratos (OCS), P0 |
| TR | Registro (índice) | Nenhum artefato (não produz conteúdo, não decide) |

### Matriz B — Dependência (quem pode depender de quem)

| Componente | Pode depender de | Não pode depender de |
|---|---|---|
| PDP | Realidade do domínio (axioma), TR (registro de PDRs) | CVP, OCS, SCM — nada derivado (formulação solução-livre) |
| CVP | PDP (P0), cenários do harness (evidência S1–S5), TR (registro de CVRs) | OCS (contratos não validam constraints), SCM |
| OCS | Properties/Needs (P1–P8, derivadas via CVP), TR (registro) | P0 direto (sem passar por Properties), SCM |
| SCM | OCS (contratos), evidência do harness, TR (registro) | CVP, PDP (não reabre constraints nem problema) |
| TR | Todos (indexa tudo) | — |

> **O harness (verificador de contratos) é uma dependência *operacional*, não
> *normativa* (ver "Tipos de dependência" abaixo):** CVP, OCS e SCM podem
> consumir cenários e vereditos do harness como evidência sem violar as
> proibições de dependência acima — a proibição "não pode depender de OCS"
> refere-se a *contratos* (vocabulário + critérios), não à ferramenta de
> verificação. O TR é espelho (Regra derivada 3): lido por todos, não
> alimenta decisão de nenhum.

### Matriz C — Ownership (autoridade final)

| Artefato | Autoridade (owner) |
|---|---|
| P0 | PDP |
| Constraints | CVP |
| Observable Contracts | OCS |
| Structural Cost Matrix | SCM |
| Traceability Registry | TR |

**Autoridade ≠ permissão (Matriz A):** a Matriz A diz quem *pode* alterar; a
Matriz C diz quem *aceita* a mudança — a autoridade final de aprovação. O
owner pode delegar a edição (permissão), mas não a aceitação: nenhuma mudança
num artefato é válida sem o veredito do owner. O TR é owner do próprio
registro; isso não o autoriza a alterar nenhum outro artefato (Matriz A).

**Ownership de artefatos derivados:** Needs (N1–N9) pertencem ao CVP
(derivadas das constraints que ele valida); Properties (P1–P8) pertencem ao
OCS (são os critérios dos contratos que ele congela). A tabela acima lista os
cinco artefatos raiz; os derivados herdam o owner da origem na cadeia.

### Tipos de dependência (normativa × operacional)

| Tipo | Participa da governança | Exemplo |
|---|---|---|
| **Normativa** | Sim — influencia decisões, faz parte da cadeia metodológica | PDP → CVP |
| **Operacional** | Não — apenas infraestrutura, não influencia decisões | CVP → Harness |

A Matriz B lista as **dependências normativas**; as menções ao harness nas
linhas de CVP e SCM são as **exceções operacionais** (declaradas na nota
acima). Dependências operacionais (importação de código, APIs, ferramentas)
estão fora da governança — o harness é o caso canônico: infraestrutura de
evidência usada por CVP/OCS/SCM sem participar das decisões. Classificar uma
dependência como operacional não a isenta da regra "violação = emenda": se
ela passar a influenciar decisões, vira normativa e entra na Matriz B.

### Regras derivadas

1. **Direção da cadeia** — dependências seguem a cadeia canônica (montante →
   jusante); nenhum componente depende de algo a jusante de si, **exceto o
   TR**, que indexa todos (espelho, Regra derivada 3).
2. **Solução-livre** — o PDP não depende de nada derivado: a formulação do
   problema nunca referencia contrato, constraint ou custo.
3. **Espelho** — o TR é lido por todos, mas **nenhum componente decide com
   base nele**; ele registra, não alimenta a cadeia.
4. **Violação = emenda** — dependência ou modificação fora da matriz exige
   emenda registrada (mesmo regime dos protocolos) antes de ser aceita.

## Invariantes do TR (componente formal de governança)

O TR deixou de ser apenas um índice: é **parte da especificação normativa** da
metodologia (papel de CMDB dos artefatos metodológicos). Como tal, possui
invariantes próprios — violação = TR inconsistente → emenda registrada (mesmo
regime dos protocolos: proposta + registro + data):

1. **Unicidade ativa** — todo artefato governado possui **exatamente um**
   registro ativo no TR; versões antigas permanecem no histórico (git), não
   no índice ativo.
2. **Imutabilidade referenciada** — todo registro referencia uma **versão
   imutável** do artefato (caminho + versão); o TR nunca aponta para conteúdo
   mutável.
3. **Rastreabilidade a P0** — toda decisão experimental é rastreável a **um
   único nó P0** por um **caminho válido no DAG** de proveniência.
4. **Evidência mínima** — todo registro possui **pelo menos uma evidência
   verificável** (log, veredito do Oracle, confirmação de domínio).
5. **Aciclicidade do DAG** — alterações no DAG de proveniência preservam
   **aciclicidade** (ordenação topológica sempre possível).

Os invariantes 1, 3 e 5 são verificáveis mecanicamente (scripts futuros); 2 e
4, por inspeção.

## Regras (manutenção)

As regras abaixo são os **procedimentos de manutenção** que operam os
invariantes acima — não os duplicam: especificam *como* mantê-los (p.ex.: a
Regra 1 executa o Invariante 1; a Regra 2 sustenta os Invariantes 1–2).

1. **Exatidão** — executa o Invariante 1: cada elemento aparece exatamente
   uma vez no índice ativo.
2. **Elo = referência** — sustenta os Invariantes 1–2: todo link é um caminho
   de arquivo verificável para uma versão imutável.
3. **Cinco colunas** — executa o Invariante 4: origem, protocolo, evidência,
   versão, auditoria; sem isso, o elemento não entra no índice.
4. **Mesmo commit** — uma mudança de versão de um artefato é refletida no
   índice no mesmo commit (o TR é mantido junto com os artefatos).
5. **O TR não decide** — registra conexões. Não é um protocolo, não promove
   nada, não valida nada.

## Baseline Metodológico (v1.0.0-methodology-baseline)

> O núcleo metodológico é considerado congelado para fins de evolução
> arquitetural. Alterações futuras deverão ser motivadas por evidência
> empírica obtida durante a validação prática, e não por refinamentos
> especulativos.

**Bloqueio de expansão:** nenhuma nova camada de governança, protocolo,
tabela, invariante ou artefato conceitual será criada a partir deste
baseline. O ciclo especulativo/documental está encerrado; este índice não
ganha novas seções conceituais, apenas novas linhas de registro e evidência.

**Regra de ouro (PDR/CVR):** só abrir novos PDRs ou CVRs se (e somente se) a
re-execução dos experimentos e a coleta de evidências do mundo real
encontrarem **falhas ou invalidações concretas** — nunca por refinamento
especulativo.

**Status:** Fase de Especificação: **encerrada**. Fase de Validação Prática:
**iniciada** — foco 100% no runtime e no harness; toda saída de execução é
indexada abaixo seguindo os cinco invariantes.
