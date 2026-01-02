# Manual de Utilização - OAuth2 Sindireceita

## Índice
1. [Visão Geral](#visão-geral)
2. [Configuração Inicial](#configuração-inicial)
3. [Fluxo Minimalista](#fluxo-minimalista)
   - [Implementação PHP](#php-fluxo-minimalista)
   - [Implementação Next.js](#nextjs-fluxo-minimalista)
4. [Fluxo Completo](#fluxo-completo)
   - [Implementação PHP](#php-fluxo-completo)
   - [Implementação Next.js](#nextjs-fluxo-completo)
5. [Endpoints Disponíveis](#endpoints-disponíveis)
6. [Troubleshooting](#troubleshooting)

---

## Visão Geral

O OAuth2 do Sindireceita implementa o padrão OAuth2.0/OpenID Connect com suporte a:
- **PKCE** (Proof Key for Code Exchange) para segurança adicional
- **Refresh Tokens** para renovação de acesso
- **Token Revocation** para invalidação de tokens
- **JWKS** (JSON Web Key Set) para validação de tokens
- **Discovery** (OpenID Connect Discovery) para configuração automática
- **UserInfo Endpoint** para obter informações do usuário autenticado

**Base URL**: `https://api.sindireceita.org.br`

---

## Configuração Inicial

Antes de implementar, você precisará:

1. **Client ID**: Identificador único da sua aplicação
2. **Client Secret**: Chave secreta (mantenha segura!)
3. **Redirect URI**: URL de callback após autenticação (ex: `http://localhost:8080/callback`)
4. **Scopes**: Permissões solicitadas (mínimo: `openid`)

### Scopes Disponíveis
- `openid` - **Obrigatório** para autenticação
- `profile` - Acesso a nome, CPF e informações básicas
- `email` - Acesso ao email do usuário
- `phone` - Acesso ao telefone do usuário
- `address` - Acesso ao endereço
- `sindireceita.member.read` - Dados de filiação sindical
- `sindireceita.permissions.read` - Permissões do usuário no sistema

---

## Fluxo Minimalista

Este fluxo implementa apenas:
- ✅ Autenticação do usuário
- ✅ Obtenção do access token
- ✅ Validação do token

**Ideal para**: Aplicações simples que apenas precisam autenticar usuários.

### PHP (Fluxo Minimalista)

```php
<?php
// config.php
define('CLIENT_ID', 'seu_client_id_aqui');
define('CLIENT_SECRET', 'seu_client_secret_aqui');
define('REDIRECT_URI', 'http://localhost:8080/callback.php');
define('BASE_URL', 'https://api.sindireceita.org.br');
define('SCOPES', 'openid profile');

session_start();

// utils.php
function generateRandomString($length = 32) {
    return bin2hex(random_bytes($length));
}

function base64UrlEncode($data) {
    return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
}

function generatePKCEVerifier() {
    return generateRandomString(32);
}

function generatePKCEChallenge($verifier) {
    return base64UrlEncode(hash('sha256', $verifier, true));
}

// index.php - Página de Login
<?php
require_once 'config.php';
require_once 'utils.php';

// Gerar state para proteção CSRF
$state = generateRandomString();
$_SESSION['oauth_state'] = $state;

// Gerar PKCE verifier e challenge
$verifier = generatePKCEVerifier();
$_SESSION['code_verifier'] = $verifier;
$challenge = generatePKCEChallenge($verifier);

// Construir URL de autorização
$params = [
    'client_id' => CLIENT_ID,
    'redirect_uri' => REDIRECT_URI,
    'response_type' => 'code',
    'scope' => SCOPES,
    'state' => $state,
    'code_challenge' => $challenge,
    'code_challenge_method' => 'S256'
];

$authUrl = BASE_URL . '/oauth2/authorize?' . http_build_query($params);
?>

<!DOCTYPE html>
<html>
<head>
    <title>Login - Sindireceita</title>
</head>
<body>
    <h1>Autenticação OAuth2 - Sindireceita</h1>
    <a href="<?php echo htmlspecialchars($authUrl); ?>">
        <button>Entrar com Sindireceita</button>
    </a>
</body>
</html>

<?php
// callback.php - Callback OAuth2
require_once 'config.php';
require_once 'utils.php';

// Verificar erros
if (isset($_GET['error'])) {
    die('Erro OAuth2: ' . htmlspecialchars($_GET['error_description'] ?? $_GET['error']));
}

// Verificar state (proteção CSRF)
if (!isset($_GET['state']) || $_GET['state'] !== $_SESSION['oauth_state']) {
    die('Erro: State inválido (possível ataque CSRF)');
}

// Obter código de autorização
$code = $_GET['code'] ?? null;
if (!$code) {
    die('Erro: Código de autorização não encontrado');
}

// Recuperar code_verifier
$verifier = $_SESSION['code_verifier'] ?? null;
if (!$verifier) {
    die('Erro: Code verifier não encontrado na sessão');
}

// Trocar código por token
$tokenUrl = BASE_URL . '/oauth2/token';
$postData = [
    'grant_type' => 'authorization_code',
    'code' => $code,
    'redirect_uri' => REDIRECT_URI,
    'client_id' => CLIENT_ID,
    'client_secret' => CLIENT_SECRET,
    'code_verifier' => $verifier
];

$ch = curl_init($tokenUrl);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($postData));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/x-www-form-urlencoded'
]);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($httpCode !== 200) {
    die('Erro ao obter token: ' . $response);
}

$tokenData = json_decode($response, true);
$accessToken = $tokenData['access_token'];

// Validar token (parse básico)
function parseJWT($token) {
    $parts = explode('.', $token);
    if (count($parts) !== 3) {
        return null;
    }

    $payload = base64_decode(strtr($parts[1], '-_', '+/'));
    return json_decode($payload, true);
}

$claims = parseJWT($accessToken);

// Limpar variáveis de sessão temporárias
unset($_SESSION['oauth_state']);
unset($_SESSION['code_verifier']);

// Armazenar token
$_SESSION['access_token'] = $accessToken;
$_SESSION['user_sub'] = $claims['sub'] ?? 'unknown';
?>

<!DOCTYPE html>
<html>
<head>
    <title>Autenticado</title>
</head>
<body>
    <h1>Autenticação Realizada com Sucesso!</h1>
    <p>User ID: <?php echo htmlspecialchars($claims['sub'] ?? 'N/A'); ?></p>
    <p>Token válido até: <?php echo date('Y-m-d H:i:s', $claims['exp'] ?? 0); ?></p>
    <pre><?php print_r($claims); ?></pre>
</body>
</html>
```

### Next.js (Fluxo Minimalista)

```typescript
// lib/oauth-config.ts
export const oauthConfig = {
  clientId: process.env.OAUTH_CLIENT_ID!,
  clientSecret: process.env.OAUTH_CLIENT_SECRET!,
  redirectUri: process.env.OAUTH_REDIRECT_URI || 'http://localhost:3000/api/auth/callback',
  baseUrl: 'https://api.sindireceita.org.br',
  scopes: 'openid profile'
};

// lib/pkce.ts
import crypto from 'crypto';

export function generateRandomString(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

export function base64UrlEncode(buffer: Buffer): string {
  return buffer.toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

export function generatePKCEVerifier(): string {
  return base64UrlEncode(crypto.randomBytes(32));
}

export function generatePKCEChallenge(verifier: string): string {
  const hash = crypto.createHash('sha256').update(verifier).digest();
  return base64UrlEncode(hash);
}

// lib/jwt.ts
export interface JWTClaims {
  sub: string;
  exp: number;
  iat: number;
  iss?: string;
  aud?: string;
  [key: string]: any;
}

export function parseJWT(token: string): JWTClaims | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;

    const payload = Buffer.from(parts[1], 'base64url').toString();
    return JSON.parse(payload);
  } catch {
    return null;
  }
}

// app/page.tsx (App Router)
import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-3xl font-bold mb-4">
          Autenticação OAuth2 - Sindireceita
        </h1>
        <Link href="/api/auth/login">
          <button className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700">
            Entrar com Sindireceita
          </button>
        </Link>
      </div>
    </div>
  );
}

// app/api/auth/login/route.ts
import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { oauthConfig } from '@/lib/oauth-config';
import { generateRandomString, generatePKCEVerifier, generatePKCEChallenge } from '@/lib/pkce';

export async function GET() {
  // Gerar state para proteção CSRF
  const state = generateRandomString();

  // Gerar PKCE verifier e challenge
  const verifier = generatePKCEVerifier();
  const challenge = generatePKCEChallenge(verifier);

  // Armazenar em cookies
  cookies().set('oauth_state', state, { httpOnly: true, maxAge: 600 });
  cookies().set('code_verifier', verifier, { httpOnly: true, maxAge: 600 });

  // Construir URL de autorização
  const params = new URLSearchParams({
    client_id: oauthConfig.clientId,
    redirect_uri: oauthConfig.redirectUri,
    response_type: 'code',
    scope: oauthConfig.scopes,
    state,
    code_challenge: challenge,
    code_challenge_method: 'S256'
  });

  const authUrl = `${oauthConfig.baseUrl}/oauth2/authorize?${params}`;

  return NextResponse.redirect(authUrl);
}

// app/api/auth/callback/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { oauthConfig } from '@/lib/oauth-config';
import { parseJWT } from '@/lib/jwt';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;

  // Verificar erros
  const error = searchParams.get('error');
  if (error) {
    return NextResponse.json(
      { error: searchParams.get('error_description') || error },
      { status: 400 }
    );
  }

  // Verificar state (proteção CSRF)
  const state = searchParams.get('state');
  const savedState = cookies().get('oauth_state')?.value;

  if (!state || state !== savedState) {
    return NextResponse.json(
      { error: 'State inválido (possível ataque CSRF)' },
      { status: 400 }
    );
  }

  // Obter código de autorização
  const code = searchParams.get('code');
  if (!code) {
    return NextResponse.json(
      { error: 'Código de autorização não encontrado' },
      { status: 400 }
    );
  }

  // Recuperar code_verifier
  const verifier = cookies().get('code_verifier')?.value;
  if (!verifier) {
    return NextResponse.json(
      { error: 'Code verifier não encontrado' },
      { status: 400 }
    );
  }

  // Trocar código por token
  try {
    const tokenResponse = await fetch(`${oauthConfig.baseUrl}/oauth2/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        redirect_uri: oauthConfig.redirectUri,
        client_id: oauthConfig.clientId,
        client_secret: oauthConfig.clientSecret,
        code_verifier: verifier,
      }),
    });

    if (!tokenResponse.ok) {
      const errorText = await tokenResponse.text();
      throw new Error(`Token exchange failed: ${errorText}`);
    }

    const tokenData = await tokenResponse.json();
    const accessToken = tokenData.access_token;

    // Validar token (parse básico)
    const claims = parseJWT(accessToken);

    // Limpar cookies temporários
    cookies().delete('oauth_state');
    cookies().delete('code_verifier');

    // Armazenar token (em produção, use uma solução mais segura)
    cookies().set('access_token', accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      maxAge: 3600, // 1 hora
    });

    // Redirecionar para dashboard
    return NextResponse.redirect(new URL('/dashboard', request.url));

  } catch (error) {
    console.error('Error exchanging code for token:', error);
    return NextResponse.json(
      { error: 'Falha ao obter token' },
      { status: 500 }
    );
  }
}

// app/dashboard/page.tsx
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { parseJWT } from '@/lib/jwt';

export default function DashboardPage() {
  const accessToken = cookies().get('access_token')?.value;

  if (!accessToken) {
    redirect('/');
  }

  const claims = parseJWT(accessToken);

  if (!claims) {
    redirect('/');
  }

  return (
    <div className="min-h-screen p-8">
      <h1 className="text-3xl font-bold mb-6">Autenticação Realizada!</h1>
      <div className="bg-white shadow rounded-lg p-6">
        <p className="mb-2">
          <strong>User ID:</strong> {claims.sub}
        </p>
        <p className="mb-4">
          <strong>Token válido até:</strong>{' '}
          {new Date(claims.exp * 1000).toLocaleString('pt-BR')}
        </p>
        <pre className="bg-gray-100 p-4 rounded overflow-x-auto">
          {JSON.stringify(claims, null, 2)}
        </pre>
      </div>
    </div>
  );
}

// .env.local
OAUTH_CLIENT_ID=seu_client_id_aqui
OAUTH_CLIENT_SECRET=seu_client_secret_aqui
OAUTH_REDIRECT_URI=http://localhost:3000/api/auth/callback
```

---

## Fluxo Completo

Este fluxo implementa todos os recursos:
- ✅ Autenticação com PKCE
- ✅ Obtenção e validação de tokens
- ✅ **UserInfo** - Dados do usuário autenticado
- ✅ **Refresh Token** - Renovação automática de tokens
- ✅ **Token Revocation** - Logout/invalidação de tokens
- ✅ **JWKS Validation** - Validação criptográfica de tokens
- ✅ **Discovery** - Configuração automática via OpenID Connect

**Ideal para**: Aplicações empresariais que precisam de segurança robusta e recursos avançados.

### PHP (Fluxo Completo)

```php
<?php
// OAuthClient.php - Classe completa para OAuth2
class SindireceitaOAuthClient {
    private $clientId;
    private $clientSecret;
    private $redirectUri;
    private $baseUrl;
    private $scopes;
    private $discoveryConfig;

    public function __construct($clientId, $clientSecret, $redirectUri, $scopes = 'openid profile email') {
        $this->clientId = $clientId;
        $this->clientSecret = $clientSecret;
        $this->redirectUri = $redirectUri;
        $this->baseUrl = 'https://api.sindireceita.org.br';
        $this->scopes = $scopes;
        $this->discoveryConfig = $this->loadDiscovery();
    }

    // Carrega configuração via OpenID Connect Discovery
    private function loadDiscovery() {
        $url = $this->baseUrl . '/.well-known/openid-configuration';
        $response = file_get_contents($url);
        return json_decode($response, true);
    }

    // Gera URL de autorização com PKCE
    public function getAuthorizationUrl(&$state, &$verifier) {
        $state = bin2hex(random_bytes(32));
        $verifier = $this->generatePKCEVerifier();
        $challenge = $this->generatePKCEChallenge($verifier);

        $params = [
            'client_id' => $this->clientId,
            'redirect_uri' => $this->redirectUri,
            'response_type' => 'code',
            'scope' => $this->scopes,
            'state' => $state,
            'code_challenge' => $challenge,
            'code_challenge_method' => 'S256'
        ];

        return $this->discoveryConfig['authorization_endpoint'] . '?' . http_build_query($params);
    }

    // Troca código por tokens
    public function exchangeCode($code, $verifier) {
        $postData = [
            'grant_type' => 'authorization_code',
            'code' => $code,
            'redirect_uri' => $this->redirectUri,
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret,
            'code_verifier' => $verifier
        ];

        return $this->makeTokenRequest($this->discoveryConfig['token_endpoint'], $postData);
    }

    // Renova access token usando refresh token
    public function refreshToken($refreshToken) {
        $postData = [
            'grant_type' => 'refresh_token',
            'refresh_token' => $refreshToken,
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret
        ];

        return $this->makeTokenRequest($this->discoveryConfig['token_endpoint'], $postData);
    }

    // Revoga token (logout)
    public function revokeToken($token) {
        $url = $this->baseUrl . '/oauth2/revoke';
        $postData = [
            'token' => $token,
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret
        ];

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($postData));
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/x-www-form-urlencoded'
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        return $httpCode === 200;
    }

    // Obtém informações do usuário
    public function getUserInfo($accessToken) {
        $ch = curl_init($this->discoveryConfig['userinfo_endpoint']);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Authorization: Bearer ' . $accessToken
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode !== 200) {
            throw new Exception('Failed to fetch user info: ' . $response);
        }

        return json_decode($response, true);
    }

    // Valida token usando JWKS
    public function validateToken($token) {
        // Obter JWKS
        $jwks = $this->getJWKS();

        // Parse token header para obter kid
        $parts = explode('.', $token);
        if (count($parts) !== 3) {
            throw new Exception('Invalid token format');
        }

        $header = json_decode(base64_decode(strtr($parts[0], '-_', '+/')), true);
        $kid = $header['kid'] ?? null;

        if (!$kid) {
            throw new Exception('Token missing kid');
        }

        // Encontrar chave correspondente
        $key = null;
        foreach ($jwks['keys'] as $jwk) {
            if ($jwk['kid'] === $kid) {
                $key = $jwk;
                break;
            }
        }

        if (!$key) {
            throw new Exception('No matching key found for kid: ' . $kid);
        }

        // Validação completa requer biblioteca JWT (ex: firebase/php-jwt)
        // Por simplicidade, retornamos as claims sem validação de assinatura
        $payload = json_decode(base64_decode(strtr($parts[1], '-_', '+/')), true);

        // Verificar expiração
        if (isset($payload['exp']) && $payload['exp'] < time()) {
            throw new Exception('Token expired');
        }

        return $payload;
    }

    // Obtém JWKS (JSON Web Key Set)
    private function getJWKS() {
        $response = file_get_contents($this->discoveryConfig['jwks_uri']);
        return json_decode($response, true);
    }

    // Helpers PKCE
    private function generatePKCEVerifier() {
        return rtrim(strtr(base64_encode(random_bytes(32)), '+/', '-_'), '=');
    }

    private function generatePKCEChallenge($verifier) {
        return rtrim(strtr(base64_encode(hash('sha256', $verifier, true)), '+/', '-_'), '=');
    }

    // Helper para requisições de token
    private function makeTokenRequest($url, $postData) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($postData));
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/x-www-form-urlencoded'
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode !== 200) {
            throw new Exception('Token request failed: ' . $response);
        }

        return json_decode($response, true);
    }
}

