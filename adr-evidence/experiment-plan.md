# Plano de Experimentos — Architecture Validation Harness

> **Fluxo (invertido):**
>
> ```
> Realidade do domínio (fonte última) → PDP → Problem Definition (P0,
> versionado) → Candidate Constraints → Validated Constraints → Architectural
> Needs → Observable Properties → Observable Contracts → Candidate Mechanisms
> → Experiments → Evidence → Architectural Decision
> ```
>
> Cada transformação tem origem · protocolo · evidência · versão · auditoria
> — o índice mestre das conexões é o **Traceability Registry**
> (`traceability-registry.md`).
>
> As Needs são derivadas das **Constraints** (`problem-definition.md`) e
> passam pelo lint de higiene (resultado puro, sem mecanismo). **Constraints
> também são hipóteses** e sua validação é operacional e rastreável pelo
> **Constraint Validation Protocol** (`constraint-validation-protocol.md`):
> ciclo de vida, seis perguntas operacionais (quem propõe, que evidência
> valida/refuta, quando reformular vs rejeitar, quem promove, como registra),
> CVR por constraint e separação de papéis (proponente ≠ validador). **E o
> próprio Problema (P0) é hipótese** — regido pelo **Problem Definition
> Protocol** (`problem-definition-protocol.md`): ciclo de vida de P0,
> classificação de mudanças (editorial/semântica/redefinição), cascata de
> invalidação (P0 muda → constraints voltam a Proposed) e versionamento
> semântico da cadeia (toda decisão carimba `decided against P0 vX.Y.Z`). P1–P8 também são hipóteses;
> cada propriedade é gerada por uma necessidade (ver
> `architectural-necessity-matrix.md`, a fonte da verdade). Arquiteturas rivais são **conceitualmente independentes** e
> tentam preservar as garantias; toda avaliação passa pelo **Property
> Oracle** (`harness/oracle.go`): arquiteturas só emitem eventos, o Oracle
> decide PASS/FAIL/UNKNOWN. O objetivo dos experimentos é **eliminar
> mecanismos concorrentes**, não provar o GAEP.

## Regras congeladas

1. **Independência conceitual** — cada arquitetura concorrente nasce de
   princípios próprios. Não pode ser uma versão renomeada da arquitetura A
   nem compartilhar as estruturas internas de A por conveniência. Se um
   rival recriar R0–R7 disfarçados, isso é um **resultado** (registradores
   como consequência inevitável do problema), não uma falha do experimento.
2. **Property Oracle** — arquiteturas nunca escrevem "PASS". Elas emitem
   eventos observáveis; só o Oracle determina os vereditos, com critérios
   operacionais pré-definidos (P1–P8 em `architectural-properties.md`).

### Regra de Decisão E1 (pré-registrada)

**Unidade de comparação:** os veredictos do Oracle sobre P1–P8 por cenário. O vocabulário de eventos é a interface comum (fixa por construção); conformidade a ele é pré-condição, não objeto da comparação. Trajetórias de eventos podem diferir entre arquiteturas. Custos estruturais não entram aqui; são desempate posterior via Structural Cost Matrix, somente em caso de empate em P1–P8.

**Protocolo:** cada arquitetura roda sob exatamente o mesmo protocolo experimental (mesmos cenários, mesmas entradas) e é avaliada exclusivamente pelo Oracle. A equivalência é aferida par a par.

**Equivalência observacional:** duas arquiteturas são equivalentes se, para todo cenário e toda propriedade P1–P8, o Oracle produzir o mesmo veredicto.

**Divergência:** qualquer célula (cenário × propriedade) com veredictos distintos caracteriza não equivalência e é registrada com a célula responsável. Divergência dispara investigação de causa (diferença real de mecanismo vs. bug de implementação) antes de qualquer conclusão sobre a hipótese.

**UNKNOWN:** não é neutro. PASS-vs-UNKNOWN ou FAIL-vs-UNKNOWN conta como divergência e é registrado como lacuna de observabilidade.

**Exceção estocástica:** se alguma candidata introduzir não determinismo legítimo (concorrência com múltiplos resultados válidos, retry, escalonamento assíncrono), a unidade passa a ser a distribuição dos veredictos sobre N execuções, com N e o critério de equivalência distribucional pré-registrados antes de executar essa candidata. Sem N e critério pré-registrados, a candidata não entra na matriz.

## Arquiteturas candidatas (princípio de origem de cada uma)

