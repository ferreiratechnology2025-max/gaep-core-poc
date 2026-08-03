# GAEP Core — Architecture Validation Harness

Este repositório **não é um PoC de uma implementação específica**. É um
**Architecture Validation Harness**: um ambiente para validar hipóteses
arquiteturais fazendo arquiteturas candidatas rivais passarem pelos mesmos
cenários contra um conjunto fixo de **propriedades observáveis** — e
registrando a matriz propriedade × arquitetura como evidência.

## O contrato

- `adr-evidence/problem-definition.md` — o nível mais alto: Problem
  Definition (P0, **versionado** — hipótese raiz, não axioma) + **Candidate→
  Validated Constraints** (hipóteses com teste de falsificação — constraints
  também são validadas, não assumidas), lint de higiene (Needs como resultado
  puro), a distinção **mecanismo vs contrato observável** (o contrato pode
  ser arquitetural mesmo com mecanismos equivalentes, se for o ponto de
  acoplamento) e as duas perguntas obrigatórias por decisão.
- `adr-evidence/problem-definition-protocol.md` — **quem valida o próprio
  problema**: ciclo de vida de P0 (Proposed→Validated/Revised/Superseded),
  classificação de mudanças (editorial / semântica / redefinição), cascata de
  invalidação (P0 muda → constraints voltam a Proposed), PDR rastreável,
  **DAG de proveniência com propagação computável** e **versionamento
  semântico da cadeia inteira** (toda decisão carimba `decided against P0
  vX.Y.Z`). Ponto de parada: **o Trilema de Münchhausen é encerrado por
  escolha explícita de axioma** (a realidade do domínio é o axioma externo ao
  sistema), não eliminado — o protocolo valida a **representação** do
  problema, não a realidade.
- `adr-evidence/traceability-registry.md` — **índice mestre e governança**:
  conecta PDR → P0 vX.Y.Z → CVRs → Needs → Properties → Contracts →
  Experiments → Evidence → ADR, percorrível em qualquer direção, com cinco
  colunas por elemento (origem · protocolo · evidência · versão · auditoria)
  e **cinco invariantes formais** (unicidade ativa, imutabilidade
  referenciada, rastreabilidade a P0, evidência mínima, aciclicidade do DAG)
  — papel de CMDB, parte da especificação normativa.

- `adr-evidence/architectural-necessity-matrix.md` — **fonte da verdade**:
  cadeia Constraint→Need→Property→Mechanisms→Experiments→Evidence→Decision,
  com a regra dos três resultados (nenhum mecanismo → remover; vários →
  ADR/Protocol; apenas um → congelar). P1–P8 também são hipóteses,
  geradas por necessidades derivadas de restrições.
- `adr-evidence/constraint-validation-protocol.md` — **quem valida as
  constraints e com base em quê**: ciclo de vida
  Proposed→Validated/Reformulated/Rejected, as seis perguntas operacionais
  (quem propõe, que evidência valida/refuta, quando reformular vs rejeitar,
  quem promove, como registra), registro CVR rastreável por constraint e
  separação de papéis (proponente ≠ validador). Constraints são hipóteses,
  não axiomas.
- `adr-evidence/observable-contract-specification.md` — **o que permanece
  arquitetural com mecanismos equivalentes**: o critério do que é um contrato
  observável (solução-livre, observável, ponto de acoplamento, critério
  operacional, sobrevive à substituição de mecanismo) e o inventário v1 —
  vocabulário de eventos + P1–P8. O harness é um **verificador de contratos**.
- `adr-evidence/structural-cost-matrix.md` — **o desempate entre arquiteturas
  que preservam as mesmas propriedades**: dimensões de custo, formato,
  regras de desempate e limitações. Só entra quando há empate em P1–P8.
- `adr-evidence/architectural-properties.md` — o conjunto de propriedades
  observáveis (P1–P8), **sem mencionar soluções**, com **critérios
  operacionais PASS/FAIL/UNKNOWN** por propriedade e o vocabulário de
  eventos observáveis. É o contrato que toda arquitetura candidata deve
  satisfazer.
- `harness/` — o **Property Oracle** (implementação dos critérios) + bateria
  de cenários S1–S5. Arquiteturas **só emitem eventos**; o Oracle decide.