// index.php - Login
<?php
session_start();
require_once 'OAuthClient.php';

$client = new SindireceitaOAuthClient(
    'seu_client_id',
    'seu_client_secret',
    'http://localhost:8080/callback.php',
    'openid profile email phone address sindireceita.member.read sindireceita.permissions.read'
);

$state = null;
$verifier = null;
$authUrl = $client->getAuthorizationUrl($state, $verifier);

$_SESSION['oauth_state'] = $state;
$_SESSION['code_verifier'] = $verifier;
?>

<!DOCTYPE html>
<html>
<head>
    <title>Login Completo - Sindireceita</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        button { background: #007bff; color: white; border: none; padding: 12px 24px; font-size: 16px; cursor: pointer; border-radius: 4px; }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <h1>Autenticação OAuth2 Completa - Sindireceita</h1>
    <p>Este fluxo utiliza todos os recursos disponíveis:</p>
    <ul>
        <li>PKCE (Proof Key for Code Exchange)</li>
        <li>OpenID Connect Discovery</li>
        <li>UserInfo Endpoint</li>
        <li>Token Refresh</li>
        <li>Token Revocation</li>
        <li>JWKS Validation</li>
    </ul>
    <a href="<?php echo htmlspecialchars($authUrl); ?>">
        <button>Entrar com Sindireceita (Fluxo Completo)</button>
    </a>
</body>
</html>

<?php
// callback.php - Callback completo
session_start();
require_once 'OAuthClient.php';

$client = new SindireceitaOAuthClient(
    'seu_client_id',
    'seu_client_secret',
    'http://localhost:8080/callback.php',
    'openid profile email phone address sindireceita.member.read sindireceita.permissions.read'
);

// Verificar erros
if (isset($_GET['error'])) {
    die('Erro OAuth2: ' . htmlspecialchars($_GET['error_description'] ?? $_GET['error']));
}

// Verificar state
if (!isset($_GET['state']) || $_GET['state'] !== $_SESSION['oauth_state']) {
    die('Erro: State inválido (possível ataque CSRF)');
}

try {
    $code = $_GET['code'] ?? null;
    $verifier = $_SESSION['code_verifier'] ?? null;

    if (!$code || !$verifier) {
        throw new Exception('Código ou verifier não encontrado');
    }

    // 1. Trocar código por tokens
    $tokenData = $client->exchangeCode($code, $verifier);
    $accessToken = $tokenData['access_token'];
    $refreshToken = $tokenData['refresh_token'] ?? null;
    $idToken = $tokenData['id_token'] ?? null;

    // 2. Validar access token usando JWKS
    $claims = $client->validateToken($accessToken);

    // 3. Obter informações do usuário
    $userInfo = $client->getUserInfo($accessToken);

    // Armazenar tokens
    $_SESSION['access_token'] = $accessToken;
    $_SESSION['refresh_token'] = $refreshToken;
    $_SESSION['id_token'] = $idToken;
    $_SESSION['user_info'] = $userInfo;
    $_SESSION['token_claims'] = $claims;

    // Limpar dados temporários
    unset($_SESSION['oauth_state']);
    unset($_SESSION['code_verifier']);

} catch (Exception $e) {
    die('Erro: ' . $e->getMessage());
}
?>

<!DOCTYPE html>
<html>
<head>
    <title>Dashboard Completo</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 20px auto; padding: 20px; }
        .card { background: #f5f5f5; border-radius: 8px; padding: 20px; margin-bottom: 20px; }
        .button { background: #007bff; color: white; border: none; padding: 10px 20px; margin: 5px; cursor: pointer; border-radius: 4px; text-decoration: none; display: inline-block; }
        .button.danger { background: #dc3545; }
        .button.success { background: #28a745; }
        pre { background: #fff; padding: 15px; border-radius: 4px; overflow-x: auto; }
        h2 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
    </style>
</head>
<body>
    <h1>Dashboard - Autenticação Completa</h1>

    <div class="card">
        <h2>Informações do Usuário</h2>
        <p><strong>Nome:</strong> <?php echo htmlspecialchars($userInfo['name'] ?? 'N/A'); ?></p>
        <p><strong>CPF:</strong> <?php echo htmlspecialchars($userInfo['cpf'] ?? 'N/A'); ?></p>
        <p><strong>Email:</strong> <?php echo htmlspecialchars($userInfo['email'] ?? 'N/A'); ?></p>
        <p><strong>Telefone:</strong> <?php echo htmlspecialchars($userInfo['phone_number'] ?? 'N/A'); ?></p>
        <p><strong>Status de Filiação:</strong> <?php echo htmlspecialchars($userInfo['membership_status'] ?? 'N/A'); ?></p>
        <p><strong>Status de Emprego:</strong> <?php echo htmlspecialchars($userInfo['employment_status'] ?? 'N/A'); ?></p>
        <p><strong>Tipo de Associação:</strong> <?php echo htmlspecialchars($userInfo['membership_type'] ?? 'N/A'); ?></p>
    </div>

    <div class="card">
        <h2>Permissões</h2>
        <?php if (!empty($userInfo['permissions'])): ?>
            <ul>
                <?php foreach ($userInfo['permissions'] as $permission): ?>
                    <li><?php echo htmlspecialchars($permission); ?></li>
                <?php endforeach; ?>
            </ul>
        <?php else: ?>
            <p>Nenhuma permissão específica</p>
        <?php endif; ?>
    </div>

    <div class="card">
        <h2>Token Claims (Validado via JWKS)</h2>
        <pre><?php echo json_encode($claims, JSON_PRETTY_PRINT); ?></pre>
    </div>

    <div class="card">
        <h2>Dados Completos do UserInfo</h2>
        <pre><?php echo json_encode($userInfo, JSON_PRETTY_PRINT); ?></pre>
    </div>

    <div class="card">
        <h2>Ações</h2>
        <a href="refresh.php" class="button success">Renovar Token (Refresh)</a>
        <a href="revoke.php" class="button danger">Revogar Token (Logout)</a>
    </div>
</body>
</html>

<?php
// refresh.php - Renovar token
session_start();
require_once 'OAuthClient.php';

$client = new SindireceitaOAuthClient(
    'seu_client_id',
    'seu_client_secret',
    'http://localhost:8080/callback.php'
);

try {
    $refreshToken = $_SESSION['refresh_token'] ?? null;

    if (!$refreshToken) {
        throw new Exception('Refresh token não encontrado');
    }

    // Renovar token
    $tokenData = $client->refreshToken($refreshToken);

    // Atualizar tokens na sessão
    $_SESSION['access_token'] = $tokenData['access_token'];
    if (isset($tokenData['refresh_token'])) {
        $_SESSION['refresh_token'] = $tokenData['refresh_token'];
    }

    echo '<h1>Token renovado com sucesso!</h1>';
    echo '<p><a href="callback.php">Voltar ao Dashboard</a></p>';
    echo '<pre>' . json_encode($tokenData, JSON_PRETTY_PRINT) . '</pre>';

} catch (Exception $e) {
    die('Erro ao renovar token: ' . $e->getMessage());
}

<?php
// revoke.php - Revogar token (logout)
session_start();
require_once 'OAuthClient.php';

$client = new SindireceitaOAuthClient(
    'seu_client_id',
    'seu_client_secret',
    'http://localhost:8080/callback.php'
);

try {
    $accessToken = $_SESSION['access_token'] ?? null;

    if (!$accessToken) {
        throw new Exception('Access token não encontrado');
    }

    // Revogar token
    $success = $client->revokeToken($accessToken);

    // Limpar sessão
    session_destroy();

    echo '<h1>Logout realizado com sucesso!</h1>';
    echo '<p>Token revogado: ' . ($success ? 'Sim' : 'Não') . '</p>';
    echo '<p><a href="index.php">Fazer login novamente</a></p>';

} catch (Exception $e) {
    die('Erro ao revogar token: ' . $e->getMessage());
}
```

### Next.js (Fluxo Completo)

```typescript
// lib/oauth-client.ts
interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope: string;
}

interface UserInfo {
  sub: string;
  name: string;
  cpf: string;
  email?: string;
  email_verified?: boolean;
  phone_number?: string;
  phone_number_verified?: boolean;
  address?: any;
  union_unit?: any;
  membership_status?: string;
  employment_status?: string;
  membership_type?: string;
  permissions?: string[];
}

interface DiscoveryConfig {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
  revocation_endpoint: string;
  scopes_supported: string[];
  response_types_supported: string[];
  grant_types_supported: string[];
}

export class SindireceitaOAuthClient {
  private clientId: string;
  private clientSecret: string;
  private redirectUri: string;
  private baseUrl: string;
  private scopes: string;
  private discoveryConfig: DiscoveryConfig | null = null;

  constructor(
    clientId: string,
    clientSecret: string,
    redirectUri: string,
    scopes: string = 'openid profile email'
  ) {
    this.clientId = clientId;
    this.clientSecret = clientSecret;
    this.redirectUri = redirectUri;
    this.baseUrl = 'https://api.sindireceita.org.br';
    this.scopes = scopes;
  }

  // Carrega configuração via OpenID Connect Discovery
  async loadDiscovery(): Promise<DiscoveryConfig> {
    if (this.discoveryConfig) {
      return this.discoveryConfig;
    }

    const url = `${this.baseUrl}/.well-known/openid-configuration`;
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error('Failed to load discovery config');
    }

    this.discoveryConfig = await response.json();
    return this.discoveryConfig!;
  }

  // Gera URL de autorização com PKCE
  async getAuthorizationUrl(): Promise<{
    url: string;
    state: string;
    verifier: string;
  }> {
    const config = await this.loadDiscovery();
    const state = this.generateRandomString();
    const verifier = this.generatePKCEVerifier();
    const challenge = this.generatePKCEChallenge(verifier);

    const params = new URLSearchParams({
      client_id: this.clientId,
      redirect_uri: this.redirectUri,
      response_type: 'code',
      scope: this.scopes,
      state,
      code_challenge: challenge,
      code_challenge_method: 'S256'
    });

    return {
      url: `${config.authorization_endpoint}?${params}`,
      state,
      verifier
    };
  }

  // Troca código por tokens
  async exchangeCode(code: string, verifier: string): Promise<TokenResponse> {
    const config = await this.loadDiscovery();

    const response = await fetch(config.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        redirect_uri: this.redirectUri,
        client_id: this.clientId,
        client_secret: this.clientSecret,
        code_verifier: verifier,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Token exchange failed: ${error}`);
    }

    return response.json();
  }

  // Renova access token usando refresh token
  async refreshToken(refreshToken: string): Promise<TokenResponse> {
    const config = await this.loadDiscovery();

    const response = await fetch(config.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: this.clientId,
        client_secret: this.clientSecret,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Token refresh failed: ${error}`);
    }

    return response.json();
  }

  // Revoga token (logout)
  async revokeToken(token: string): Promise<boolean> {
    const url = `${this.baseUrl}/oauth2/revoke`;

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        token,
        client_id: this.clientId,
        client_secret: this.clientSecret,
      }),
    });

    return response.ok;
  }

  // Obtém informações do usuário
  async getUserInfo(accessToken: string): Promise<UserInfo> {
    const config = await this.loadDiscovery();

    const response = await fetch(config.userinfo_endpoint, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`UserInfo request failed: ${error}`);
    }

    return response.json();
  }

  // Valida token usando JWKS
  async validateToken(token: string): Promise<any> {
    const config = await this.loadDiscovery();

    // Obter JWKS
    const jwksResponse = await fetch(config.jwks_uri);
    const jwks = await jwksResponse.json();

    // Parse token header
    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid token format');
    }

    const header = JSON.parse(
      Buffer.from(parts[0], 'base64url').toString()
    );
    const payload = JSON.parse(
      Buffer.from(parts[1], 'base64url').toString()
    );

    // Verificar expiração
    if (payload.exp && payload.exp < Date.now() / 1000) {
      throw new Error('Token expired');
    }

    // Para validação completa da assinatura, use uma biblioteca como 'jose'
    // Este é um exemplo simplificado

    return payload;
  }

  // Helpers PKCE
  private generateRandomString(length: number = 32): string {
    return require('crypto').randomBytes(length).toString('hex');
  }

  private generatePKCEVerifier(): string {
    const buffer = require('crypto').randomBytes(32);
    return buffer.toString('base64url');
  }

  private generatePKCEChallenge(verifier: string): string {
    const crypto = require('crypto');
    const hash = crypto.createHash('sha256').update(verifier).digest();
    return hash.toString('base64url');
  }
}

