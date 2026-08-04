# E1 × A — Achados (Fase 5B)

## Parte 0 — Verificação de concorrência de D (S4×P8)

Objetivo: determinar se o PASS de D em S4×P8 (500 execuções, 100% PASS,
Fases 3–5) reflete concorrência genuína entre atores ou serialização do
ambiente. Implementado em
`experiments/e1_architectures/d_concurrency_test.go`
(`TestE1_ArchitectureD_ConcurrencyCheck`) — não altera D nem o pré-registro
estocástico original; apenas reexecuta o mesmo script S4 sob dois regimes de
paralelismo.

| regime | N | PASS | FAIL | UNKNOWN | freq PASS |
|---|---|---|---|---|---|
| GOMAXPROCS=1 | 50 | 50 | 0 | 0 | 100% |
| GOMAXPROCS=4 | 50 | 49 | 1 | 0 | 98% |

Distribuição mudou entre os regimes: **sim** (1 FAIL surgiu em
GOMAXPROCS=4). Resultado estável em 5 repetições consecutivas do teste
(mesma contagem 50/0 e 49/1 em todas).

### Bloco Observed / Hypothesized Cause / Evidence Required / Status

```
Observed Results
  GOMAXPROCS=1 → S4×P8 PASS=50/50, FAIL=0/50 (massa pontual).
  GOMAXPROCS=4 → S4×P8 PASS=49/50, FAIL=1/50 (variância presente).
  Resultado reproduzido de forma idêntica em 5 execuções consecutivas do
  teste nesta máquina.

Hypothesized Cause
  A variância surgida em GOMAXPROCS=4 é consistente com concorrência
  genuína entre os atores de c1 e c2: com mais de um processador
  disponível, o escalonador do runtime Go tem oportunidade de intercalar a
  entrega das caixas postais de forma que nem sempre produz uma trajetória
  contígua por capacidade. Em GOMAXPROCS=1, a ausência de paralelismo real
  (um único thread de SO executando as goroutines) parece reduzir — mas,
  por este teste, não elimina _a priori_ — as janelas em que tal
  intercalação seria observável. A causa exata (agendamento em nível de
  runtime vs. características deste script/hardware específico) não é
  estabelecida por este teste.

Evidence Required
  (a) Repetição em outros ambientes/escalonadores (Linux, GOMAXPROCS maior
  que NumCPU, outra máquina); (b) uma sonda independente que force ou
  detecte intercalação em nível de ator (ex.: atraso instrumentado no
  handler) para confirmar que o escalonador consegue intercalar as caixas
  postais de c1/c2 neste ambiente fora deste script específico; (c) esta
  mesma verificação repetida após a Fase 5B Parte 1 (bateria de A) para
  contextualizar o quinto ponto de dados.

Status
  HIPÓTESE. Este teste demonstra que a variância NÃO é zero sob
  paralelismo — ou seja, o PASS de 500/500 execuções (Fases 3–5, regime de
  paralelismo não fixado) não pode mais ser tratado como determinístico por
  construção do paradigma; ele é, no mínimo, sensível ao regime de
  paralelismo. Isso NÃO estabelece qual dos dois é o mecanismo "real" nem
  resolve a divergência S4×P8 (D=PASS modal vs. B/C/E=FAIL) registrada em
  `E1_D_findings.md` — apenas eleva a prioridade das evidências pendentes
  ali listadas.
```

## Parte 1 — Bateria de A (dirigida por `kernel/`) — BLOQUEADA em PASSO 0

### PASSO 0 — Mapeamento comando → registrador → evento

Fontes lidas (somente leitura, nenhuma modificada): `kernel/registers.go`
(register file $R0$–$R7$, thread-safe, sem noção de identidade de
capacidade) e `models_comp/lifecycle.go` (autômato de 4 estados
IDLE/RUNNING/WAIT_OUTCOME/RELEASE, 6 eventos de controle) e
`models_comp/policy.go` (`GovernancePolicy`, guarda de escrita por célula:
domínio de $R6$ Health, monotonicidade de $R2$ NextAct, futuro de $R1$
Watchdog — não é uma política de admissão por identidade).

