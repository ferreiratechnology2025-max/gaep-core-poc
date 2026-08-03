# adr-evidence

Evidência do **GAEP Core Architecture Validation Harness** (ex-"PoC"). O
programa é **refutation-first**: experimentos tentam mostrar que a arquitetura
de registradores *não* é necessária, e a evidência é reportada com honestidade
mesmo quando enfraquece o caso de $R0$–$R7$.

## Contrato e método

- `problem-definition.md` — **nível mais alto**: Problem Definition (P0,
  **versionado** — hipótese raiz, não axioma) + **Candidate→Validated
  Constraints** (hipóteses com teste de falsificação), lint de higiene
  (Needs como resultado puro), a distinção **mecanismo vs contrato
  observável** e as duas perguntas obrigatórias por decisão.
- `problem-definition-protocol.md` — **quem valida o próprio problema**: ciclo
  de vida de P0, classificação de mudanças (editorial / semântica /
  redefinição), cascata de invalidação (P0 muda → constraints voltam a
  Proposed), PDR rastreável, **DAG de proveniência com propagação computável**
  e **versionamento semântico da cadeia inteira** (`decided against P0
  vX.Y.Z`). Ponto de parada: **Trilema de Münchhausen encerrado por axioma**
  (a realidade do domínio é o axioma externo ao sistema), não eliminado — o
  protocolo valida a representação, não a realidade. Sem ele, o problema
  seria o novo axioma.
- `traceability-registry.md` — **índice mestre e governança**: PDR → P0
  vX.Y.Z → CVRs → Needs → Properties → Contracts → Experiments → Evidence →
  ADR, com cinco colunas por elemento (origem · protocolo · evidência ·
  versão · auditoria) e **cinco invariantes formais** (unicidade ativa,
  imutabilidade referenciada, rastreabilidade a P0, evidência mínima,
  aciclicidade do DAG) — papel de CMDB, parte da especificação normativa.
  Não é um protocolo — registra conexões.

- `architectural-necessity-matrix.md` — **fonte da verdade**: a cadeia
  Constraint→Need→Property→Mechanisms→Experiments→Evidence→Decision e a
  regra dos três resultados (nenhum mecanismo → remover; vários →
  ADR/Protocol; apenas um → congelar na architecture).
- `constraint-validation-protocol.md` — **quem valida as constraints e com
  base em quê**: ciclo de vida Proposed→Validated/Reformulated/Rejected, as
  seis perguntas operacionais, registro CVR rastreável e separação de papéis
  (proponente ≠ validador; domínio + harness). Sem este protocolo, o viés
  apenas sobe um nível — dos mecanismos para as constraints.
- `observable-contract-specification.md` — **o que permanece arquitetural com
  mecanismos equivalentes**: critério do contrato observável + inventário v1
  (vocabulário de eventos + P1–P8). O harness é verificador de contratos.
- `structural-cost-matrix.md` — **desempate** entre arquiteturas que
  preservam as mesmas propriedades (dimensões, formato, regras, limitações;
  vazia até o E1).
- `architectural-properties.md` — **propriedades observáveis (P1–P8)**, sem
  mencionar soluções, com **critérios operacionais PASS/FAIL/UNKNOWN** por
  propriedade e o vocabulário de eventos observáveis. Regras congeladas:
  (1) arquiteturas concorrentes **conceitualmente independentes**; (2)
  avaliação **somente via Property Oracle**.
- `../harness/` — implementação do Oracle + bateria de cenários S1–S5.
- `falsification-matrix.md` — conceitos candidatos × como falsificá-los ×
  qual propriedade deve quebrar × status.
- `experiment-plan.md` — ordem E0–E5 (propriedades → arquiteturas concorrentes
  → cenários), com mapeamento para os artefatos existentes.

## Experimentos

| ID | Pergunta | Onde | Log |
|---|---|---|---|
| E0 | Remoção de componentes: qual elemento é arquiteturalmente necessário? | `experiments/e0_minimality` | `E0_evidence.log` |
| E1 | Sincronização: mutação não sincronizada corrompe? (demo) | `experiments/e1_registers` | `E1_evidence.log` |
| E3 | Registradores: invariantes ou representação? (A/B/B′/C sob sincronização igual) | `experiments/e1_models` | `E1v2_evidence.log` |

## Achados atuais (resumo honesto)

- **E1v2**: com sincronização igual, o Modelo B′ (estado fechado tipado, sem
  registradores) preserva todas as propriedades medidas do Modelo C
  (registradores) → **registradores não provados necessários**; a superfície
  fechada e guardada é o que importa.
- **E0**: R1/R2/R6/R7 pinados por invariantes; **R0/R3/R4/R5 com lacuna de
  requisito** (especificar a propriedade ou remover). SESSION/WATCHDOG/
  WAIT_OUTCOME/RELEASE são load-bearing no ciclo de vida.
- Próximo: **E1 arquiteturas concorrentes** (A GAEP, B sem registradores,
  C event sourcing, D actor, E state machine) × P1–P8 — ver
  `experiment-plan.md`.

## Notas

- **`go test -race ./...`** falha de propósito em
  `TestScenario1_DirectPointerMutationOnControlStruct` (`e1_registers`): a
  corrida de dados intencional é a evidência de que mutação não sincronizada
  corrompe. Rode os demais experimentos com `-race` sem problema.
- **Determinismo:** E0 e E1v2 são determinísticos por construção (barreira no
  E1, amostragem cronológica e invariante de timestamp por escrita no E1v2).
  Verificado estável em 20 execuções consecutivas do E1v2.
- **Linguagem:** Go por decisão explícita; port Rust é verificação futura
  possível, não parte do harness.

## Baseline Metodológico (v1.0.0-methodology-baseline)

> O núcleo metodológico é considerado congelado para fins de evolução
> arquitetural. Alterações futuras deverão ser motivadas por evidência
> empírica obtida durante a validação prática, e não por refinamentos
> especulativos.

**Bloqueio de expansão:** nenhuma nova camada de governança, protocolo,
tabela, invariante ou artefato conceitual será criada a partir deste
baseline. O ciclo especulativo/documental está encerrado.

**Regra de ouro:** só abrir novos PDRs ou CVRs se (e somente se) a
re-execução dos experimentos e a coleta de evidências do mundo real
encontrarem falhas ou invalidações concretas.

**Status:** Fase de Especificação: **encerrada**. Fase de Validação Prática:
**iniciada**.
