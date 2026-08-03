# Problem Definition Protocol (PDP)

> **O último axioma cai.** O CVP tornou as Constraints hipóteses; mas o
> Problema (P0) continuava sendo tratado como verdade por autoridade. Se P0
> muda — e problemas mudam — toda a árvore derivada
> (Constraints → Needs → Properties → Contracts) torna-se potencialmente
> inválida. Este protocolo é o equivalente do CVP para P0: como um problema
> nasce, quem o altera, quais mudanças são editoriais e quais invalidam
> constraints validadas, e como a propagação dessa invalidação é registrada e
> auditada. A partir daqui **nem o problema é axioma**: ele participa do mesmo
> ciclo de hipóteses, evidências e decisões auditáveis.
> **Ponto de parada (Trilema de Münchhausen encerrado por axioma, não
> eliminado):** o protocolo não valida a realidade — valida a representação
> auditável de um problema observado no domínio. O mundo real é a única fonte
> última: o axioma externo ao sistema (ver §0).

## 0. A fonte última: a realidade (o ponto de parada)

**O protocolo não valida a realidade; valida a representação.** P0 é uma
formulação auditável de um problema *observado no domínio* — e é só isso que
este protocolo valida. O mundo real continua sendo a única fonte última: é
dele que vêm os sinais de invalidação (Q7/§4) e é dele que emana a autoridade
do domínio (Q2). O PDP controla apenas **a representação** desse mundo.

**Consequência — o Trilema de Münchhausen é encerrado por escolha explícita
de axioma, não eliminado:** como em sistemas formais, o axioma não é provado
dentro do sistema — ele é declarado. O ponto de parada epistemológico é:
*a realidade do domínio é o axioma externo ao sistema*. O protocolo não
precisa validar a si mesmo porque não é a fonte de verdade — a realidade é. A
pergunta "quem valida o PDP?" tem resposta trivial: *ninguém*; o que existe é
uma **escolha de axioma declarada e auditável**, não uma prova. O PDP não diz
o que é verdadeiro no mundo; diz como manter uma formulação auditável daquilo
que o domínio declara. Se a formulação deixar de representar o mundo, a
evidência do mundo (§4) dispara o protocolo — não um meta-protocolo.

## 1. O problema como nó raiz de um DAG de proveniência

```
P0 (raiz do DAG de proveniência)
 ├── C1 ──► N1 ──► P2
 ├── C2 ──► N3 ──► P1
 ├── C3 ──► N4 ──► P4
 ├── C3 ──► N8 ──► P3
 ├── C4 ──► N2 ──► P5
 ├── C5 ──► N5 ──► P7
 ├── C6 ──► N6 ──► P8
 └── C7 ──► N7 ──► P6
      (N9 — meta — deriva do conjunto P1–P8, não de uma constraint)
```

**DAG, não árvore:** as arestas são de derivação (Constraint → Need →
Property → contrato/cenário). Um nó pode ter vários pais e vários filhos
(C3 deriva N4 e N8; um contrato pode servir a várias propriedades). O
mapeamento canônico vive em `problem-definition.md` §3; este DAG é a forma
executável dele.

**Propagação computável:** a cascata de uma mudança em P0 não é heurística —
é **alcançabilidade no DAG**. Dada a mudança classificada (§4): os artefatos
que entram em reavaliação são exatamente os **alcançáveis a partir das
arestas afetadas** (MINOR) ou do nó raiz (MAJOR); os demais permanecem
válidos. Isso reduz drasticamente o retrabalho: P0 pode mudar e apenas as
constraints do ramo alcançável voltarem a Proposed.

Regra fundamental: **se P0 muda, o conjunto alcançável da mudança é
reavaliado conforme a classificação** — editorial → nada; semântica (MINOR)
→ galhos alcançáveis voltam a Proposed e os demais passam por revisão de
proveniência; redefinição (MAJOR) → o DAG inteiro é reavaliado. Em nenhum
caso uma constraint derivada permanece automaticamente válida sob um
problema diferente.

## 2. Ciclo de vida de P0

```
Proposed ──validação──► Validated
   ▲                        │
   │                        ├─ mudança MINOR (escopo) ──► Revised ──revalidação──► Validated (P0 v1.1.0)
   └── nova versão          │
       v2.x nasce aqui      └─ mudança MAJOR (redefinição) ──► Superseded ──► novo Proposed v2.x

Toda mudança em P0 dispara a cascata de invalidação (§5):
  MINOR → constraints tocadas voltam a Proposed; intocadas → revisão de proveniência
  MAJOR → todas as constraints voltam a Proposed; a árvore re-deriva do novo P0
```

