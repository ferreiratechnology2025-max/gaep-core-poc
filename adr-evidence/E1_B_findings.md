# E1 × B — Achados (S4 × P8)

## Matriz P1–P8 × S1–S5 (verditos do Property Oracle)

| property | S1 | S2 | S3 | S4 | S5 |
|---|---|---|---|---|---|
| P1 Auditabilidade | PASS | PASS | PASS | PASS | PASS |
| P2 Autorização antes do efeito | UNKNOWN | PASS | UNKNOWN | UNKNOWN | UNKNOWN |
| P3 Falha bloqueia efeitos | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P4 Estado parcial recuperável | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P5 Tolerância a execução assíncrona | UNKNOWN | UNKNOWN | PASS | UNKNOWN | UNKNOWN |
| P6 Kernel independente do domínio | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P7 Admissão determinística | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | PASS |
| P8 Isolamento de execuções | UNKNOWN | UNKNOWN | UNKNOWN | FAIL | UNKNOWN |

UNKNOWN = o cenário não exercitou o gatilho da propriedade no trace de B.
P6 é estrutural (dual-run + inspeção de fonte), não baseada em eventos.

## Blocos Observed / Hypothesized Cause / Evidence Required / Status

### B × S4 × P8

```
Observed Results
  S4 × P8 = FAIL (veredito do Oracle).

Hypothesized Cause
  Ausência de isolamento entre capacidades em Context Object + Policy
  Pipeline (contextos imutáveis por capacidade, mas ordem de travessia
  única e fixa — o interleaving de comandos permanece observável).

Evidence Required
  Trace de S4 mostrando a sequência exata de eventos que viola P8.

Status
  HIPÓTESE. Não comprovada pelo Oracle. Requer análise do trace.
```

## Nota de convergência

B (Context Object + Policy Pipeline) produz a MESMA matriz observável que
C (Event Sourcing) e E (Hierarchical State Machine) — incluindo o FAIL de
P8 em S4. Três mecanismos estruturalmente distintos (log append-only +
projeção; máquinas por capacidade; pipeline de estágios puros sobre
objetos imutáveis) convergem no mesmo veredicto por cenário. Isso pesa
contra o poder discriminativo da bateria (H2), mas continua sendo
evidência a favor de H1 (o contrato observável abstrai o mecanismo).
Ambas permanecem HIPÓTESES.
