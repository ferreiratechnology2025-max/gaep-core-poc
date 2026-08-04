# E1 × C — Achados (S4 × P8)

Observed Results
  S4 × P8 = FAIL (veredito do Oracle).

Hypothesized Cause
  Ausência de isolamento entre capacidades em Event Sourcing
  (single projection + shared execution order).

Evidence Required
  Trace de S4 mostrando a sequência exata de eventos que viola P8.

Status
  HIPÓTESE. Não comprovada pelo Oracle. Requer análise do trace.
