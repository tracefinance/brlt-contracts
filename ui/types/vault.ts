import { fromJson, toJson } from './model';

// Status enum for vault
export type VaultStatus = 'pending' | 'active' | 'recovering' | 'recovered';

// Vault model
export interface IVault {
  id: number;
  name: string;
  contractName: string;
  walletId: number;
  chainType: string;
  recoveryAddress: string;
  signers: string[];
  address?: string;
  status: string;
  quorum: number;
  inRecovery: boolean;
  recoveryDeadline?: string;
  createdAt: string;
  updatedAt: string;
}

// Factory object for data transformation
export const Vault = {
  fromJson(json: any): IVault {
    return fromJson<IVault>(json);
  },
  
  fromJsonArray(jsonArray: any[]): IVault[] {
    return jsonArray.map(json => Vault.fromJson(json));
  }
};

// Request/Response types
export interface ICreateVaultRequest {
  name: string;
  recovery_address: string;
  signer_addresses: string[];
  signature_threshold: number;
  whitelisted_tokens?: string[];
}

export const CreateVaultRequest = {
  create(
    name: string,
    recoveryAddress: string,
    signerAddresses: string[],
    signatureThreshold: number,
    whitelistedTokens?: string[]
  ): ICreateVaultRequest {
    return {
      name,
      recovery_address: recoveryAddress,
      signer_addresses: signerAddresses,
      signature_threshold: signatureThreshold,
      whitelisted_tokens: whitelistedTokens
    };
  },
  toJson(request: ICreateVaultRequest): any {
    return toJson(request);
  }
};

export interface IUpdateVaultRequest {
  name: string;
}

export const UpdateVaultRequest = {
  create(name: string): IUpdateVaultRequest {
    return { name };
  },
  toJson(request: IUpdateVaultRequest): any {
    return toJson(request);
  }
};

export interface ITokenRequest {
  address: string;
}

export const TokenRequest = {
  create(address: string): ITokenRequest {
    return { address };
  },
  toJson(request: ITokenRequest): any {
    return toJson(request);
  }
};

export interface IVaultFilter {
  status?: string;
  address?: string;
}

// Response types
export interface ITokenActionResponse {
  vault_id: number;
  token_address: string;
  tx_hash: string;
}

export interface IRecoveryResponse {
  vault_id: number;
  status: string;
  action: string;
  tx_hash: string;
  recovery_initiated?: string;
  executable_after?: string;
} 