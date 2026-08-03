# Propriedades Arquiteturais Observáveis + Property Oracle

> Este documento é o **contrato do harness**, congelado antes do primeiro
> harness ser escrito. Ele lista **apenas propriedades observáveis** — sem
> mencionar soluções (registradores, microkernel, `ExecutionPlan`,
> `WAIT_OUTCOME`, `R3`, `R5` são **candidatos a solução**, nunca pressupostos).
> Toda arquitetura candidata (A–E) deve satisfazer este conjunto; a
> evidência é a matriz propriedade × arquitetura.

## Regras congeladas (Decisões 1 e 2)

1. **Independência conceitual das arquiteturas concorrentes.** Cada
   arquitetura nasce de princípios próprios (Context Object + Policy
   Pipeline, Event Sourcing, Actor Model, Hierarchical State Machine). Não
   pode ser uma transformação sintática da arquitetura A, nem compartilhar
   estruturas internas de A por conveniência. *Se um rival acabar recriando
   R0–R7 disfarçados, isso também é um resultado* — sugere que os
   registradores são consequência inevitável do problema, não escolha
   arbitrária.
2. **Toda avaliação passa pelo Property Oracle.** Arquiteturas **não se
   auto-avaliam**: elas apenas emitem **eventos observáveis**. Somente o
   Oracle determina PASS / FAIL / UNKNOWN para cada propriedade, usando os
   critérios operacionais abaixo. Nenhuma arquitetura pode escrever "PASS".

## Vocabulário de eventos observáveis

Solução-livre: nenhum evento nomeia registrador, kernel ou plano. Cada
arquitetura emite estes fatos e nada mais.

| Evento | Significado observável |
|---|---|
| `AdmissionRequested` | Um pedido de admissão chega |
| `AdmissionAccepted` / `AdmissionRejected` | Veredito de admissão |
| `AuthorizationGranted` / `AuthorizationDenied` | Autorização concedida/negada para uma capacidade |
| `CapabilityStarted` / `CapabilitySucceeded` / `CapabilityFailed` | Ciclo de uma execução |
| `RollbackExecuted` | Estado parcial revertido |
| `ExternalEffect` | Efeito observável fora do sistema |
| `AuditWritten` | Registro de auditoria escrito |
| `TimedOut` | Expiração de tempo (watchdog) |

Eventos carregam a capacidade à qual se referem (vazia para eventos globais).

## Critérios operacionais (PASS / FAIL / UNKNOWN)

Cada propriedade tem critérios **pré-definidos**, para que arquiteturas
diferentes sejam julgadas pelo mesmo padrão, sem interpretação subjetiva.

| Prop | Critério operacional |
|---|---|
| **P1** Execução auditável | **PASS**: todo evento que altera estado é **seguido** de um `AuditWritten` que o cobre (auditoria pós-fato; auditoria antes de evento descoberto é ignorada). **FAIL**: existe evento de estado sem auditoria subsequente. **UNKNOWN**: nenhum evento de estado no cenário |
| **P2** Autorização antes do efeito | **PASS**: nenhum `ExternalEffect` ocorre antes do primeiro `AuthorizationGranted` da mesma capacidade. **FAIL**: existe `ExternalEffect` anterior à autorização. **UNKNOWN**: nenhum efeito externo no cenário |
| **P3** Falha bloqueia efeitos posteriores | **PASS**: após `CapabilityFailed`, nenhum `ExternalEffect` ocorre até uma nova autorização. **FAIL**: existe efeito após a falha sem nova autorização. **UNKNOWN**: nenhuma falha no cenário |
| **P4** Estado parcial recuperável | **PASS**: todo `CapabilityFailed` é seguido de `RollbackExecuted` antes do próximo `CapabilityStarted` daquela capacidade (e nenhuma falha fica pendente no fim do trace). **FAIL**: falha sem rollback. **UNKNOWN**: nenhuma falha no cenário |
| **P5** Tolerância assíncrona | **PASS**: nenhum `ExternalEffect` ocorre após `TimedOut` para a mesma capacidade (sucesso tardio é ignorado). **FAIL**: efeito após o timeout — critério mais estrito que o caso do sucesso tardio isolado. **UNKNOWN**: nenhum `TimedOut` no cenário |
| **P6** Independência de domínio | **PASS/FAIL** avaliado **estruturalmente** (não por eventos): trocar a política de domínio não altera o núcleo (execução dupla com políticas distintas sobre o mesmo núcleo) + inspeção. **UNKNOWN**: não testado |
| **P7** Admissão determinística | **PASS**: duas execuções da mesma entrada produzem o mesmo padrão `AdmissionAccepted/Rejected`. **FAIL**: divergência. **UNKNOWN**: menos de duas execuções com admissão |
| **P8** Isolamento entre execuções | **PASS** (critério v1, simplificado): eventos de cada capacidade são contíguos — nenhuma execução é observada dentro da janela de outra. **FAIL**: interleaving observado. **UNKNOWN**: sem concorrência no cenário. *Nota: UNKNOWN é deliberado e mais estrito que PASS-por-ausência* — o exemplo ilustrativo da crítica mostrava `P8 = PASS` no cenário de falha isolada; aqui, propriedade não exercitada não conta como passa |

**P9 — Simplicidade** é qualitativa (Alto/Médio/Baixo), usada só como
desempate, nunca como justificativa de necessidade.

## Oracle (implementação)

`harness/oracle.go` implementa exatamente estes critérios. Contrato:

```
Arquitetura → executa cenário → emite eventos → Oracle → P1 PASS / P2 FAIL / ...
```

`Oracle.Evaluate(trace)` gradeia uma execução; `Oracle.EvaluateRuns(runs)`
gradeia repetições (necessário para P7). O Oracle não conhece arquiteturas.

## Bateria de cenários (v1)

`harness/scenarios.go`. Cenários são **sondas de propriedade**; alguns são
adversariais por construção (a arquitetura bem-comportada deve emiti-los de
forma a passar, a mal-comportada falha):

| ID | Cenário | Propriedades exercitadas |
|---|---|---|
| S1 | Capacidade falha no meio | P1, P3, P4 |
| S2 | Efeito externo antes da autorização (adversarial) | P2 |
| S3 | Resultado tardio após timeout — ignorado (nominal) e aplicado (adversarial) | P5 |
| S4 | Duas capacidades intercaladas | P8 |
| S5 | Pedido de admissão idêntico, duas execuções | P7 |

## Regras de uso

1. Nenhuma menção a soluções neste arquivo — uma linha que nomeie
   "registrador", "R3", "WAIT_OUTCOME" etc. está mal escrita.
2. `?`/UNKNOWN na matriz não conta como passa: é motivo para **não** elevar
   o candidato.
3. Toda arquitetura roda os mesmos cenários; a matriz propriedade ×
   arquitetura é a evidência que alimenta os ADRs (necessidade demonstrada,
   não escolha).
