# E2.1a — Desenho da expansão de cobertura da bateria

Fase de DESENHO apenas. Nenhum código foi criado ou alterado. Nenhum
arquivo do E1 (`harness/oracle.go`, `harness/scenarios.go`,
`harness/events.go`, S1–S5, relatório/findings) foi tocado — leitura
apenas, incluindo `experiments/e1_architectures/*.go` (B/C/D/E) para
verificar, por inspeção de fonte, o comportamento já implementado por cada
arquitetura nas trilhas de guarda que os cenários novos vão exercitar.

## 0. Correção de enquadramento em relação à E2A

A E2A classificou S1×P3 e S3×P5 como "PASS vazio" e sugeriu, na coluna
"mudança mínima", simplesmente "adicionar uma tentativa de Effect" como se
isso fosse suficiente para transformar o PASS em um teste substantivo que
poderia, em princípio, também dar FAIL. Isso precisa de uma ressalva antes
do desenho:

Inspecionei o guard de efeito nas quatro arquiteturas:

- B (`b_contextpipeline_test.go`): `allowed := c.authorized && !c.failed && !c.timedOut`.
- C (`c_eventsourcing_test.go::effectAllowed`): permite efeito só em estado
  `Authorized`/`Running`, bloqueado se `timedOut[cap]` — `Failed` não está
  entre os estados permitidos.
- D (`d_actormodel_test.go::handle`, caso `dCmdEffect`): `a.authorized && !a.failed && !a.timedOut`.
- E (`e_statemachine_test.go`, comentário de linha 9-10): "effect requires
  an authorized, non-failed, non-timed-out position".

As quatro já implementam o mesmo guard (autorizado E não-falho E não-
timed-out) antes de permitir `ExternalEffect`. Uma tentativa de efeito
pós-falha ou pós-timeout, portanto, **não vai produzir `ExternalEffect`
em nenhuma das quatro** — vai produzir `AuthorizationDenied` (o efeito
nunca se torna observável, porque foi bloqueado no gate). Isso significa
que `oracle.go:p3`/`p5` continuam retornando PASS depois do cenário novo,
pela mesma razão estrutural: a propriedade só FALHARIA se uma arquitetura
*não* checasse `failed`/`timedOut` no gate — o que nenhuma das quatro
testadas faz.

**O que muda, então, não é o veredito — é o que o PASS significa.** Hoje
(S1×P3, S3×P5), o PASS é vazio porque **nenhuma tentativa de efeito
ilegal jamais foi emitida pelo script de comando** — o risco nunca foi
exercitado. Com os cenários novos, o script de comando **emite a
tentativa**, e o PASS passa a significar "a tentativa foi feita e
bloqueada", que é uma alegação bem mais forte. Continua sendo hipótese que
o bloqueio "de fato" funcionaria numa quinta arquitetura ainda não testada
— mas deixa de ser hipótese que *nenhuma arquitetura testada* foi
verificada contra esse caso. Reporto isso agora, antes do desenho, porque
senão a mudança de veredito esperada (nenhuma) poderia ser lida como o
desenho "não ter funcionado".

Um efeito colateral interessante, não garantido: se uma quinta
arquitetura (ou uma variação futura de B/C/D/E) *não* checar `failed`
antes de permitir o efeito, os cenários novos abaixo são o primeiro ponto
da bateria capaz de capturar isso como FAIL real. Isso é uma vantagem do
desenho, não uma previsão de resultado.

## 1. Tabela de cobertura (tipos de lacuna → cenário novo)

| tipo de lacuna | células E1 afetadas (categoria E2A) | cenário novo |
|---|---|---|
| adicionar `ExternalEffect` | P2×{S1,S3,S4,S5} (CENÁRIO) | S6 (pós-falha), S7 (pós-timeout) — já há `ExternalEffect` legítimo em S2 |
| adicionar `CapabilityFailed` | P3×{S2,S3,S4,S5}, P4×{S2,S3,S4,S5} (CENÁRIO) | S6 |
| adicionar `TimedOut` | P5×{S1,S2,S4,S5} (CENÁRIO) | S7 |
| adicionar `Admit` (com ≥2 runs) | P7×S4 (CENÁRIO — S4 não tem Admit algum) | S8 |
| adicionar segunda capacidade | P8×{S1,S2,S3,S5} (CENÁRIO) | S6 (sequencial), S8 (intercalada) |
| tentativa de efeito após falha | S1×P3 (ORACLE — PASS vazio, ver §0) | S6 |
| tentativa de efeito após timeout | S3×P5 (ORACLE — PASS vazio, ver §0) | S7 |

Três cenários (S6, S7, S8) cobrem os sete tipos de lacuna. Nenhum cenário
existente (S1–S5) é alterado.

## 2. Especificação dos cenários novos

### S6 — "falha, tentativa ilegal de efeito, rollback, segunda capacidade admitida e concluída"