| ID | Arquitetura | Princípio próprio (não deriva de A) |
|---|---|---|
| **A** | GAEP atual | Registradores $R0$–$R7$ + microkernel + lifecycle |
| **B** | Context Object + Policy Pipeline | Objetos de contexto (session, transaction, capability, audit) atravessando um pipeline de políticas; sem células indexadas |
| **C** | Event Sourcing | Estado derivado de log de eventos; rollback = replay |
| **D** | Actor Model | Estado encapsulado em atores com caixas postais |
| **E** | Hierarchical State Machine | Máquina de estados hierárquica por capacidade |

**Restrição:** nenhuma delas simula registradores internamente. B não possui
"células R0–R7 disfarçadas"; se ao implementar B descobrirmos que ele precisa
recriar o equivalente, registramos a descoberta na matriz como evidência de
que os registradores são consequência, não escolha.

## Especificações operacionais das concorrentes (E1)

> Mapa do que é **determinado** pelo princípio declarado de cada concorrente
> vs. o que é **pendente** de decisão do Arquiteto. Não especifica
> implementação: apenas explicita as decisões em aberto (resolvidas na Fase
> 2, gate do Arquiteto). Item 8 é **descritivo** (leitura de `kernel/`): a
> variável "ciclo de execução" da Arquitetura A **não está em `kernel/`**
> (o pacote contém apenas o register file; o lifecycle está documentado em
> `falsification-matrix.md` e `models_comp/lifecycle.go`) — a comparação de
> ciclo usa essa fonte, não `kernel/`.

### B — Context Object + Policy Pipeline

1. **Princípio de origem:** "Objetos de contexto (session, transaction,
   capability, audit) atravessando um pipeline de políticas; sem células
   indexadas" (`experiment-plan.md`, tabela de candidatas).
2. **Representação de estado:** DECORRE — estado em objetos de contexto
   tipados (session, transaction, capability, audit); explicitamente **sem
   células indexadas** (≠ array fixo de A).
3. **Sincronização/concorrência:** NÃO DECORRE → DECISÃO DO ARQUITETO
   NECESSÁRIA: disciplina de concorrência (mutex único, por contexto, ou
   lock-free?).
4. **Aplicação de policy:** DECORRE (parcial) — políticas aplicadas via
   **pipeline**; não decorre a composição/ordem dos estágios → DECISÃO DO
   ARQUITETO NECESSÁRIA: ordem e granularidade dos estágios.
5. **Ciclo de execução:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   mapeamento do ciclo (admissão→autorização→execução→auditoria) sobre a
   travessia do pipeline.
6. **Proveniência:** fonte usada: princípio declarado apenas; juízo de
   independência é do Arquiteto → DECISÃO DO ARQUITETO NECESSÁRIA: atestar
   independência (garantia de processo, não verificável no artefato).
7. **Determinismo:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   estágios puros e ordem fixa? (fonte potencial de variabilidade; aciona a
   cláusula estocástica da Regra de Decisão E1 se N e critério não
   pré-registrados).
8. **Variáveis vs A (descritivo):** MUDA — representação de estado
   (contextos sem células) e aplicação de policy (pipeline vs policy
   por-registro). MANTÉM — nenhuma assertável do princípio. Sincronização e
   ciclo: DECISÃO DO ARQUITETO NECESSÁRIA (itens 3 e 5).

### C — Event Sourcing

1. **Princípio de origem:** "Estado derivado de log de eventos; rollback =
   replay" (`experiment-plan.md`, tabela de candidatas).
2. **Representação de estado:** DECORRE — estado é projeção derivada de um
   log de eventos (fonte de verdade = log); rollback = replay (≠ `Restore`
   de A).
3. **Sincronização/concorrência:** NÃO DECORRE → DECISÃO DO ARQUITETO
   NECESSÁRIA: serialização de appends ao log; concorrência de projeções?
4. **Aplicação de policy:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   política aplicada na admissão do evento, na projeção, ou em ambos?
5. **Ciclo de execução:** DECORRE (parcial) — ciclo = eventos → projeção →
   efeito; não decorre quem decide admissão/execução → DECISÃO DO ARQUITETO
   NECESSÁRIA: origem dos comandos/eventos.
6. **Proveniência:** fonte usada: princípio declarado apenas; juízo de
   independência é do Arquiteto → DECISÃO DO ARQUITETO NECESSÁRIA: atestar
   independência (garantia de processo, não verificável no artefato).
7. **Determinismo:** DECORRE (condicional) — replay é determinístico **se**
   a projeção for função pura do log e a ordem de aplicação fixa → DECISÃO
   DO ARQUITETO NECESSÁRIA: pureza da projeção e ordenação (fonte potencial
   de variabilidade).