- `adr-evidence/falsification-matrix.md` — conceitos candidatos × como
  falsificá-los × qual propriedade deve quebrar × status.
- `adr-evidence/experiment-plan.md` — ordem dos experimentos (E0–E5), com
  as **regras congeladas**: arquiteturas concorrentes conceitualmente
  independentes; avaliação exclusivamente via Property Oracle.

## Regra metodológica

> **Nenhum conceito entra na Architecture Specification antes de sobreviver a
> um experimento de falsificação.** ADRs registram necessidade demonstrada —
> "a remoção desta solução quebra a propriedade Px" —, não escolha.

Decisão por evidência:

| Resultado | Decisão |
|---|---|
| Conceito indispensável (remoção quebra propriedade) | Architecture Specification |
| Conceito substituível mantendo as propriedades | ADR / Reference Implementation |
| Conceito removível sem perda | Eliminado |

## Estrutura

| Caminho | Papel |
|---|---|
| `harness/` | Property Oracle + eventos observáveis + bateria de cenários S1–S5 |
| `kernel/` | Núcleo da arquitetura candidata A (registradores $R0$–$R7$, policy injetada) |
| `models_comp/` | Candidatos de representação de estado (A/B/B′) + invariantes + lifecycle |
| `experiments/` | E0 (remoção de componentes), E1 (matriz de arquiteturas, scaffold), E3 (registradores A/B/B′/C) |
| `adr-evidence/` | Propriedades + oracle, matriz de falsificação, plano e logs de evidência |

## Estado atual da evidência (resumo honesto)

- **E1v2** (`experiments/e1_models`): sob sincronização igual, o Modelo B′
  (estado fechado tipado, sem registradores) preserva todas as propriedades
  medidas do Modelo C (registradores) → **registradores não provados
  necessários pelas propriedades medidas**; a superfície fechada e guardada é
  a propriedade que importa.
- **E0** (`experiments/e0_minimality`): R1/R2/R6/R7 pinados por invariantes;
  **R0/R3/R4/R5 com lacuna de requisito** (ou a propriedade é especificada,
  ou eles saem). SESSION/WATCHDOG/WAIT_OUTCOME/RELEASE são load-bearing no
  ciclo de vida.
- Próximo passo: implementar as **arquiteturas concorrentes** (A GAEP, B Context
  Object + Policy Pipeline, C event sourcing, D actor, E state machine) com
  **independência conceitual** (regra congelada) e rodá-las pela bateria
  S1–S5 via Property Oracle (ver `experiment-plan.md`, E1).

## Notas

- **Linguagem:** Go por decisão explícita; um port Rust é verificação futura
  possível, não parte do harness.
- **`go test -race ./...`** falha de propósito em
  `TestScenario1_DirectPointerMutationOnControlStruct` (`e1_registers`): a
  corrida de dados intencional é a evidência de que mutação não sincronizada
  corrompe. Rode os demais experimentos com `-race` sem problema.
- Validação: `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...`
  (estável em 20 execuções consecutivas do E1v2).

## Baseline Metodológico (v1.0.0-methodology-baseline)

> O núcleo metodológico é considerado congelado para fins de evolução
> arquitetural. Alterações futuras deverão ser motivadas por evidência
> empírica obtida durante a validação prática, e não por refinamentos
> especulativos.

**Bloqueio de expansão:** nenhuma nova camada de governança, protocolo,
tabela, invariante ou artefato conceitual será criada a partir deste
baseline. O ciclo especulativo/documental está encerrado.

**Composição do núcleo (congelado):** PDP (problema) · CVP (constraints) ·
OCS (contratos) · SCM (desempate por custo) + TR (governança e conexões), com
matrizes de responsabilidade, dependência e ownership e os cinco invariantes
do TR. O sistema é fechado operacionalmente e aberto epistemicamente: a
realidade do domínio (o axioma externo) pode sempre reabrir P0.

**Regra de ouro:** só abrir novos PDRs ou CVRs se (e somente se) a
re-execução dos experimentos e a coleta de evidências do mundo real
encontrarem falhas ou invalidações concretas.

**Status:** Fase de Especificação: **encerrada**. Fase de Validação Prática:
**iniciada**.
