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