- **Proposed** — formulação solução-livre, com **sinais de invalidação**
  explícitos (o que no mundo real indicaria que o problema mudou). Sem sinais,
  não entra no protocolo.
- **Validated** — aceito pelo domínio (externo à arquitetura) e coerente com
  a evidência acumulada; permanece enquanto não houver sinal de invalidação.
- **Revised** — mudança **semântica** de escopo (§4) que exige reavaliação
  parcial da árvore.
- **Superseded** — mudança **major** (§4): o problema foi redefinido; a versão
  antiga deixa de ser referência e toda a árvore volta à estaca inicial.

**Anti-axioma para P0:** Validated ≠ imutável. Um problema validado significa
apenas "até aqui, resistiu aos sinais de invalidação". Os sinais vêm do
**domínio** (o mundo real), não do harness — o harness não pode detectar que o
problema mudou, só que a implementação não o satisfaz mais.

**Transições não terminais:** Revised (MINOR) volta a Validated após
revalidação da nova versão (P0 v1.1.0 → Validated). Superseded (MAJOR) gera
uma nova versão v2.x que nasce como Proposed. Nenhum estado é "fim de linha":
tudo segue o ciclo de hipóteses.

## 3. As sete perguntas operacionais

> Q = pergunta operacional do PDP (distinto das Properties P1–P8).

### Q1 — Como um problema é proposto?

Preenchendo um **PDR** (§6): formulação em resultado puro (lint de
`problem-definition.md` §1), origem (qual contexto do domínio o motiva) e
**sinais de invalidação**. Entra como `Status: Proposed`; não tem autoridade
antes da validação.

### Q2 — Quem pode modificá-lo?

**Separação de dois papéis** (mesma regra do CVP): quem propõe não valida.
A **autoridade sobre o problema é do domínio** — externa à arquitetura —, não
de quem implementa ou modela. O harness **não** modifica P0 (ele só verifica
contratos). Em projeto solo, a sobreposição de papéis é declarada como
limitação de conflito de interesse no PDR (mesma nota do CVP §P5).

### Q3 — Quais mudanças são apenas editoriais?

**Editorial** (PATCH no versionamento, §7): não altera o conjunto de
resultados exigidos pelo problema — ortografia, formatação, reformulação que
preserva o mesmo significado. Não exige reavaliação de nada na árvore.
Regra de bolso: "se a frase nova produz as mesmas constraints na matriz de
necessidade, é editorial".

### Q4 — Quais mudanças invalidam Constraints já validadas?

Toda mudança que altere o **conjunto de resultados exigidos** é
**semântica** (MINOR) ou **major** (MAJOR):

| Tipo | Efeito em P0 | Efeito na árvore | Exemplo |
|---|---|---|---|
| **Editorial** | PATCH | nenhum — nada é reavaliado | reescrita de redação sem mudança de significado |
| **Semântica** (escopo) | MINOR | **constraints tocadas voltam a Proposed** (aquelas cujos testes de falsificação ou resultados esperados são afetados); as intocadas passam por revisão de proveniência e permanecem | "impedir efeitos sem autorização" → "impedir efeitos sem autorização **exceto em modo recovery aprovado**" |
| **Major** (redefinição) | MAJOR | **todas** as constraints voltam a Proposed; Needs/Properties/Contracts re-derivados | "impedir efeitos sem autorização" → "permitir execução offline, com autorização sincronizada posteriormente" |

A pergunta decisiva para classificar: **"o problema continua sendo o mesmo
problema?"** — sim com escopo alterado → MINOR; não → MAJOR.

### Q5 — Como a propagação da invalidação é registrada e auditada?

- **PDR imutável** após fechamento, versionado em `adr-evidence/pdrs/` (§6).
- Cada artefato derivado (C*, N*, P*, contratos) carrega o carimbo
  **"valid for P0 vX.Y.Z"** — a versão do problema contra a qual foi derivado.
- Uma mudança em P0 gera um **registro de propagação**: para cada artefato da
  árvore, o status transita de "valid for P0 v1.0" para
  "**requires re-validation under P0 v1.1**" (MINOR) ou "**derived under
  superseded P0 v1.0 → re-derive**" (MAJOR).