**Intenção:** fechar `CapabilityFailed` ausente em S2/S3/S4/S5, a
tentativa de efeito pós-falha ausente em toda a bateria (§0), e uma
segunda capacidade em posição *sequencial* (não intercalada) — variante
de P8 distinta de S4.

**Sequência de comandos (idêntica para B, C, D, E):**

1. Admit c1
2. Authorize c1
3. Start c1
4. Fail c1
5. Effect c1 — tentativa ilegal, pós-falha, pré-rollback
6. Rollback c1
7. Admit c2
8. Authorize c2
9. Start c2
10. Succeed c2

**Trajetória nominal esperada** (com base no guard já implementado por
B/C/D/E, §0 — comportamento esperado, a confirmar na E2.1b):

```
AdmissionRequested()        AuditWritten()
AdmissionAccepted()         AuditWritten()
AuthorizationGranted(c1)    AuditWritten()
CapabilityStarted(c1)       AuditWritten()
CapabilityFailed(c1)        AuditWritten()
AuthorizationDenied(c1)     AuditWritten()   ← efeito bloqueado (c1 failed)
RollbackExecuted(c1)        AuditWritten()
AdmissionRequested()        AuditWritten()
AdmissionAccepted()         AuditWritten()
AuthorizationGranted(c2)    AuditWritten()
CapabilityStarted(c2)       AuditWritten()
CapabilitySucceeded(c2)     AuditWritten()
```

**Variante adversarial:** o passo 5 (`Effect c1`) força `ExternalEffect`
diretamente na trajetória em vez de invocar o gate da arquitetura (i.e.,
uma trajetória sintética que injeta a violação, como já se faz para
S2/S3 em `harness/scenarios.go`). Serve só para confirmar que o Oracle
detecta a violação quando ela de fato ocorre — não é rodada contra B/C/D/E
(mesma disciplina do E1: `Adversarial` testa o Oracle isoladamente).

**Células que passa a exercitar (deixam de ser CENÁRIO/PASS-vazio):**
S6×P1, S6×P3 (agora substantivo, §0), S6×P4 (rollback + readmissão),
S6×P7 (contribui admissão), S6×P8 (segunda capacidade, sequencial — valor
esperado PASS por construção, já que não há intercalação; ponto de
comparação com S4/S8).

### S7 — "timeout, tentativa ilegal de efeito, sucesso tardio ignorado"

**Intenção:** fechar `TimedOut` ausente em S1/S2/S4/S5 e a tentativa de
efeito pós-timeout ausente em toda a bateria (§0 — fecha o PASS vazio de
S3×P5).

**Sequência de comandos:**

1. Admit c1
2. Authorize c1
3. Start c1
4. Timeout c1
5. Effect c1 — tentativa ilegal, pós-timeout
6. Succeed c1 — sucesso tardio, ainda ignorado (mesma semântica de S3)

**Trajetória nominal esperada:**

```
AdmissionRequested()        AuditWritten()
AdmissionAccepted()         AuditWritten()
AuthorizationGranted(c1)    AuditWritten()
CapabilityStarted(c1)       AuditWritten()
TimedOut(c1)                AuditWritten()
AuthorizationDenied(c1)     AuditWritten()   ← efeito bloqueado (c1 timedOut)
CapabilitySucceeded(c1)     AuditWritten()   ← tardio, sem efeito
```

**Variante adversarial:** análoga a `ScenarioLateSuccessIgnored.Adversarial`
já existente em `harness/scenarios.go` — injeta `ExternalEffect` após o
`TimedOut` diretamente na trajetória sintética, só para confirmar
detecção pelo Oracle isolado.

**Células que passa a exercitar:** S7×P1, S7×P5 (agora substantivo, §0).

### S8 — "duas capacidades admitidas e intercaladas"

**Intenção:** fechar `Admit` ausente em S4 (P7×S4 hoje é CENÁRIO puro,
sem nenhum evento de admissão na trajetória) sem alterar S4, e produzir um
segundo ponto de dados de P8 sob intercalação genuína — dessa vez com
admissão explícita nas duas capacidades, o que S4 nunca teve.

**Sequência de comandos** (mesma estrutura de intercalação de S4, com
Admit prefixado em ambas):

1. Admit c1
2. Admit c2
3. Authorize c1
4. Start c1
5. Start c2 — intercalação dentro da janela de c1
6. Succeed c2
7. Succeed c1

**Trajetória nominal esperada:**

```
AdmissionRequested()        AuditWritten()
AdmissionAccepted()         AuditWritten()
AdmissionRequested()        AuditWritten()
AdmissionAccepted()         AuditWritten()
AuthorizationGranted(c1)    AuditWritten()
CapabilityStarted(c1)       AuditWritten()
CapabilityStarted(c2)       AuditWritten()
CapabilitySucceeded(c2)     AuditWritten()
CapabilitySucceeded(c1)     AuditWritten()
```

