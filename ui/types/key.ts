/* eslint-disable @typescript-eslint/no-explicit-any */
import { fromJson, fromJsonArray } from './model';

/**
 * Interface representing a cryptographic key
 */
export interface IKey {
  id: string;
  name: string;
  type: 'ecdsa' | 'rsa' | 'ed25519' | 'symmetric';
  curve?: string;
  createdAt: string;
  tags?: Record<string, string>;
  publicKey?: string;
}

/**
 * Factory functions for IKey
 */
export const Key = {
  /**
   * Converts a plain JSON object from the API to an IKey
   */
  fromJson(json: any): IKey {
    return fromJson<IKey>(json);
  },

  /**
   * Converts an array of plain JSON objects from the API to IKey objects
   */
  fromJsonArray(jsonArray: any[]): IKey[] {
    return fromJsonArray<IKey>(jsonArray);
  }
};

/**
 * Interface representing a request to create a key
 */
export interface ICreateKeyRequest {
  name: string;
  type: 'ecdsa' | 'rsa' | 'ed25519' | 'symmetric';
  curve?: string;
  tags?: Record<string, string>;
}

/**
 * Factory functions for ICreateKeyRequest
 */
export const CreateKeyRequest = {
  create(
    name: string,
    type: 'ecdsa' | 'rsa' | 'ed25519' | 'symmetric',
    curve?: string,
    tags?: Record<string, string>
  ): ICreateKeyRequest {
    return { name, type, curve, tags };
  }
};

/**
 * Interface representing a request to update a key
 */
export interface IUpdateKeyRequest {
  name?: string;
  tags?: Record<string, string>;
}

/**
 * Factory functions for IUpdateKeyRequest
 */
export const UpdateKeyRequest = {
  create(
    name?: string,
    tags?: Record<string, string>
  ): IUpdateKeyRequest {
    const request: IUpdateKeyRequest = {};
    if (name !== undefined) {
      request.name = name;
    }
    if (tags !== undefined) {
      request.tags = tags;
    }
    return request;
  }
};

/**
 * Interface representing a request to import an existing key
 */
export interface IImportKeyRequest {
  name: string;
  type: 'ecdsa' | 'rsa' | 'ed25519' | 'symmetric';
  curve?: string;
  privateKey: string;
  publicKey?: string;
  tags?: Record<string, string>;
}

/**
 * Factory functions for IImportKeyRequest
 */
export const ImportKeyRequest = {
  create(
    name: string,
    type: 'ecdsa' | 'rsa' | 'ed25519' | 'symmetric',
    privateKey: string,
    curve?: string,
    publicKey?: string,
    tags?: Record<string, string>
  ): IImportKeyRequest {
    return { name, type, curve, privateKey, publicKey, tags };
  }
};

/**
 * Interface representing a request to sign data
 */
export interface ISignDataRequest {
  data: string; // Base64 encoded
  rawData?: boolean;
}

/**
 * Interface representing the response for signing data
 */
export interface ISignDataResponse {
  signature: string; // Base64 encoded
}

/**
 * Interface representing a public key export
 */
export interface IPublicKeyExport {
  format: 'pem' | 'der' | 'jwk';
  publicKey: string;
  fingerprint: string;
}

// Key Types (matches backend internal/types/crypto.go)
export const KEY_TYPES = {
  ECDSA: 'ecdsa',
  RSA: 'rsa',
  ED25519: 'ed25519',
  SYMMETRIC: 'symmetric'
};

// Curve Names (matches backend internal/types/crypto.go where defined)
export const CURVE_NAMES = {
  P256: 'P-256',       
  SECP256K1: 'secp256k1'
};
