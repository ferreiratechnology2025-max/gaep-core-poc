# E1 × D — Achados (cláusula estocástica, N=50)

## Pré-registro (congelado pelo Arquiteto)

- N = 50 execuções por cenário (cada execução = instância fresca de D).
- Veredito representativo por célula = **veredito modal** sobre as N execuções.
- Frequência modal < 50% ⇒ célula marcada **"alta variância"**.
- Cenários de capacidade única (S1/S2/S3/S5) devem ser **massa pontual**
  (um ator ⇒ ordem interna preservada). Se não forem, é bug de
  implementação, não propriedade do paradigma.
- D é **não determinístico** por construção (concorrência genuína de
  atores; entrega entre atores governada pelo escalonador do runtime).

## Matriz distribucional P1–P8 × S1–S5 (D, N=50)

| cell | PASS | FAIL | UNKNOWN | modal | freq |
|---|---|---|---|---|---|
| S1 × P1 | 50 | 0 | 0 | PASS | 100% |
| S1 × P2 | 0 | 0 | 50 | UNKNOWN | 100% |
| S1 × P3 | 50 | 0 | 0 | PASS | 100% |
| S1 × P4 | 50 | 0 | 0 | PASS | 100% |
| S1 × P5 | 0 | 0 | 50 | UNKNOWN | 100% |
| S1 × P6 | 0 | 0 | 50 | UNKNOWN | 100% |
| S1 × P7 | — | — | — | PASS (conjunto N) | N/N |
| S1 × P8 | 0 | 0 | 50 | UNKNOWN | 100% |
| S2 × P1 | 50 | 0 | 0 | PASS | 100% |
| S2 × P2 | 50 | 0 | 0 | PASS | 100% |
| S2 × P3 | 0 | 0 | 50 | UNKNOWN | 100% |
| S2 × P4 | 0 | 0 | 50 | UNKNOWN | 100% |
| S2 × P5 | 0 | 0 | 50 | UNKNOWN | 100% |
| S2 × P6 | 0 | 0 | 50 | UNKNOWN | 100% |
| S2 × P7 | — | — | — | PASS (conjunto N) | N/N |
| S2 × P8 | 0 | 0 | 50 | UNKNOWN | 100% |
| S3 × P1 | 50 | 0 | 0 | PASS | 100% |
| S3 × P2 | 0 | 0 | 50 | UNKNOWN | 100% |
| S3 × P3 | 0 | 0 | 50 | UNKNOWN | 100% |
| S3 × P4 | 0 | 0 | 50 | UNKNOWN | 100% |
| S3 × P5 | 50 | 0 | 0 | PASS | 100% |
| S3 × P6 | 0 | 0 | 50 | UNKNOWN | 100% |
| S3 × P7 | — | — | — | PASS (conjunto N) | N/N |
| S3 × P8 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P1 | 50 | 0 | 0 | PASS | 100% |
| S4 × P2 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P3 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P4 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P5 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P6 | 0 | 0 | 50 | UNKNOWN | 100% |
| S4 × P7 | — | — | — | UNKNOWN (conjunto N) | sem admissões |
| S4 × P8 | **50** | 0 | 0 | **PASS** | **100%** |
| S5 × P1 | 50 | 0 | 0 | PASS | 100% |
| S5 × P2 | 0 | 0 | 50 | UNKNOWN | 100% |
| S5 × P3 | 0 | 0 | 50 | UNKNOWN | 100% |
| S5 × P4 | 0 | 0 | 50 | UNKNOWN | 100% |
| S5 × P5 | 0 | 0 | 50 | UNKNOWN | 100% |
| S5 × P6 | 0 | 0 | 50 | UNKNOWN | 100% |
| S5 × P7 | — | — | — | PASS (conjunto N) | N/N |
| S5 × P8 | 0 | 0 | 50 | UNKNOWN | 100% |

P7: avaliado **em conjunto** sobre as N execuções (protocolo do Oracle para
determinismo); a distribuição individual é UNKNOWN por construção.
Células de **alta variância**: **nenhuma** (todas as células com veredito
não-UNKNOWN têm frequência modal 100%).