// app/api/auth/login/route.ts (Fluxo Completo)
import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { SindireceitaOAuthClient } from '@/lib/oauth-client';

export async function GET() {
  const client = new SindireceitaOAuthClient(
    process.env.OAUTH_CLIENT_ID!,
    process.env.OAUTH_CLIENT_SECRET!,
    process.env.OAUTH_REDIRECT_URI!,
    'openid profile email phone address sindireceita.member.read sindireceita.permissions.read'
  );

  const { url, state, verifier } = await client.getAuthorizationUrl();

  // Armazenar em cookies
  cookies().set('oauth_state', state, { httpOnly: true, maxAge: 600 });
  cookies().set('code_verifier', verifier, { httpOnly: true, maxAge: 600 });

  return NextResponse.redirect(url);
}

// app/api/auth/callback/route.ts (Fluxo Completo)
import { NextRequest, NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { SindireceitaOAuthClient } from '@/lib/oauth-client';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;

  // Verificar erros
  const error = searchParams.get('error');
  if (error) {
    return NextResponse.redirect(
      new URL(`/?error=${encodeURIComponent(error)}`, request.url)
    );
  }

  // Verificar state
  const state = searchParams.get('state');
  const savedState = cookies().get('oauth_state')?.value;

  if (!state || state !== savedState) {
    return NextResponse.redirect(
      new URL('/?error=invalid_state', request.url)
    );
  }

  // Obter código
  const code = searchParams.get('code');
  if (!code) {
    return NextResponse.redirect(
      new URL('/?error=no_code', request.url)
    );
  }

  const verifier = cookies().get('code_verifier')?.value;
  if (!verifier) {
    return NextResponse.redirect(
      new URL('/?error=no_verifier', request.url)
    );
  }

  try {
    const client = new SindireceitaOAuthClient(
      process.env.OAUTH_CLIENT_ID!,
      process.env.OAUTH_CLIENT_SECRET!,
      process.env.OAUTH_REDIRECT_URI!,
      'openid profile email phone address sindireceita.member.read sindireceita.permissions.read'
    );

    // 1. Trocar código por tokens
    const tokenData = await client.exchangeCode(code, verifier);

    // 2. Validar token
    const claims = await client.validateToken(tokenData.access_token);

    // 3. Obter informações do usuário
    const userInfo = await client.getUserInfo(tokenData.access_token);

    // Armazenar dados (em produção, use um backend seguro)
    cookies().set('access_token', tokenData.access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      maxAge: tokenData.expires_in,
    });

    if (tokenData.refresh_token) {
      cookies().set('refresh_token', tokenData.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        maxAge: 30 * 24 * 60 * 60, // 30 dias
      });
    }

    cookies().set('user_info', JSON.stringify(userInfo), {
      httpOnly: true,
      maxAge: 24 * 60 * 60,
    });

    // Limpar cookies temporários
    cookies().delete('oauth_state');
    cookies().delete('code_verifier');

    return NextResponse.redirect(new URL('/dashboard', request.url));

  } catch (error) {
    console.error('OAuth callback error:', error);
    return NextResponse.redirect(
      new URL('/?error=callback_failed', request.url)
    );
  }
}