8. **Variáveis vs A (descritivo):** MUDA — representação (log+projeção vs
   células) e ciclo (replay vs `Restore`). MANTÉM — nenhuma assertável do
   princípio. Sincronização e policy: DECISÃO DO ARQUITETO NECESSÁRIA
   (itens 3 e 4).

### D — Actor Model

1. **Princípio de origem:** "Estado encapsulado em atores com caixas
   postais" (`experiment-plan.md`, tabela de candidatas).
2. **Representação de estado:** DECORRE — estado encapsulado em atores
   (inacessível externamente); comunicação via caixas postais (mensagens).
3. **Sincronização/concorrência:** DECORRE (parcial) — encapsulamento
   implica comunicação por mensagens/caixas postais (sem acesso direto ao
   estado de outro ator); não decorre a topologia de atores e o
   escalonamento → DECISÃO DO ARQUITETO NECESSÁRIA: topologia,
   escalonamento, ordem de entrega.
4. **Aplicação de policy:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   política como ator dedicado, guarda em cada ator, ou na fronteira?
5. **Ciclo de execução:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   coreografia do ciclo (admissão→autorização→execução→auditoria) entre os
   atores.
6. **Proveniência:** fonte usada: princípio declarado apenas; juízo de
   independência é do Arquiteto → DECISÃO DO ARQUITETO NECESSÁRIA: atestar
   independência (garantia de processo, não verificável no artefato).
7. **Determinismo:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA: ordem
   de entrega de mensagens e escalonamento (fonte clássica de
   não-determinismo; aciona a cláusula estocástica da Regra de Decisão E1
   se N e critério não pré-registrados).
8. **Variáveis vs A (descritivo):** MUDA — representação (estado em atores
   vs células) e sincronização (mensagens vs mutex único). MANTÉM — nenhuma
   assertável do princípio. Policy e ciclo: DECISÃO DO ARQUITETO NECESSÁRIA
   (itens 4 e 5).

### E — Hierarchical State Machine

1. **Princípio de origem:** "Máquina de estados hierárquica por
   capacidade" (`experiment-plan.md`, tabela de candidatas).
2. **Representação de estado:** DECORRE — estado = posição em máquina de
   estados hierárquica por capacidade (super/subestados por capability).
3. **Sincronização/concorrência:** NÃO DECORRE → DECISÃO DO ARQUITETO
   NECESSÁRIA: uma máquina por capacidade em concorrência; kernel de
   transição compartilhado?
4. **Aplicação de policy:** NÃO DECORRE → DECISÃO DO ARQUITETO NECESSÁRIA:
   política como guarda de transição, pré-condição de estado, ou externa?
5. **Ciclo de execução:** DECORRE (parcial) — ciclo = transições de estado
   (super/subestados); não decorre o driver de disparo (eventos, scan,
   chamadas) → DECISÃO DO ARQUITETO NECESSÁRIA: driver de disparo e origem
   das entradas.
6. **Proveniência:** fonte usada: princípio declarado apenas; juízo de
   independência é do Arquiteto → DECISÃO DO ARQUITETO NECESSÁRIA: atestar
   independência (garantia de processo, não verificável no artefato).
7. **Determinismo:** DECORRE (condicional) — dada mesma entrada e mesmo
   estado, a transição é determinística → determinismo no teste; não
   decorre o tratamento de entradas simultâneas → DECISÃO DO ARQUITETO
   NECESSÁRIA: entradas concorrentes (fonte potencial de variabilidade).
8. **Variáveis vs A (descritivo):** MUDA — representação (máquina de
   estados vs células) e ciclo (transições vs lifecycle explícito). MANTÉM
   — nenhuma assertável do princípio. Sincronização e policy: DECISÃO DO
   ARQUITETO NECESSÁRIA (itens 3 e 4).

## Fluxo de avaliação (congelado)

```
Arquitetura (A–E)
   ↓ executa cenário (S1–S5)
Eventos observáveis
   ↓
Property Oracle (harness/oracle.go)
   ↓
P1 ✓  P2 ✗  P3 UNKNOWN  …   (matriz propriedade × arquitetura)
```

## Ordem