O princípio de origem congelado de A é "Registradores $R0$–$R7$ +
microkernel + lifecycle" (`experiment-plan.md`, linha 64). Três pontos do
protocolo S1–S5 (idêntico ao usado por B/C/D/E) não decorrem desse
princípio e não foram resolvidos em nenhuma decisão de Arquiteto já
registrada (busca em `experiment-plan.md` e `traceability-registry.md` não
encontrou uma seção "A" equivalente às seções B/C/D/E que resolvem
"DECISÃO DO ARQUITETO NECESSÁRIA"). Reportando em vez de inventar, por
restrição explícita da fase:

1. **Identidade de capacidade.** `kernel.RegisterSet` é UM único arquivo de
   8 registradores, sem indexação por capacidade. $R0$ Session guarda um
   único `uint64`. S1–S3 e S5 usam uma capacidade (c1); S4 usa duas (c1,
   c2) com trajetórias que devem poder ser observadas como intercaladas ou
   contíguas (P8). Duas leituras possíveis do princípio congelado, ambas
   compatíveis com "registradores + microkernel", produzem experimentos
   diferentes:
   - (a) um `RegisterSet` por capacidade → P8 passa a ser garantido por
     construção (memória disjunta), o que testaria isolamento por
     alocação, não pelo mecanismo de registradores em si;
   - (b) um único `RegisterSet` compartilhado, com no máximo uma
     capacidade "ativa" por vez nele representável → força serialização
     das duas capacidades de S4, o que testa uma pergunta diferente da que
     B/C/D/E testam (elas mantêm os dois contextos de capacidade
     coexistentes).
   Nenhuma das duas é "a" arquitetura A sem uma escolha do Arquiteto.

2. **Política de admissão por identidade.** B/C/D/E recebem
   `policy func(cap string) bool` e o adaptador decide aceitar/rejeitar a
   admissão por capacidade. `GovernancePolicy` não tem esse formato: ela
   valida (célula, valor antigo, valor novo), não (identidade da
   capacidade). Não há, em `kernel/` ou `models_comp/`, uma superfície de
   decisão de admissão por capacidade já definida — sintetizar uma
   (ex.: "aceitar sse a escrita em $R0$ Session for não-zero") seria uma
   regra inventada pelo adaptador, não uma regra que já governa A.

3. **Rollback vs. estado absorvente.** `Lifecycle.Step` documenta RELEASE
   como absorvente: `EvRelease`/`EvFail` a partir de RELEASE retornam erro
   ("RELEASE is absorbing"); não há evento de saída de RELEASE. O
   protocolo S1 exige Fail → Rollback → capacidade novamente utilizável
   (P4: estado parcial recuperável, não terminação permanente). Em B/C/D/E
   rollback SEMPRE retorna a um estado operável. Representar
   `RollbackExecuted` depois de entrar em RELEASE exigiria descartar a
   instância do autômato e criar uma nova — o que não é "rollback" nos
   termos do próprio autômato de A (é abandono do autômato, não uma
   transição dele).

Pontos 1–3 são decisões arquiteturais não cobertas pela spec congelada.
Per a instrução da fase ("Se o mapeamento for ambíguo ... PARE e reporte a
decisão específica necessária. NÃO invente mapeamento"), a Parte 1 para
aqui: nenhum adaptador foi implementado, nenhum arquivo de `kernel/` ou
`models_comp/` foi tocado, e nenhuma matriz P1–P8 × S1–S5 foi produzida
para A nesta rodada.

### Eventos que A não consegue produzir sem decisão prévia

Consequência dos pontos 1–3: `AdmissionRequested/Accepted/Rejected` (ponto
2), a distinção `AuthorizationGranted` de um simples `EvSession` do
lifecycle (o autômato de 4 estados não tem um passo de autorização
separado de sessão), e `RollbackExecuted` como retorno a estado operável
(ponto 3) ficam todos indeterminados até a decisão do Arquiteto. Isso
deixaria, na pior leitura, P1/P2/P3/P4/P7/P8 todos como UNKNOWN por falta
de eventos — não porque A não os satisfaça, mas porque o mapeamento não
foi definido.

## Validação (Parte 0, componente implementado)

`gofmt -l .` limpo · `go vet ./...` limpo · `go build ./...` OK ·
`go test ./...` OK (todos os pacotes, incluindo o teste novo). `-race`:
mesma limitação das fases anteriores neste ambiente
(`CGO_ENABLED=1` ausente no toolchain local — `go: -race requires cgo`);
não executado.
