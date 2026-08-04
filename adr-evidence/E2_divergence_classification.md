# E2 — Classificação de Divergências

## Enquadramento (congelado)
Toda divergência é registrada em três camadas:
- Observed Results (fato)
- Supported Inference (fato)
- Architectural Interpretation (decisão)

## Classes de divergência
- Classe A — Contratual: viola decisão normativa congelada. Ação: corrigir
  ou revisitar a decisão.
- Classe B — Semântica: interpretações diferentes de especificação ambígua.
  Ação: esclarecer a especificação.
- Classe C — Operacional: concorrência/escalonamento/ambiente. Ação: medir
  e documentar.
- Achado de sensibilidade de mecanismo (saída primária do harness): uma
  propriedade é satisfeita por alguns mecanismos e não por outros, por
  construção. Ação: registrar como evidência. Não corrigir — corrigir
  destruiria a evidência.

## Divergências observadas

### 1. P8×S4 — B/C/E=FAIL, D=PASS
Observed: B, C e E produzem interleaving observável; D não.
Inference: B/C/E não implementam isolamento suficiente para satisfazer
P8 nos cenários S4 e S8. D apresenta comportamento compatível com
isolamento por caixa postal na maior parte das execuções observadas,
embora esse comportamento seja sensível ao escalonamento.
Interpretation: P8 requer mecanismo explícito de isolamento. Achado de
sensibilidade de mecanismo. Sem correção.

### 2. P8×S8 — B/C/E=FAIL, D=PASS
Observed: réplica de P8×S4 com admissões.
Interpretation: confirma #1; aumenta confiança. Achado. Sem correção.

### 3. P8×S9 (B/C/E) — PASS
Observed: script sequencial (sem intercalação) → B/C/E passam.
Interpretation: controle. Confirma que #1/#2 são causados pela
intercalação do cenário, não por algo intrínseco a B/C/E.

### 4. P8×S9 (D) — PASS modal, 48/50 em 1 de 15 baterias
Observed: mesmo em script sequencial, D apresentou 2 FAIL em uma bateria.
Inference: a concorrência dos atores pode intercalar emissões mesmo sem
intercalação no script; consistente com a sensibilidade a GOMAXPROCS da
Fase 5B.
Interpretation: Classe C — Operacional. Medir/documentar. (Estudo
GOMAXPROCS opcional para quantificar.)

### 5. P3×S10 — B/C/D=FAIL, E=PASS
Observed: após Fail+Rollback, efeito sem reautorização: B/C/D emitem
(FAIL); E bloqueia (PASS). Estável em 750 execuções de D.
Inference: o rollback de B/C/D restaura autorização; o de E força posição
não efeito-elegível.
Interpretation: Achado de sensibilidade de mecanismo — P3-sob-rollback é
mecanismo-dependente. Registrado sob a DECISÃO A (reautorização exigida);
alternativa (corrigir B/C/D) anotada, não adotada.

## Conclusão
Nenhuma divergência atual é defeito (arquitetura falhando sua própria
spec). As divergências são achados de sensibilidade de mecanismo (#1–#3,
#5) ou operacionais (#4). O harness registra a matriz; não força
convergência. A bateria agora discrimina em duas famílias independentes
de comportamento: isolamento concorrente (P8) e semântica de rollback
(P3×S10).

## Status do E2

**E2 = CLOSED** (encerramento formal)

### Conclusão
O E2 atingiu seu objetivo: aumentou o poder discriminativo da bateria
em duas famílias independentes de sensibilidade de mecanismo (P8
isolamento concorrente; P3 semântica de rollback), sem introduzir
regressões nem corrigir comportamentos. Todas as divergências foram
classificadas como achados de sensibilidade de mecanismo, não defeitos.

### Decisões metodológicas congeladas
- **A permanece não comparada por decisão metodológica consciente**:
  executar A agora não responderia uma pergunta formulada previamente,
  seria coleta de dados por critério de completude documental. A não
  será executada até que uma Questão de Pesquisa Aberta justifique.
- **Nenhuma correção arquitetural decorre automaticamente dos achados**:
  divergências são registradas como evidência; corrigir mecanismos
  para satisfazer o contrato é decisão futura, não consequência
  necessária do E2.
- **Toda nova fase experimental deve declarar explicitamente a pergunta
  de pesquisa que pretende responder antes da elaboração de qualquer
  plano de implementação.** O roadmap é dirigido por "o que ainda não
  sabemos", não por "o que ainda não testamos". Proteção contra
  expansão especulativa.

### Status das hipóteses (atualizadas pelo E2)

| Hipótese | Status anterior (E1) | Status atual (E2) |
|---|---|---|
| **H1** — Contrato abstrai mecanismo | Suportada para P1–P7 em B/C/E | **Suportada em escopo ampliado.** O E2 expandiu o domínio observado (novos cenários e protocolo); a hipótese permanece verdadeira para o conjunto ampliado de propriedades exercitadas, mas não é "provada" universalmente. |
| **H2** — Bateria com baixo poder discriminativo | Suportada | **Resolvida no contexto da bateria original.** O E2 demonstrou que o baixo poder discriminativo podia ser reduzido por expansão dirigida da bateria. A hipótese permanece verdadeira como descrição do E1, não como descrição da bateria após o E2. |
| **H3** — P8/S4 discrimina Actor Model | Aberta, questionada pelos próprios dados | **Continua aberta**: evidência adicional em S8/S9 aumenta o peso observacional mas não resolve a causalidade (sensibilidade a escalonador permanece). |
| **H4** — A não pode ser comparada sem decisões de projeto | Suportada | **Inalterada**: A permanece não comparada por decisão metodológica, não por pendência. |