| ID | Experimento | Objetivo | Saída / artefatos |
|---|---|---|---|
| **E0** | Problema + constraints (candidatas→validadas) + necessidades + propriedades + contratos | Congelar a cadeia Realidade→PDP→Constraint→Need→Property→Contract→Mechanisms→Evidence→Decision, com lint de higiene nas Needs, teste de falsificação por constraint (protocolo CVP), ciclo de vida do próprio problema (PDP + versionamento + DAG de proveniência) e inventário de contratos observáveis | `problem-definition.md` (topo); `problem-definition-protocol.md` (novo); `constraint-validation-protocol.md` (novo); `observable-contract-specification.md` (novo); `structural-cost-matrix.md` (novo, vazio); `traceability-registry.md` (novo — índice mestre); `architectural-necessity-matrix.md` (fonte da verdade); `architectural-properties.md`; `harness/` (oracle + S1–S5) |
| **E1** | Arquiteturas concorrentes | Implementar A–E de forma independente e rodar a bateria S1–S5 | Matriz propriedade × arquitetura (`experiments/e1_architectures`) + preenchimento da coluna "evidência" da necessity matrix; decisões por regra dos três resultados |
| **E2** | Cenário WAIT_OUTCOME | Fronteira Kernel↔Runtime | Qual arquitetura preserva P5/P8 sob governança assíncrona? (novo cenário no harness) |
| **E3** | Cenário Rollback | Necessidade do estado transacional | P4 sob falha no meio de transação (estende S1) |
| **E4** | Cenário Runtime malicioso | Validar responsabilidades | P2/P3/P8 sob Runtime adversário (outcomes fabricados, timeouts, mutações) |
| **E5** | Cenário ExecutionPlan | Comparar modelos de governança | P7/P6 sob comandos/eventos vs plano pré-voo |
| **E6** | Custos estruturais | Comparar **custos** entre arquiteturas que preservam as mesmas propriedades | Matriz Property → arquiteturas → custo estrutural → trade-offs → decisão (`structural-cost-matrix.md`, protocolo congelado; preenchida quando houver empate em P1–P8 — desempate, nunca critério de correção) |

## Mapeamento com o estado atual

| Artefato | Papel |
|---|---|
| `harness/` | Núcleo do harness (eventos + oracle + cenários) — **novo, E0** |
| `experiments/e1_models` (A/B/B′/C) | Embrião de E1 no nível de representação de estado — B′ foi o primeiro rival que preservou tudo (mas agora B é redefinido por princípios próprios, não como tradução) |
| `experiments/e1_registers` | Evidência de suporte (por que locking é necessário) |
| `experiments/e0_minimality` | Falsificação por remoção em nível de componente |
| `kernel.RegisterPolicy`, M3 rollback, I2 | Já exercitam P6, P4, P1 no nível de implementação |
| `models_comp/lifecycle.go` | Objeto de ataque de E2: hoje assume WAIT_OUTCOME como estado do kernel |
| `adr-evidence/traceability-registry.md` | Índice mestre das conexões (PDR → P0 → CVRs → Needs → Properties → Contracts → Experiments → Evidence → ADR) — mantido junto com os artefatos |

## Regras de entrada na spec

0. **Problema válido e versionado** — a decisão é tomada contra `P0 vX.Y.Z`
   (PDP); se o problema mudou, a cascata de invalidação volta os galhos
afetados a Proposed antes de qualquer decisão nova.
1. **Need** em termos de capacidade (sem mecanismo).
2. **Property** observável gerada (P-id + critério operacional).
3. **Falsifique construindo o rival independente** e rodando a bateria pelo Oracle.
4. Aplicar a regra dos três resultados (necessity matrix):
   - Nenhum mecanismo indispensável → **remover** da Architecture Specification.
   - Vários mecanismos equivalentes → propriedade fica na spec; mecanismo → **ADR / Protocol Specification**; **contrato observável pode permanecer** se for o ponto de acoplamento entre arquiteturas.
   - Apenas um mecanismo preserva → **congelar** na architecture.
5. **Validar a constraint** via **Constraint Validation Protocol**: o teste de
   falsificação a mantém (validada), reformula ou rejeita; veredito registrado
   em CVR com evidência de domínio + harness, proponente ≠ validador
   (`constraint-validation-protocol.md`).
6. Verificar **discriminabilidade** (N9): se todas as arquiteturas passam em tudo, o conjunto de propriedades não distingue → refinar/adicionar propriedades.
7. **Contrato observável** (inventário v1 em `observable-contract-specification.md`): o vocabulário de eventos + P1–P8 permanecem arquiteturais mesmo com mecanismos equivalentes — são o ponto de acoplamento; o mecanismo escolhido (se houver empate) desce a Reference Implementation por custo estrutural (`structural-cost-matrix.md`).
