export type AuthMethod = 'local' | 'ldap' | 'oauth2' | 'azure-ad' | 'saml';

export type OAuthProvider = 'google' | 'github' | 'microsoft';

export interface AuthConfig {
  id: string;
  method: AuthMethod;
  isActive: boolean;
  
  // Config Local
  local?: {
    requireEmailVerification: boolean;
    minPasswordLength: number;
    maxLoginAttempts: number;
    lockoutDuration: number;
  };
  
  // Config LDAP
  ldap?: {
    url: string;
    port: number;
    baseDn: string;
    bindDn?: string;
    bindPassword?: string;
    userSearchFilter: string;
    groupSearchFilter?: string;
  };
  
  // Config OAuth2
  oauth2?: {
    provider: OAuthProvider;
    clientId: string;
    clientSecret: string;
    redirectUri: string;
    scopes: string[];
  };
  
  // Config Azure AD
  azureAd?: {
    tenantId: string;
    clientId: string;
    clientSecret: string;
    redirectUri: string;
    authority: string;
  };
  
  // Config SAML
  saml?: {
    entryPoint: string;
    issuer: string;
    redirectUrl: string;
    cert: string;
  };
}

export interface UpdateAuthConfigRequest {
  method: AuthMethod;
  isActive: boolean;
  local?: AuthConfig['local'];
  ldap?: AuthConfig['ldap'];
  oauth2?: AuthConfig['oauth2'];
  azureAd?: AuthConfig['azureAd'];
  saml?: AuthConfig['saml'];
}