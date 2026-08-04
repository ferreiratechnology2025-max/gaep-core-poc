# E1 — Relatório de Poder Discriminativo do Oracle

Artefato de síntese do E1 (Architecture Validation Harness). Consolida as
Fases 3–5B: implementação independente de B (Context Object + Policy
Pipeline), C (Event Sourcing), D (Actor Model) e E (Hierarchical State
Machine) contra a bateria S1–S5, avaliadas apenas pelo Property Oracle
(P1–P8) — nenhuma arquitetura se auto-avalia. A é reportada como **não
comparada** (ver §4). Estrutura: Resultados observados (fatos brutos,
sem interpretação) → Inferências suportadas (afirmações proporcionais à
evidência) → Hipóteses abertas (não comprovadas) → Limitações → Agenda do
E2. Misturar essas camadas contaminaria a conclusão; mantidas separadas
deliberadamente.

## 1. Resultados observados

### 1.1 Matriz B/C/E (idêntica nas três — convergência verificada ao vivo em `d_test.go::refBCEMatrix`)

| property | S1 | S2 | S3 | S4 | S5 |
|---|---|---|---|---|---|
| P1 Auditabilidade | PASS | PASS | PASS | PASS | PASS |
| P2 Autorização antes do efeito | UNKNOWN | PASS | UNKNOWN | UNKNOWN | UNKNOWN |
| P3 Falha bloqueia efeitos | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P4 Estado parcial recuperável | PASS | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P5 Tolerância a execução assíncrona | UNKNOWN | UNKNOWN | PASS | UNKNOWN | UNKNOWN |
| P6 Kernel independente do domínio | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |
| P7 Admissão determinística | UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN | PASS |
| P8 Isolamento de execuções | UNKNOWN | UNKNOWN | UNKNOWN | **FAIL** | UNKNOWN |

B, C e E são mecanismos estruturalmente distintos — pipeline de estágios
puros sobre contextos imutáveis; log append-only + projeção; máquina de
estados hierárquica por capacidade — e produzem a matriz **byte-a-byte
idêntica**, incluindo o mesmo FAIL em S4×P8.

### 1.2 Matriz D (Actor Model, cláusula estocástica N=50, veredito modal — `E1_D_findings.md`)

Idêntica à de B/C/E em toda célula informativa, **exceto S4×P8**: D = PASS
modal (50/50, freq 100%, estável em 500 execuções nas Fases 3–5) contra
FAIL de B/C/E. Única divergência de toda a bateria (10/11 células
informativas convergem). Cenários de capacidade única (S1/S2/S3/S5) são
massa pontual em todas as N=50 execuções — sem variância.

### 1.3 Verificação de concorrência de D (Fase 5B Parte 0 — `E1_A_findings.md`)

Reexecução de S4 sob GOMAXPROCS=1 e GOMAXPROCS=4 (N=50 cada, mesmo script,
D e pré-registro intocados):

| regime | PASS | FAIL | freq PASS |
|---|---|---|---|
| GOMAXPROCS=1 | 50 | 0 | 100% |
| GOMAXPROCS=4 | 49 | 1 | 98% |

Resultado idêntico em 5 repetições consecutivas. O PASS de 500/500 das
Fases 3–5 (regime de paralelismo não fixado) deixou de ser tratável como
determinístico por construção do paradigma: sob paralelismo real, o Oracle
já registrou 1 FAIL em 50.

### 1.4 Estratégias observacionais distintas convergindo (S2×P2)

Em S2, D emite `AuthorizationDenied` explicitamente antes da autorização
(como B); C e E não emitem evento algum nesse ponto (silenciam a
tentativa). O Oracle aceita as duas estratégias e chega ao mesmo veredito
(P2 = PASS) nas quatro arquiteturas. É o único ponto da bateria em que
duas formas observacionalmente diferentes de "recusar o mesmo efeito"
foram exercitadas e avaliadas de forma idêntica pelo Oracle.

### 1.5 Arquitetura A — não comparada