## Tabela de convergência D × (B/C/E) — células informativas

| cell | D | B/C/E | converge |
|---|---|---|---|
| S1 × P1 | PASS | PASS | SIM |
| S1 × P3 | PASS | PASS | SIM |
| S1 × P4 | PASS | PASS | SIM |
| S2 × P1 | PASS | PASS | SIM |
| S2 × P2 | PASS | PASS | SIM |
| S3 × P1 | PASS | PASS | SIM |
| S3 × P5 | PASS | PASS | SIM |
| S4 × P1 | PASS | PASS | SIM |
| **S4 × P8** | **PASS** | **FAIL** | **NÃO** |
| S5 × P1 | PASS | PASS | SIM |
| S5 × P7 | PASS | PASS | SIM |

**Convergência: 10/11 células informativas.** A única divergência é
**S4 × P8** (D = PASS modal vs B/C/E = FAIL), estável em 10 execuções do
teste (500 execuções de D no total, todas PASS=50/50).

## Lacunas de observabilidade (assimetria de protocolo, não divergência de mecanismo)

| cell | D | B/C/E |
|---|---|---|
| S1 × P7 | PASS (conjunto N=50) | UNKNOWN (avaliado com run único em S1–S3) |
| S2 × P7 | PASS (conjunto N=50) | UNKNOWN (idem) |
| S3 × P7 | PASS (conjunto N=50) | UNKNOWN (idem) |

P7 é massa pontual em D nos cenários de capacidade única (todas as N
execuções produzem o mesmo padrão de admissão). A assimetria decorre do
protocolo de avaliação (D tem N=50 runs; os testes de B/C/E avaliam P7 com
run único em S1–S3), não de diferença de mecanismo.

## Bloco Observed / Hypothesized Cause / Evidence Required / Status

### Divergência D × (B/C/E) em S4 × P8

```
Observed Results
  S4 × P8: veredito modal de D = PASS (50/50 execuções, frequência 100%),
  estável em 10 execuções do teste (500 execuções de D). B/C/E = FAIL na
  mesma célula (Fases 3–5). Primeira divergência de toda a bateria.

Hypothesized Cause
  No Actor Model, cada capacidade tem sua própria caixa postal processada
  por um ator single-writer, e o coletor funde as trajetórias em ordem de
  recebimento. Neste ambiente, o escalonador drenou a caixa de c1 por
  completo antes de escalonar c2 (c1 criado e mensageado primeiro), de
  modo que a trajetória fundida ficou contígua por capacidade (P8 PASS).
  Em B/C/E, um único despachante/ordem total preserva o interleaving
  (P8 FAIL). A diferença pode ser propriedade do paradigma (isolamento
  por caixa postal) OU artefato do escalonador deste ambiente.

Evidence Required
  (a) Análise do trace fundido de S4 em execuções de D (verificar
  contiguidade por capacidade); (b) repetição do experimento em outro
  ambiente/escalonador (ex.: GOMAXPROCS distintos, Linux) para medir se a
  distribuição de P8 varia; (c) Fase 5B (bateria de A) para completar a
  tabela de poder discriminativo.

Status
  HIPÓTESE. Não comprovada pelo Oracle nem por análise causal. Requer as
  evidências acima.
```

## Notas metodológicas

1. **Frase congelada (regra 1 do Arquiteto):** as três implementações
   independentes (B, C, E) convergem para FAIL em S4×P8. A causa permanece
   hipótese. D acrescenta um quarto ponto de dados com veredito distinto
   (PASS) — a causa dessa divergência também permanece hipótese.
2. **AuthorizationDenied:** D emite a negação explícita (como B) quando o
   efeito é bloqueado; C/E silenciam. O Oracle aceita estratégias
   observacionais distintas e chega ao mesmo veredito em S2 (P2 = PASS) —
   evidência positiva de robustez do contrato.
3. **Poder discriminativo:** a bateria NÃO é totalmente não-discriminativa:
   S4×P8 distingue D (PASS) de B/C/E (FAIL) sob o escalonamento deste
   ambiente. A solidez dessa discriminação depende das evidências (a)
   acima — permanece hipótese.