- O histórico completo é auditável pelo diff de git (mesmo mecanismo do CVP).

### Q6 — Como toda a cadeia é reavaliada após uma mudança em P0?

Passos obrigatórios:

1. Classificar a mudança (§4) e registrar o PDR com a data e o responsável.
2. **MINOR:** marcar as constraints afetadas como Proposed (novo CVR ou
   reabertura), rodar revisão de proveniência nas intocadas, re-validar as
   afetadas pelo CVP, e re-derivar apenas os galhos tocados.
3. **MAJOR:** marcar **toda** a árvore como derivada de problema supersedido;
   Constraints voltam a Proposed; Needs/Properties/Contracts são re-derivados
   a partir do novo P0 (a matriz de necessidade é reaberta linha por linha).
4. Registrar a transição de versão de cada artefato (§7) para que qualquer
   decisão futura responda "foi tomada contra qual versão do problema?" — e
   indexar tudo no Traceability Registry (§8).

### Q7 — Qual evidência mostra que P0 deixou de representar o domínio?

Esta pergunta define **quando iniciar um novo PDR** — o gatilho que impede
P0 de permanecer "Validated" indefinidamente. Toda evidência vem do mundo
real (§0), nunca do harness:

1. **Nova legislação / regulamentação** — o ambiente normativo muda os
   resultados exigidos.
2. **Mudança de requisitos do cliente/domínio** — novo caso de uso que o P0
   atual não cobre (ex.: execução offline com autorização posterior).
3. **Descoberta de caso de uso inexistente** — um resultado que P0 exigia é
   declarado irrelevante pelo domínio.
4. **Evidência experimental repetida de incompatibilidade** — cenários válidos
   violam sistematicamente P0 como formulado, mesmo após refino das
   propriedades (N9 + CVP em cascata).

Quem declara que P0 deixou de representar o domínio é o **domínio**; a
evidência é registrada no PDR, que abre a transição para Revised/Superseded
(§2). O harness **não** produz essa evidência — ele só diz se a implementação
satisfaz o contrato.

## 4. Sinais de invalidação de P0 (quando o protocolo dispara)

A pergunta operacional **Q7** define a evidência; esta seção lista as fontes
de sinal. O protocolo **não espera** que alguém "perceba" que o problema
mudou; ele exige sinais declarados desde o PDR. Fontes de sinal:

1. **Ambiente normativo** — nova legislação/regulamentação muda os resultados
   exigidos (Q7.1).
2. **Domínio** — novo requisito de negócio que o P0 atual não cobre (ex.: o
   exemplo da crítica: execução offline com autorização posterior) (Q7.2).
3. **CVP em cascata** — uma constraint reformulada/rejeitada com recorrência
   sugere que o problema está mal formulado (revisar P0 antes de reformular a
   mesma constraint pela segunda vez) (Q7.4).
4. **Discriminabilidade (N9)** — se P1–P8 nada distingue entre arquiteturas
   apesar de cenários válidos, suspeita-se das **propriedades**, mas se isso
   persistir após refino, suspeita-se do **problema** (propriedades derivadas
   dele) (Q7.4).

Isso fecha o **ciclo de evolução**: Experimentos → evidência empírica → Q7
("P0 ainda representa o domínio?") → novo PDR → novo P0 → cascata
(alcançabilidade no DAG) → re-derivação → volta à cadeia. O sistema é
**fechado operacionalmente** (toda transformação tem origem, protocolo,
evidência, versão e auditoria) e **aberto epistemicamente** (a realidade pode
sempre reabrir P0 — nada é selado). O diagrama completo vive no Traceability
Registry (`traceability-registry.md`).

## 5. Cascata de invalidação (o caso da crítica)

```
P0 v1.0: "O sistema deve impedir efeitos sem autorização."
    ↓  seis meses depois, domínio declara:
P0 v2.0 (MAJOR): "O sistema deve permitir execução offline,
                 com autorização sincronizada posteriormente."
    ↓
C1 (autorização prévia)  → Proposed  (teste de falsificação mudou de natureza)
C2, C3, C4, C5, C6, C7   → Proposed  (derivadas de problema supersedido)
N1–N8                    → re-derivar (N1 "sem autorização válida" incompatível
                                     com autorização posterior)
P1–P8                    → re-derivar do novo conjunto de Needs
Contratos (vocabulário)  → revisar: novo resultado exigido pode exigir novos
                           eventos (ex.: AuthorizationSynchronizedLater) →
                           emenda na OCS antes de código
```