A (registradores $R0$–$R7$ + microkernel + lifecycle) não foi submetida à
bateria. O Passo 0 da Fase 5B identificou três decisões arquiteturais não
cobertas pela spec congelada de A (identidade de capacidade, política de
admissão por identidade, rollback vs. estado absorvente em
`models_comp/lifecycle.go`) e nenhuma delas foi resolvida por decisão
prévia do Arquiteto — ao contrário de B/C/D/E, que tiveram essas decisões
resolvidas na Fase 2. Decisão do Arquiteto: não especificar A' post-hoc
(ver §4).

## 2. Inferências suportadas

Afirmações proporcionais à evidência acima — não hipóteses, mas também não
fatos garantidos além do que os dados mostram.

**I1 — O contrato observável abstrai o mecanismo para P1, P2, P3, P4, P5 e
P7.** Três implementações independentes, de paradigmas genuinamente
distintos, convergem byte-a-byte nessas seis propriedades em toda a
bateria. Esta é a inferência mais bem sustentada do E1: não depende de
nenhuma célula isolada, é a convergência de uma matriz inteira replicada
três vezes.

**I2 — A bateria atual tem baixo poder discriminativo.** Das 40 células
da matriz P1–P8 × S1–S5, apenas 1 (S4×P8) separa alguma arquitetura de
outra; as demais são convergência ou UNKNOWN (cenário não exercitou a
propriedade). Um harness cujo objetivo é comparar arquiteturas rivais e
que produz discriminação em 1/40 células não está, no estado atual,
cumprindo bem esse objetivo — mesmo sendo, ao mesmo tempo, uma evidência
forte a favor de I1.

**I3 — P8, tal como operacionalizado (contiguidade por capacidade no
trace), é a única propriedade da bateria atual com sensibilidade
observada a diferenças de mecanismo/ambiente.** É também a única
propriedade cujo resultado mudou dentro da MESMA arquitetura (D) apenas
variando GOMAXPROCS — o que a torna, ao mesmo tempo, o ponto mais
interessante e o mais frágil da matriz atual.

**I4 — O Oracle tolera estratégias observacionais distintas para a mesma
decisão sem perder poder de veredito** (§1.4). É evidência de robustez do
vocabulário de eventos (não exige uma única forma de recusar um efeito),
mas é um único ponto de dados.

## 3. Hipóteses abertas (não comprovadas)

**H1 — "O contrato observável abstrai o mecanismo interno."**
Suportada por I1 para P1–P5/P7. Não testada para P6 (estrutural, fora do
protocolo de eventos) nem P8 de forma robusta (ver H3). Condição de
refutação: uma quinta implementação independente que produza uma matriz
divergente em qualquer célula hoje convergente.

**H2 — "A bateria S1–S5 não tem poder discriminativo suficiente para
comparar arquiteturas" (N9).** Suportada por I2. Enfraquecida, não
refutada, pela existência de S4×P8. Condição de refutação: expandir a
bateria (agenda do E2, §5) e observar se novas células continuam
convergindo ou passam a discriminar.

**H3 — "P8/S4 discrimina Actor Model de mecanismos com despachante
único."** Enfraquecida pela Parte 0 (§1.3): sob GOMAXPROCS=4 a mesma
arquitetura D já produziu 1 FAIL em 50, então "D = PASS em P8" não é uma
propriedade estável do paradigma neste estado da evidência — pode ser
sensibilidade ao escalonador. Permanece hipótese de trabalho (não
descartada), rebaixada de "suportada" para "aberta e ativamente
questionada pelos próprios dados". Condição de resolução: as evidências
(a)–(c) já listadas em `E1_D_findings.md` e `E1_A_findings.md`.