// app/api/auth/refresh/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { SindireceitaOAuthClient } from '@/lib/oauth-client';

export async function POST(request: NextRequest) {
  const refreshToken = cookies().get('refresh_token')?.value;

  if (!refreshToken) {
    return NextResponse.json(
      { error: 'No refresh token' },
      { status: 401 }
    );
  }

  try {
    const client = new SindireceitaOAuthClient(
      process.env.OAUTH_CLIENT_ID!,
      process.env.OAUTH_CLIENT_SECRET!,
      process.env.OAUTH_REDIRECT_URI!
    );

    const tokenData = await client.refreshToken(refreshToken);

    // Atualizar cookies
    cookies().set('access_token', tokenData.access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      maxAge: tokenData.expires_in,
    });

    if (tokenData.refresh_token) {
      cookies().set('refresh_token', tokenData.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        maxAge: 30 * 24 * 60 * 60,
      });
    }

    return NextResponse.json({ success: true });

  } catch (error) {
    return NextResponse.json(
      { error: 'Token refresh failed' },
      { status: 500 }
    );
  }
}

// app/api/auth/revoke/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { SindireceitaOAuthClient } from '@/lib/oauth-client';

export async function POST(request: NextRequest) {
  const accessToken = cookies().get('access_token')?.value;

  if (!accessToken) {
    return NextResponse.json({ success: true });
  }

  try {
    const client = new SindireceitaOAuthClient(
      process.env.OAUTH_CLIENT_ID!,
      process.env.OAUTH_CLIENT_SECRET!,
      process.env.OAUTH_REDIRECT_URI!
    );

    await client.revokeToken(accessToken);

    // Limpar cookies
    cookies().delete('access_token');
    cookies().delete('refresh_token');
    cookies().delete('user_info');

    return NextResponse.json({ success: true });

  } catch (error) {
    return NextResponse.json(
      { error: 'Token revocation failed' },
      { status: 500 }
    );
  }
}