**Nenhuma constraint antiga permanece Validated sob o novo P0.**

## 6. Esquema do Problem Definition Record (PDR)

Arquivo: `adr-evidence/pdrs/PDR-<nn>-<slug>.md`

| Campo | Conteúdo |
|---|---|
| PDR | id sequencial |
| Problema | P0, versão (vX.Y.Z), formulação (resultado puro) |
| Origem | contexto do domínio que motiva o problema |
| Sinais de invalidação | o que indicaria que o problema mudou |
| Classificação da mudança | editorial / semântica (MINOR) / redefinição (MAJOR) |
| Impacto na árvore | constraints afetadas (retornam a Proposed) e as intocadas |
| Registro de propagação | por artefato: "valid for P0 v1.0" → "requires re-validation" |
| Data da decisão | data + responsável validador |
| Próximo passo | quais CVRs abrir, quais galhos re-derivar |

## 7. Versionamento semântico da cadeia

Convenção para responder **"esta decisão foi tomada contra qual versão do
problema?"**:

```
P0 v1.0.0
│  ├── C1 v1.0 (Validated, for P0 v1.0.0)
│  ├── C2 v1.0 (Validated, for P0 v1.0.0)
│  └── C3 v1.1 (Reformulated, for P0 v1.0.0)
↓
Needs v1.2     (derivado de P0 v1.0.0, reflete C1–C7 atuais)
↓
Properties v1.2 (P1–P8, for Needs v1.2)
↓
Contracts v1.2  (vocabulário + critérios, for Properties v1.2)
```

Regras:

- **P0** — MAJOR = redefinição; MINOR = mudança semântica de escopo; PATCH =
  editorial.
- **Artefatos derivados** — cada um tem a própria versão e o carimbo
  `for P0 vX.Y.Z`. **Dois versionamentos independentes:** `P0 v2.1.0` e
  `Constraint C3 v1.4.0` evoluem sem obrigação de sincronia — uma constraint
  pode mudar três vezes sem tocar P0, e P0 pode mudar com apenas as
  constraints do ramo alcançável do DAG (§1) precisando de nova validação.
  Não existe acoplamento artificial entre as versões.
- **Decisão arquitetural** — obrigatoriamente carimba `decided against P0
  vX.Y.Z`; sem o carimbo, a decisão não entra no registro.
- **Evidência** (logs de experimento, vereditos do harness) também carimba
  `for P0 vX.Y.Z` — a versão do problema contra a qual foi produzida. **Evidência
  não atravessa MAJOR:** uma decisão sobre P0 v2.x não pode citar evidência
  produzida sob P0 v1.x; se quiser, precisa re-executar os cenários.
- **Transições MAJOR** zeram a árvore (todos os artefatos re-derivados); as
  versões antigas permanecem no histórico (git), nunca são reescritas.

**Estado atual (registro de versão):** P0 v1.0.0 (Proposed/rascunho);
C1–C7 v1.0 (Proposed); Needs/Properties/Contracts v1.0 (derivados de P0
v1.0.0). Nenhum CVR fechado; zero PDRs. O registro canônico do estado atual
vive no topo de `problem-definition.md`.

## 8. Traceability Registry (índice mestre)

Os protocolos definem **como** cada elemento evolui (PDP → problema; CVP →
constraints; OCS → contratos; SCM → desempate). O **Traceability Registry**
(`traceability-registry.md`) registra **como todos se conectam** — a ligação
PDR → P0 vX.Y.Z → CVRs → Needs → Properties → Contracts → Experiments →
Evidence → ADR, percorrível em qualquer direção. Todo PDR, CVR, emenda de
contrato e decisão é indexado lá com as cinco colunas: **origem · protocolo ·
evidência · versão · auditoria**. O TR não decide nada — registra conexões.
Desde a revisão que o formalizou, ele é **parte da especificação normativa**
da metodologia (papel de CMDB dos artefatos), com **cinco invariantes
próprios** — unicidade ativa, imutabilidade referenciada, rastreabilidade a
P0, evidência mínima e aciclicidade do DAG — e a **matriz congelada de
responsabilidades, dependências e ownership** entre os cinco componentes
(quem pode alterar o quê; quem pode depender de quem; autoridade final por
artefato; dependência normativa vs operacional; ver
`traceability-registry.md`).
