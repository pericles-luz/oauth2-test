# OAuth2 Test Tool - Sindireceita

Ferramenta web para testar e debugar o fluxo OAuth2 do Sindireceita. Implementa um cliente OAuth2 completo com suporte a PKCE, permitindo configuração dinâmica de credenciais e visualização detalhada de todas as requisições HTTP.

## Características

- ✅ **Fluxo OAuth2 Completo** - Authorization Code Flow com PKCE obrigatório
- ✅ **Configuração Dinâmica** - Configure client_id, client_secret e scopes via interface web
- ✅ **Histórico Persistente** - Todas as requisições HTTP são salvas em SQLite para análise
- ✅ **Testes de Endpoints** - Token refresh, revocation, JWKS validation, OIDC discovery
- ✅ **Interface HTMX** - SPA-like experience sem JavaScript framework pesado
- ✅ **Validação JWT** - Valida tokens usando JWKS do servidor

## Requisitos

- Go 1.25+
- Credenciais OAuth2 do Sindireceita (client_id e client_secret)

## Instalação

### 1. Clone o repositório

```bash
git clone https://github.com/pericles-luz/oauth2-test.git
cd oauth2-test
```

### 2. Configure as variáveis de ambiente

```bash
cp .env.example .env
# Edite .env com suas configurações
```

### 3. Compile o projeto

```bash
go build -o oauth2-test ./cmd/server
```

### 4. Execute o servidor

```bash
./oauth2-test
```

O servidor iniciará em `http://localhost:8080` (ou na porta configurada em `SERVER_PORT`).

## Uso

### 1. Configuração Inicial

1. Acesse `http://localhost:8080`
2. Preencha o formulário com suas credenciais OAuth2:
   - **Client ID** - Identificador do seu aplicativo
   - **Client Secret** - Chave secreta (mantenha seguro!)
   - **Redirect URI** - Deve ser `http://localhost:8080/auth/callback` ou o domínio configurado
   - **Scopes** - Selecione as permissões desejadas (openid é obrigatório)
3. Clique em "Salvar Configuração"

### 2. Testar Fluxo OAuth2

1. Após salvar a configuração, clique em "Iniciar Fluxo OAuth2"
2. Você será redirecionado para a página de login do Sindireceita
3. Faça login com CPF + token de 6 dígitos (ou certificado digital)
4. Autorize as permissões solicitadas
5. Você será redirecionado de volta para o dashboard com as informações do usuário

### 3. Dashboard

O dashboard exibe:

- **Informações do Usuário** - Dados retornados do endpoint `/oauth2/userinfo`
- **Tokens** - Access token, refresh token, ID token (JWT)
- **Testes de Endpoints** - Botões para testar funcionalidades adicionais
- **Histórico** - Link para visualizar todas as requisições HTTP

### 4. Histórico de Requisições

Acesse `/history` para ver todas as requisições HTTP capturadas:

- Método, URL, status, duração
- Clique em qualquer linha para ver detalhes completos
- Headers, body, response completos

## Endpoints da API

| Rota | Método | Descrição |
|------|--------|-----------|
| `/` | GET | Página de configuração |
| `/config` | POST | Salvar configuração OAuth2 |
| `/auth/login` | GET | Iniciar fluxo OAuth2 |
| `/auth/callback` | GET | Callback OAuth2 |
| `/dashboard` | GET | Dashboard pós-autenticação |
| `/test/refresh` | POST | Testar refresh token |
| `/test/revoke` | POST | Revogar access token |
| `/test/jwks` | GET | Validar JWT com JWKS |
| `/test/discovery` | GET | OIDC Discovery |
| `/history` | GET | Listar histórico |
| `/history/{id}` | GET | Detalhes de requisição |

## Estrutura do Projeto

```
.
├── cmd/server/           # Application entry point
│   └── main.go
├── internal/
│   ├── handlers/         # HTTP handlers
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── storage/          # Database operations
├── migrations/           # SQL migrations
├── static/               # CSS e JavaScript
│   ├── css/styles.css
│   └── js/htmx.min.js
├── templates/            # HTML templates
│   ├── base.html
│   ├── home.html
│   ├── dashboard.html
│   ├── history.html
│   └── endpoints/
└── go.mod
```

## Scopes Disponíveis

| Scope | Descrição |
|-------|-----------|
| `openid` | Obrigatório para OIDC |
| `profile` | Nome, CPF, seccional |
| `email` | Endereço de e-mail |
| `phone` | Número de telefone |
| `address` | Endereço completo |
| `membership` | Status de filiação e vínculo |
| `permissions` | Permissões e roles |
| `union_unit` | Detalhes da seccional |

## Tecnologias Utilizadas

- **Go 1.25+** - Linguagem de programação
- **Chi Router** - HTTP router
- **golang.org/x/oauth2** - Cliente OAuth2 com PKCE
- **HTMX 2.0** - Interatividade sem JavaScript pesado
- **SQLite** - Banco de dados para histórico
- **HTML/CSS** - Interface responsiva

## Deploy em Produção

### Build

```bash
CGO_ENABLED=0 go build -o oauth2-test ./cmd/server
```

### Estrutura no Servidor

```
/opt/oauth2-test/
├── oauth2-test           # Binary
├── templates/            # Diretório completo
├── static/               # Diretório completo
├── .env                  # Configuração
└── oauth2-test.db        # Criado em runtime
```

### HAProxy Configuration

```
frontend oauth2_frontend
    bind *:80
    bind *:443 ssl crt /etc/haproxy/certs/oauth2.sindireceita.org.br.pem
    acl is_oauth2 hdr(host) -i oauth2.sindireceita.org.br
    use_backend oauth2_backend if is_oauth2

backend oauth2_backend
    server oauth2_app 127.0.0.1:8080 check
```

### Systemd Service

```ini
[Unit]
Description=OAuth2 Test Tool
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/oauth2-test
EnvironmentFile=/opt/oauth2-test/.env
ExecStart=/opt/oauth2-test/oauth2-test
Restart=always

[Install]
WantedBy=multi-user.target
```

## Segurança

- ✅ PKCE obrigatório (S256)
- ✅ State parameter para proteção CSRF
- ✅ Session cookies HTTP-only
- ✅ Validação de JWT via JWKS
- ✅ HTTPS obrigatório em produção

## Obter Credenciais OAuth2

Para obter credenciais OAuth2 (Client ID e Client Secret), entre em contato com:

**dti@sindireceita.org.br**

Forneça:
- Nome do aplicativo
- Logo (opcional)
- Redirect URIs
- Scopes necessários
- Descrição do uso

## Licença

MIT

## Suporte

Para dúvidas ou problemas, entre em contato com a DTI do Sindireceita.
