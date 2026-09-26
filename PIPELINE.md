# Pipeline de CI/CD — docval

Documento de apoio à apresentação da pipeline definida em
[`.github/workflows/ci-cd.yml`](.github/workflows/ci-cd.yml), relacionando
cada decisão de design ao conteúdo visto em aula (V&V, Shift-Left,
Deploy vs. Release, Continuous Delivery vs. Deployment).

## Estratégia de branches e ambientes

A pipeline usa uma mecânica de **3 branches → 3 ambientes**, um GitFlow
simplificado:

| Branch | Ambiente (GitHub Environment) | Papel |
|---|---|---|
| `develop` | `development` | Integração contínua entre desenvolvedores |
| `release/X.X.X` | `homologacao` | Validação pré-produção de uma versão candidata |
| `main` | `production` | Produção |

O fluxo normal de promoção é:

```
feature/* ──PR──► develop ──PR──► release/X.X.X ──PR──► main
                  (deploy:         (deploy:              (deploy:
                   development)     homologacao)           production)
```

Cada push em `feature/**`, `develop`, `release/X.X.X` ou `main` dispara o
mesmo pipeline de verificação (estática + dinâmica + build), e só então o
job correspondente àquela branch é executado — deploy nos três ambientes,
ou abertura automática de PR/release nas branches de desenvolvimento.
Esses jobs finais são mutuamente exclusivos (`if` por branch), então nunca
mais de um roda na mesma execução.

## Visão geral dos jobs

Tudo roda em **um único workflow** (`ci-cd.yml`). Os jobs de deploy e de
automação são todos "folhas" da mesma árvore de dependências
(`needs: [build]`), e cada um só executa para a branch que lhe diz
respeito — nunca mais de um por execução:

```
static-analysis ─┐
                  ├─► test ─┐
codeql ───────────┘         ├─► build ─┬─► deploy-development   (branch develop)
                             │          ├─► deploy-homologacao   (branch release/X.X.X)
                             │          ├─► deploy-production    (branch main)
                             │          ├─► auto-pr-feature      (branch feature/**)
                             │          └─► auto-release         (branch develop)
```

| # | Job | O que faz | Dispara em |
|---|-----|-----------|------------|
| 1 | `static-analysis` | `gofmt`, `go vet`, `golangci-lint`, `gitleaks` | todo push/PR em feature/\*\*, develop, release/\*\* ou main |
| 1b| `codeql` | SAST nativo do GitHub (CodeQL) | todo push/PR em feature/\*\*, develop, release/\*\* ou main |
| 2 | `test` | `go test -race` + quality gate de cobertura (≥ 80%) | após passo 1 |
| 3 | `build` | build multiplataforma, imagem Docker (tag = SHA, só publicada em develop/release/main), scanner Trivy | após passos 1 e 2 |
| 4a| `deploy-development` | promove a imagem para `:dev`, smoke test | push em `develop` |
| 4b| `deploy-homologacao` | promove a imagem para `:homolog-X.X.X`, smoke test completo | push em `release/X.X.X` |
| 4c| `deploy-production` | promove a imagem para `:latest`/`:X.X.X`, cria Release | push em `main` |
| 5a| `auto-pr-feature` | abre PR `feature/* → develop` (se ainda não existir) | push em `feature/**`, **só se 1, 1b, 2 e 3 passarem** |
| 5b| `auto-release` | cria `release/X.X.X` e abre PR `→ main` | push em `develop`, **só se 1, 1b, 2 e 3 passarem** |

Importante: o **artefato é o mesmo em todos os ambientes** — o job `build`
compila e escaneia a imagem uma única vez (tag = SHA do commit); os jobs de
deploy apenas repaginam essa mesma imagem com uma nova tag (`:dev`,
`:homolog-X.X.X`, `:latest`/`:X.X.X`). Isso garante que o que foi validado
em homologação é bit a bit o mesmo artefato que vai para produção.

## Relação com o conteúdo da aula

