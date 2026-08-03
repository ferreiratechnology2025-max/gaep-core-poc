# Matriz de Necessidade Arquitetural — fonte da verdade

> **Fluxo (invertido — elimina viés de confirmação):**
>
> ```
> Problem Definition → Architectural Constraints → Architectural Needs →
> Observable Properties → Candidate Mechanisms → Experiments → Evidence →
> Architectural Decision
> ```
>
> A pergunta não é "o GAEP funciona?" nem "os registradores são
> necessários?". É:
>
> > **"Qual mecanismo é indispensável para qualquer sistema que satisfaça
> > esta necessidade?"**
>
> As **Needs** aqui são derivadas das **Constraints** de
> `problem-definition.md` (nível acima desta matriz) e passaram pelo lint de
> higiene (resultado puro, sem mecanismo).
>
> **Regra de decisão (três resultados possíveis por linha):**
>
> | Resultado | Decisão |
> |---|---|
> | **Nenhum mecanismo é indispensável** (a propriedade é emergente/derivável) | Remover da Architecture Specification |
> | **Vários mecanismos equivalentes** preservam a propriedade | A **propriedade** permanece na spec; o **mecanismo** desce para ADR ou Protocol Specification. *O **contrato observável** pode permanecer* se for o ponto de acoplamento entre arquiteturas (interoperabilidade, não implementação) |
> | **Apenas um mecanismo** preserva a propriedade | Congelar na architecture — a única categoria que "congela" um conceito |
>
> Arquitetura responde *que capacidade o sistema precisa possuir*; mecanismo
> responde *como essa capacidade é realizada*. **O "como" deve perder.
> Sempre. Até que sobre apenas um.** *A exceção é o contrato observável
> (ex.: o vocabulário de eventos do harness, o critério P1–P8, uma interface
> `Authorize(...)`) — ele pode ser arquitetural mesmo com mecanismos
> equivalentes, porque define o acoplamento entre arquiteturas, não a
> implementação (ver `problem-definition.md`, seção 5).*

## Cadeia por linha

```
Constraint C1 (problem-definition.md)
  ↓ gera
Need N1
  ↓ gera
Property P2
  ↓ que exige
Mecanismos candidatos: R5 (GAEP) · capability token · lease · ACL · capability object
  ↓ experimentos no harness
Evidência
  ↓
Decisão arquitetural (remover / ADR-Protocol / congelar)
```

## Matriz (v1 — a preencher com o domínio)

| Need | Property | Cenário (harness) | Mecanismos candidatos | Evidência atual | Decisão provisória |
|---|---|---|---|---|---|
| **N1** Nenhum efeito observável ocorre sem autorização válida | P2 | S2 | **R5 SKK Badge (GAEP)** · capability token · lease · ACL · capability object | E1v2 M2: a superfície guardada rejeita mutações não autorizadas; token/lease/ACL **não testados** | P2 na spec. **R5: provável "vários equivalentes" → desce para ADR** (só E1 pode provar o contrário) |
| **N2** Conseguir interromper uma execução pendente | P5 | S3 | **WAIT_OUTCOME (GAEP)** · callback · promise · channel · actor mailbox | Nenhuma (S3 é cenário novo) | Pendente; provável "vários" → **WAIT_OUTCOME candidato a descer** |
| **N3** Toda execução deve ser reconstruível do estado persistido | P1 | S1 | **R7 + geração de escrita (GAEP)** · audit ledger · event log · actor journal | I2 pinado (E0/E1v2): carimbo por escrita; o *ledger em si* não testado | Pendente — I2 prova o carimbo, não o mecanismo de ledger |
| **N4** Após uma falha, o sistema volta a um estado consistente conhecido | P4 | S1 | **rollback/Restore (GAEP)** · transaction object · event replay · actor restart | M3 (E1v2): rollback exato é **propriedade de locking** — todo modelo com mutex + Restore passa | P4 na spec. **Mecanismo: vários equivalentes → ADR** |
| **N5** Decisão de admissão determinística para a mesma entrada | P7 | S5 | **policy pipeline (GAEP)** · função pura · projeção event-sourced · máquina de estados | Oracle testa repetições (S5); arquiteturas ainda não rodadas | Pendente |
| **N6** Execuções não interferem entre si | P8 | S4 | **fronteira do microkernel (GAEP)** · encapsulamento de ator · isolamento de processo/capability · estado hierárquico | S4 é cenário novo | Pendente |
| **N7** Decisões de governança independem da semântica de domínio das ações | P6 | dual-run (estrutural) | **policy injetada (RegisterPolicy, GAEP)** · plugin · pipeline de políticas | `kernel.RegisterPolicy` é injetado externamente — núcleo sem regras de domínio | P6 na spec. **Mecanismo: provável "vários" → ADR** |
| **N8** Uma falha impede efeitos posteriores | P3 | S1 | **bloco transacional (GAEP)** · barreira de exceção · event sourcing (sem parciais) · supervisão de ator | S1 nominal passa P3 no Oracle; rivais não rodados | Pendente |
| **N9 (meta)** O conjunto P1–P8 **distingue** arquiteturas? | — | E1 completo | — | Nenhuma | **A verificar no E1** — se todas as arquiteturas passam em tudo, o conjunto não discrimina → refinar/adicionar propriedades |

## Procedimento para preencher uma linha

1. **Need** em termos de capacidade (sem mecanismo, sem solução).
2. **Property** observável gerada (P-id) + critério operacional
   (`architectural-properties.md`).
3. **Mecanismos candidatos** de princípios distintos — incluindo o do GAEP,
   sem favoritismo.
4. **Experimento**: rodar cada mecanismo (dentro de sua arquitetura
   hospedeira) pela bateria do harness (S1–S5 + variantes adversariais).
5. **Evidência**: vereditos do Oracle + logs em `adr-evidence/`.
6. **Decisão**: aplicar a regra dos três resultados.

## Discriminabilidade (N9 — anti-benchmark)

O harness deve reportar, por propriedade, **quais arquiteturas ela separou**.
Uma propriedade que nada separa é candidata a refino (ou revela que a
necessidade não é real). Sem isso, o resultado "todas passam" é lido
corretamente como *"o conjunto não distingue"*, não como "são equivalentes".

## Relação com os demais documentos

| Documento | Papel |
|---|---|
| **Este (necessity matrix)** | Fonte da verdade: cadeia Constraint→Need→Property→Mechanisms→Experiments→Evidence→Decision |
| `problem-definition.md` | Nível acima: Problem Definition + Constraints (origem das Needs) + lint de higiene |
| `architectural-properties.md` | Critérios operacionais PASS/FAIL/UNKNOWN (o elo "Property" da cadeia) |
| `falsification-matrix.md` | Evidência de remoção em nível de componente (superseded como fonte da verdade) |
| `experiment-plan.md` | Ordem E0–E6, no fluxo invertido; inclui a matriz de custo estrutural (planejada) |
| `harness/` | Oracle + cenários S1–S5 (o elo "Experimento") |