**H4 — "A não pode ser comparada sem decisões de projeto que
descaracterizariam a arquitetura testada."** Suportada pelo Passo 0 da
Fase 5B (§1.5): os três bloqueios não são lacunas de implementação
resolvíveis por leitura mais cuidadosa do código, mas ausências de decisão
que qualquer resolução tornaria uma arquitetura nova (A'), não A. Não
refutável dentro do E1 tal como desenhado — só por uma decisão explícita e
antecipada do Arquiteto sobre o que A deveria ser, tomada antes (não
depois) de rodar o experimento.

## 4. Limitações

1. **A não comparada.** Ver §1.5/H4. O E1 compara 4 de 5 arquiteturas
   candidatas; a conclusão sobre poder discriminativo (I2) e sobre
   abstração de mecanismo (I1) é sustentada apenas por B/C/D/E.
2. **`-race` não executado.** Mesma limitação em todas as fases:
   `CGO_ENABLED=1` ausente neste toolchain (`go: -race requires cgo`).
   Concorrência de D avaliada apenas por comportamento observável
   (Oracle), não por detecção de data race em nível de memória.
3. **Poder discriminativo limitado** (I2) — a bateria atual não é
   suficiente para separar a maioria dos pares de arquiteturas na maioria
   das propriedades; ausência de discriminação em 39/40 células é tanto
   evidência positiva (I1) quanto limitação de desenho experimental.
4. **H3 e H4 são correlacionais, não causais.** Nenhum bloco
   Observed/Hypothesized Cause deste relatório ou dos findings por
   arquitetura estabelece mecanismo causal; todos permanecem HIPÓTESE por
   desenho, conforme a regra de decisão pré-registrada do E1.

## 5. Agenda do E2

Consequência direta de I2/H2: o gargalo do E1 não é a falta de mais
paradigmas candidatos, é a bateria não discriminar os que já existem. A
prioridade do E2 é expandir poder discriminativo antes de expandir a
lista de arquiteturas:

1. **Novos cenários (S6–S8)** desenhados especificamente para produzir
   divergência entre mecanismos hoje convergentes — não variações dos
   cenários existentes, mas condições que só um subconjunto dos
   paradigmas consegue satisfazer (ex.: recuperação após falha parcial de
   *duas* capacidades simultaneamente, admissão sob política que muda
   durante a execução).
2. **Novas propriedades (P9–P11)**, com critério operacional definido
   *antes* da implementação (mesma disciplina de pré-registro usada para
   P1–P8), visando cobrir o que a bateria atual deixa em UNKNOWN na
   maioria das células (P2/P3/P4/P5/P6/P8 fora dos cenários que os
   disparam diretamente).
3. **Revisão do critério operacional de P8** (contiguidade no trace) à
   luz de I3/H3: um critério mais robusto a variação de escalonador (ou
   um cenário que force interleaving de forma determinística, em vez de
   depender do runtime) resolveria a fragilidade observada em §1.3 sem
   depender de repetir o experimento em outros ambientes.
4. **A permanece fora do E2** até existir uma decisão explícita e
   pré-registrada do Arquiteto sobre os três pontos do §1.5 — tomada
   antes de qualquer tentativa de implementação, não como adaptação
   posterior aos resultados.

## 6. Conclusão

| pergunta | resposta | força da evidência |
|---|---|---|
| O contrato abstrai mecanismo para P1–P5/P7? | Sim, dentro do escopo testado (três paradigmas independentes, B/C/E, mais D nas células não afetadas por concorrência) — ressalva: a mesma convergência é consistente com I2 (bateria de baixo poder discriminativo); as duas leituras não se excluem. | Forte — convergência replicada 3–4×, mas ambígua entre "abstração real" e "bateria fraca" sem I2 ao lado. |
| A bateria discrimina mecanismos? | Não, em 39/40 células. | Forte — mesma convergência, lida como limite. |
| P8 mede isolamento de forma estável? | Não estabelecido — o único resultado positivo (D×S4) muda com GOMAXPROCS na mesma arquitetura. | Fraca — 1 ponto de dados, já contestado pelos próprios dados. |
| A foi comparada? | Não — bloqueada por decisões de projeto não cobertas pela spec congelada. | N/A — ausência documentada, não resultado. |

O E1 encerra como experimento metodologicamente consistente: a conclusão
é proporcional à evidência, o achado positivo (I1) e o achado de limite
(I2) são reportados lado a lado sem que um dilua o outro, e nenhuma
hipótese é apresentada como fato. A agenda do E2 (§5) é a consequência
direta desse resultado — discriminar antes de expandir.