**Verificação (estática) vs. Validação (dinâmica).**
Os jobs 1 e 1b fazem verificação: não executam o programa, apenas analisam
o código-fonte (linting, análise de complexidade, SAST, varredura de
segredos). O job 2 faz validação dinâmica: executa a suíte de testes
unitários de fato, com o *race detector* ativado. Os smoke tests dos jobs
4a/4b executam o binário compilado de verdade em cada ambiente.

**Shift-Left.** A ordem dos jobs é deliberada: a verificação estática
(segundos) roda antes dos testes dinâmicos (mais lenta), que rodam antes do
build (mais lento ainda). Um erro de formatação é barrado em segundos, sem
gastar tempo de CI com testes ou build de imagem.

**Artefato imutável.** O job `build` compila os binários e a imagem Docker
uma única vez, versionados pelo SHA do commit
(`ghcr.io/.../docval:<sha>`). Esse mesmo artefato é promovido — nunca
recompilado — de `development` a `homologacao` e depois a `production`,
apenas recebendo novas tags.

**Deploy vs. Release.** Os três jobs de deploy são atos técnicos de
*deploy* (implantar a imagem em um ambiente). A criação da *Release* no
GitHub (binários + changelog), disparada só quando a imagem chega a
produção, é o ato de *release* — o momento em que a entrega é formalmente
disponibilizada aos usuários.

**Continuous Delivery, não Continuous Deployment.** Cada branch fica
sempre em estado liberável para o seu ambiente, mas avançar de um ambiente
para o outro é uma decisão deliberada (abrir e mergear um PR de
`develop` → `release/X.X.X` → `main`), não algo automático a cada commit
isolado. Isso corresponde a Continuous **Delivery**.

## Os três ambientes (GitHub Environments)

A pipeline usa três ambientes do GitHub Actions (`Settings → Environments`),
que aparecem na aba **Environments** do repositório com o histórico de
deploys de cada um — eles são criados automaticamente na primeira vez que
o workflow referenciar cada nome (`development`, `homologacao`,
`production`), mas é recomendável configurar manualmente:

- **`development`** — sem restrições; recebe todo push em `develop`.
- **`homologacao`** — pode ter *Deployment branches* restrito a
  `release/**`, para impedir que qualquer outra branch implante ali.
- **`production`** — deve ter **Required reviewers** configurado, exigindo
  aprovação manual de um responsável antes do job rodar, e *Deployment
  branches* restrito a `main`.

## Automação do fluxo (Auto PR / Auto Release)

Os jobs `auto-pr-feature` (5a) e `auto-release` (5b), dentro do próprio
`ci-cd.yml`, automatizam a *orquestração* entre branches — a única ação
manual que sobra para quem desenvolve é abrir uma `feature/*` a partir de
`develop`, dar push, e depois aprovar/mergear os PRs que o próprio
pipeline abre.

**O ponto central: os dois são o último elo da cadeia `needs`.** Como
`auto-pr-feature` e `auto-release` declaram `needs: [build]`, e `build`
por sua vez depende de `test` e `codeql` (que dependem de
`static-analysis`), o GitHub Actions **pula automaticamente** esses jobs
se qualquer verificação anterior falhar — é o comportamento padrão de
`needs`, sem precisar de nenhuma condição extra. Um push com lint
quebrado, teste falhando ou vulnerabilidade crítica na imagem nunca abre
PR nem cria branch de release.

| Job | Dispara quando | Faz o quê |
|---|---|---|
| `auto-pr-feature` | push em `feature/**`, **e** static-analysis + codeql + test + build com sucesso | Abre PR `feature/* → develop` automaticamente (se ainda não existir) |
| `auto-release` | push em `develop`, **e** static-analysis + codeql + test + build com sucesso | Cria a branch `release/X.X.X` a partir de `develop` **e** já abre o PR `release/X.X.X → main` |

Fluxo ponta a ponta:

```
dev cria "feature/login" a partir de develop
        │  git push origin feature/login
        ▼
 ci-cd.yml valida (lint, CodeQL, testes, build)
        │
        ├── falhou em qualquer etapa ──► pipeline para aqui, NENHUM PR é aberto
        │
        └── tudo passou ──► job auto-pr-feature abre PR feature/login -> develop
                    │
                    ▼
             pessoa revisa e MERGEIA o PR em develop  (ação manual)
                    │  (esse merge é, ele mesmo, um novo push em develop)
                    ▼
             ci-cd.yml valida de novo (lint, CodeQL, testes, build)
                    │
                    ├── falhou ──► pipeline para aqui, NENHUMA release é criada
                    │
                    └── tudo passou ──► job auto-release cria release/X.X.X
                                (versão calculada automaticamente) e já
                                abre PR release/X.X.X -> main
                                (push em release/X.X.X dispara o deploy
                                 em "homologacao")
                    │
                    ▼
             pessoa valida em homologação e MERGEIA o PR em main  (manual)
                    │
                    ▼
             ci-cd.yml  ──►  deploy em "production" (com aprovação manual,
                              se o ambiente tiver Required reviewers)
```

A versão `X.X.X` é calculada automaticamente pelo job `auto-release`: ele
olha a última branch `release/X.X.X` já criada e soma 1 ao número de
patch (ex.: `1.2.0` → `1.2.1`); se não existir nenhuma ainda, começa em
`0.1.0`. Ao chegar em `main`, essa mesma versão é recuperada da mensagem
do commit de merge pelo job `build` — usada para nomear a tag da imagem e
a Release no GitHub.

As únicas ações manuais que continuam existindo de propósito são os dois
cliques de **aprovar/mergear o PR** — é o ponto de controle humano entre
um ambiente e o próximo (equivalente ao *Required reviewers* de um
ambiente sensível, discutido em aula). Tudo o mais — abrir PR, criar
branch de release, promover a imagem entre ambientes — é automático e
condicionado ao sucesso das verificações.

### Configuração da automação (obrigatório, uma única vez)

Os jobs `auto-pr-feature` e `auto-release` escrevem no repositório (criam
branch, abrem PR) e **por padrão do GitHub Actions, ações feitas com o
token automático (`GITHUB_TOKEN`) não disparam outros workflows** — é uma
proteção anti-loop. Isso significa que, sem configurar nada, a branch
`release/X.X.X` seria criada mas o `ci-cd.yml` **não rodaria** sobre ela.

A solução recomendada pelo próprio GitHub é usar um **Personal Access
Token (PAT)** seu nesses dois jobs, guardado como o secret `GH_PAT`:

1. Crie um token em **github.com → Settings → Developer settings → Personal
   access tokens → Fine-grained tokens → Generate new token**.
2. Restrinja o **Repository access** a este repositório
   (`cpf-cnpj-validator`).
3. Em **Permissions → Repository permissions**, conceda:
   - `Contents`: **Read and write**
   - `Pull requests`: **Read and write**
4. Gere o token e copie o valor.
5. No repositório: **Settings → Secrets and variables → Actions → New
   repository secret**, nome `GH_PAT`, valor = o token copiado.

Sem esse secret configurado, os jobs `auto-pr-feature` e `auto-release`
falham de propósito com uma mensagem explicativa (`::error::`), em vez de
rodar silenciosamente com o `GITHUB_TOKEN` padrão e produzir um PR "morto"
que nunca dispara validação.

## Rollback

Como o artefato é imutável e versionado, o rollback é apenas redirecionar
a tag `:latest` para uma versão anterior conhecida:

```bash
docker pull ghcr.io/<owner>/<repo>:1.1.0        # versão anterior conhecida
docker tag  ghcr.io/<owner>/<repo>:1.1.0 ghcr.io/<owner>/<repo>:latest
docker push ghcr.io/<owner>/<repo>:latest
```

## Ferramentas de verificação estática e dinâmica usadas

- **Estática:** `gofmt`, `go vet`, `golangci-lint` (agrega `errcheck`,
  `staticcheck`, `ineffassign`, `gocritic`, `revive`), **CodeQL**
  (`github/codeql-action`), **Gitleaks** (varredura de segredos), **Trivy**
  (varredura de CVEs na imagem — estática sobre as camadas do contêiner).
- **Dinâmica:** `go test -race` (testes unitários com detecção de *race
  conditions*), quality gate de cobertura de testes, smoke test do binário
  compilado rodando de fato em development e homologação.