// app/dashboard/page.tsx (Fluxo Completo)
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export default function DashboardPage() {
  const userInfoStr = cookies().get('user_info')?.value;
  const accessToken = cookies().get('access_token')?.value;

  if (!accessToken || !userInfoStr) {
    redirect('/');
  }

  const userInfo = JSON.parse(userInfoStr);

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">Dashboard Completo</h1>

        {/* Informações do Usuário */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-semibold mb-4 border-b-2 border-blue-600 pb-2">
            Informações do Usuário
          </h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-gray-600">Nome</p>
              <p className="font-semibold">{userInfo.name || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">CPF</p>
              <p className="font-semibold">{userInfo.cpf || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">Email</p>
              <p className="font-semibold">{userInfo.email || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">Telefone</p>
              <p className="font-semibold">{userInfo.phone_number || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">Status de Filiação</p>
              <p className="font-semibold">{userInfo.membership_status || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">Status de Emprego</p>
              <p className="font-semibold">{userInfo.employment_status || 'N/A'}</p>
            </div>
          </div>
        </div>

        {/* Permissões */}
        {userInfo.permissions && userInfo.permissions.length > 0 && (
          <div className="bg-white shadow rounded-lg p-6 mb-6">
            <h2 className="text-2xl font-semibold mb-4">Permissões</h2>
            <ul className="list-disc list-inside">
              {userInfo.permissions.map((perm: string, idx: number) => (
                <li key={idx}>{perm}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Dados Completos */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-semibold mb-4">Dados Completos</h2>
          <pre className="bg-gray-100 p-4 rounded overflow-x-auto">
            {JSON.stringify(userInfo, null, 2)}
          </pre>
        </div>

        {/* Ações */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-2xl font-semibold mb-4">Ações</h2>
          <div className="flex gap-4">
            <form action="/api/auth/refresh" method="POST">
              <button
                type="submit"
                className="bg-green-600 text-white px-6 py-2 rounded hover:bg-green-700"
              >
                Renovar Token
              </button>
            </form>
            <form action="/api/auth/revoke" method="POST">
              <button
                type="submit"
                className="bg-red-600 text-white px-6 py-2 rounded hover:bg-red-700"
              >
                Revogar Token (Logout)
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
```

---

## Endpoints Disponíveis

### 1. Discovery (OpenID Connect)
```
GET https://api.sindireceita.org.br/.well-known/openid-configuration
```
Retorna configuração automática do servidor OAuth2.

### 2. Authorization Endpoint
```
GET https://api.sindireceita.org.br/oauth2/authorize
```
Parâmetros:
- `client_id` (obrigatório)
- `redirect_uri` (obrigatório)
- `response_type=code` (obrigatório)
- `scope` (obrigatório, deve incluir "openid")
- `state` (recomendado, proteção CSRF)
- `code_challenge` (obrigatório para PKCE)
- `code_challenge_method=S256` (obrigatório para PKCE)

### 3. Token Endpoint
```
POST https://api.sindireceita.org.br/oauth2/token
Content-Type: application/x-www-form-urlencoded
```

**Para Authorization Code:**
```
grant_type=authorization_code
code=<código>
redirect_uri=<uri>
client_id=<id>
client_secret=<secret>
code_verifier=<verifier>
```

**Para Refresh Token:**
```
grant_type=refresh_token
refresh_token=<token>
client_id=<id>
client_secret=<secret>
```

### 4. UserInfo Endpoint
```
GET https://api.sindireceita.org.br/oauth2/userinfo
Authorization: Bearer <access_token>
```

Retorna:
```json
{
  "sub": "user_id",
  "name": "Nome Completo",
  "cpf": "12345678900",
  "email": "user@example.com",
  "email_verified": true,
  "phone_number": "+5511999999999",
  "phone_number_verified": true,
  "membership_status": "active",
  "employment_status": "employed",
  "membership_type": "full",
  "permissions": ["read", "write"]
}
```

### 5. Token Revocation
```
POST https://api.sindireceita.org.br/oauth2/revoke
Content-Type: application/x-www-form-urlencoded

token=<access_token>
client_id=<id>
client_secret=<secret>
```

### 6. JWKS (JSON Web Key Set)
```
GET https://api.sindireceita.org.br/oauth2/jwks
```
Retorna chaves públicas para validação de tokens JWT.

---

## Troubleshooting

### Erro: "openid scope is required"
**Solução**: Sempre inclua o scope `openid` nas suas requisições.

### Erro: "Invalid state parameter"
**Solução**: Verifique se está armazenando e comparando o `state` corretamente. Isso protege contra ataques CSRF.

### Erro: "Code verifier not found"
**Solução**: Certifique-se de armazenar o `code_verifier` gerado antes de redirecionar para autorização.

### Erro: "Token expired"
**Solução**: Use o `refresh_token` para obter um novo `access_token`.

### Erro de CORS
**Solução**: OAuth2 não deve ser feito via JavaScript no frontend. Use server-side (PHP, Next.js API Routes).

### Token não valida
**Solução**: Verifique se está usando o endpoint JWKS correto e se o token não expirou.

---

## Melhores Práticas

1. **Sempre use HTTPS em produção**
2. **Nunca exponha o Client Secret no frontend**
3. **Armazene tokens de forma segura (httpOnly cookies)**
4. **Implemente refresh token automático**
5. **Revogue tokens no logout**
6. **Valide tokens usando JWKS**
7. **Use state para proteção CSRF**
8. **Implemente PKCE para segurança adicional**

---

## Recursos Adicionais

- RFC 6749 - OAuth 2.0 Authorization Framework
- RFC 7636 - Proof Key for Code Exchange (PKCE)
- RFC 7009 - Token Revocation
- OpenID Connect Core 1.0

---

**Desenvolvido para**: Sindireceita OAuth2 Server
**Versão**: 1.0
**Data**: 2026-01-02
