# Matriz de Falsificação — GAEP Core POC

> **SUPERSEDED como fonte da verdade** por
> `architectural-necessity-matrix.md` (cadeia completa
> Need→Property→Mechanisms→Experiments→Evidence→Decision). Mantida como
> evidência de remoção em nível de componente (E0).

> **Regra metodológica:** nenhum conceito entra na Architecture Specification
> antes de sobreviver a um experimento de falsificação. Tente primeiro
> eliminá-lo — ou construa um rival que o substitua. Só permanece na
> especificação aquilo cuja remoção destrói uma propriedade essencial do
> sistema (P1–P8 em `architectural-properties.md`). Um harness prova
> comportamento de *arquiteturas*, não de código — a evidência relevante é
> "o rival preservou a propriedade?", não "a implementação funcionou?".

**Status:**
- **Não-falsificado** — nenhuma remoção tentada, ou a remoção tentada quebrou a propriedade (o conceito segura sua posição por enquanto).
- **Falsificado** — a remoção não quebrou nenhuma propriedade obrigatória (candidato a sair da spec).
- **Lacuna de requisito** — não é possível definir *qual* propriedade obrigatória quebraria; ou a propriedade é especificada, ou o conceito sai.

| Conceito | Hipótese | Como tentar removê-lo | Propriedade que deve quebrar | Evidência atual | Status |
|---|---|---|---|---|---|
| **R0–R7 (unidade)** | Primitiva arquitetural | Substituir por estado tipado fechado (Modelo B′) | Qualquer propriedade que B′ não preserve (auditabilidade determinística? isolamento?) | E1v2: B′ = C em M1/M1b/M2/M3/auditoria | **Não-falsificado; evidência favorece "representação"** → decisão ADR |
| **R0 Session** | Sessão precisa viver no kernel como célula | Session Context externo ao kernel | L2/L3 (ciclo de vida exige sessão ativa) — prova o *conceito*, não a localização | E0: SESSION load-bearing no lifecycle; R0 não pinado por I0–I4 | Conceito essencial; **localização não demonstrada** |
| **R1 Watchdog** | Watchdog é do kernel | Delegar watchdog ao Runtime | I4 (desarmado ou futuro) + L4 (timeout força RELEASE) | E0: essencial | Não-falsificado |
| **R2 NextAct** | Sequência do próximo ato | Derivar de eventos/comandos | I1 (monotonicidade do próximo ato) | E0: essencial (mas derivável? pergunta em aberto) | Não-falsificado; derivabilidade em aberto |
| **R3 Isolation** | Isolamento garantido pelo kernel | Garantia do Runtime | — não definida | E0: não pinado por nenhum invariante | **Lacuna de requisito** |
| **R4 Stderr** | Contador de stderr no kernel | Log externo ao kernel | — não definida | E0: não pinado | **Lacuna de requisito** |
| **R5 SKK Badge** | Token de autorização no kernel | Authorization Token externo | — não definida | E0: não pinado | **Lacuna de requisito** |
| **R6 Health** | Status de saúde no kernel | Derivar de checks externos | I3 (domínio válido de saúde) | E0: essencial | Não-falsificado |
| **R7 Timestamp + geração de escrita** | Audit ledger do kernel | Sem carimbo automático por escrita | I2 (auditabilidade determinística: o timestamp avança a cada escrita) | E0/E1v2: essencial para I2 | Não-falsificado |
| **WAIT_OUTCOME** | Estado do kernel (participante ativo do protocolo) | Mover para o protocolo Runtime↔Kernel (contrato assíncrono) | Governança assíncrona do kernel (timeout, accountability, isolamento enquanto pendente)? | Nenhuma — experimento E1 da nova ordem | **Não-falsificado ainda** |
| **Microkernel / Runtime** | Camada obrigatória entre Kernel e Capability | Pipeline / event processor / actor model / state machine engine | Quais garantias deixam de existir (isolamento, rollback, desacoplamento)? | Nenhuma — experimento E4 | **Não-falsificado ainda** |
| **ExecutionPlan** | Unidade correta de governança | Comandos/eventos | A política deixa de ser avaliável antecipadamente (pré-voo)? | Nenhuma — experimento E6 | **Não-falsificado ainda** |

## Pergunta de conformidade (meta)

Uma implementação conforme GAEP que use *Session Context, Transaction Object,
Authorization Token e Audit Ledger* — sem nunca expor "R0…R7" — é conforme?

- **Conforme** → registradores são representação; a spec define
  propriedades/invariantes, e R0–R7 são uma *notação normativa* (decisão ADR).
- **Não conforme** → é preciso provar *qual propriedade essencial* exige
  R0–R7 especificamente (não apenas um estado fechado e guardado).

Evidência disponível: o E1v2 mostrou que B′ (struct fechada, sem registradores)
preserva todas as propriedades medidas → **pende a favor de "conforme"**.
A decisão é ADR, e as linhas "lacuna de requisito" (R3/R4/R5) são os lugares
onde uma propriedade que *só* registradores garantem ainda poderia aparecer —
ou onde a resposta é "remover".

## Leituras obrigatórias

1. Linha com **lacuna de requisito** não é "remover já": é "a propriedade
   obrigatória não foi definida". Primeiro tente defini-la; se não for
   possível, o conceito não tem justificativa e sai.
2. **Não-falsificado** é estado provisório, não veredito: a matriz vive e cada
   experimento novo (E1–E6) atualiza uma linha.
3. ADRs resultantes registram **necessidade demonstrada** (o que a remoção
   destrói), não escolha.