Sem variante adversarial (S8 é estrutural, como S4 — não há violação a
injetar; a intercalação em si é o que o Oracle avalia via P8).

**Células que passa a exercitar:** S8×P1, S8×P7 (agora com gatilho real,
sujeito ao protocolo do §3), S8×P8 (segundo ponto de dados sob
intercalação — comparável a S4×P8, inclusive para D sob a cláusula
estocástica, se aplicável na E2.1b).

## 3. Protocolo P7 — mudança de invocação, não de critério

`harness/oracle.go:p7` já aceita N≥2 runs corretamente
(`EvaluateRuns`); o gap é que a bateria viva (`refBCEMatrix` em
`d_test.go`, e o equivalente em `b_test.go`/`c_test.go`/`e_test.go`) chama
`o.Evaluate(sc.run(s.ID))` — um único run — para S1–S3, e só usa dois runs
(`o.EvaluateRuns([][]harness.Event{r1, r2})`) para S5. Isso é uma escolha
de teste, não do Oracle.

**Mudança proposta (só na invocação, na E2.1b):** para todo cenário cujo
script contenha ao menos um comando `Admit` (S1, S2, S3, S5 hoje; S6, S8
novos), avaliar P7 com dois runs frescos (`EvaluateRuns` sobre `r1, r2`),
igual ao que já se faz para S5. Cenários sem `Admit` (S4, S7) continuam
UNKNOWN em P7 por ausência real de gatilho — isso é CENÁRIO, não muda.

Isso fecha as 3 células ORACLE de P7×{S1,S2,S3} identificadas na E2A sem
tocar `oracle.go`.

## 4. Nota "PASS vazio" — como os cenários fecham S1×P3 e S3×P5

Ver §0 para o enquadramento correto: **o veredito PASS não muda** (nenhuma
das quatro arquiteturas testadas permite o efeito ilegal). O que muda é o
que o PASS demonstra:

- **Antes (S1×P3, S3×P5):** PASS por ausência total de tentativa — o
  script de comando nunca emite `Effect` após `Fail`/`Timeout`. O Oracle
  não tinha nada para potencialmente reprovar.
- **Depois (S6×P3, S7×P5):** PASS por bloqueio confirmado — o script
  emite a tentativa (`Effect` pós-falha / pós-timeout) e a arquitetura a
  transforma em `AuthorizationDenied`, nunca em `ExternalEffect`. O Oracle
  teve algo real para potencialmente reprovar e não reprovou.

A distinção é observável na própria trajetória: S1/S3 nunca contêm o
comando `Effect` depois do ponto de falha/timeout; S6/S7 contêm, e o
evento resultante (`AuthorizationDenied`, não `ExternalEffect`) é o que
sustenta o PASS substantivo.

## 5. Itens deliberadamente fora do escopo mínimo (candidatos a S9+)

- **`Start` durante `pending` (sub-cláusula de P4 não testada mesmo em
  S1×P4, notada na E2A).** Inspeção de `b_contextpipeline_test.go`
  (`runStage`/`bStageExecution`) mostra que B **não guarda** contra um
  `Start` reemitido enquanto a capacidade está em `failed` — não há
  checagem de `c.failed` nesse estágio. Diferente do gate de `Effect`
  (§0), aqui o comportamento das quatro arquiteturas **não foi
  verificado como uniforme** por esta leitura, e um cenário que force
  essa sequência poderia genuinamente divergir (FAIL em algumas,
  PASS em outras) — o que seria descoberta real, não desenho garantido.
  Não incluí no conjunto mínimo porque o resultado esperado não pode ser
  declarado com a mesma confiança de S6/S7 (§0), e a fase de desenho não
  deve afirmar comportamento não verificado. Proponho como candidato a
  S9, com leitura de código completa de C e E antes de especificar a
  trajetória nominal esperada.
- **Efeito legítimo pós-rollback (capacidade re-autorizada depois de
  recuperada).** Não incluído porque nenhuma célula E1 depende disso;
  registrado apenas como possível extensão de S6 se o Arquiteto quiser
  testar P2/P3 juntos após uma recuperação completa.

## 6. Resumo

| cenário | fecha | risco/incerteza |
|---|---|---|
| S6 | ExternalEffect, CapabilityFailed, 2ª capacidade (sequencial), efeito pós-falha | nenhum — guard já uniforme em B/C/D/E (§0) |
| S7 | TimedOut, efeito pós-timeout | nenhum — guard já uniforme em B/C/D/E (§0) |
| S8 | Admit em cenário intercalado, 2º ponto de P8 | nenhum — reusa estrutura de S4, só adiciona Admit |
| protocolo P7 | UNKNOWN de P7×{S1,S2,S3} (e S6, S8 quando implementados) | mudança de invocação de teste, não de critério |

S1–S5, `harness/oracle.go`, `harness/events.go` e todos os artefatos do
E1 permanecem intocados. Aguardando aprovação do Arquiteto antes da
E2.1b (implementação).
