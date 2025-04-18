/* eslint-disable @typescript-eslint/no-explicit-any */
import { fromJson, fromJsonArray } from './model';
import type { IToken } from './token';

/**
 * Interface representing a wallet
 */
export interface IWallet {
  id: string;
  keyId: string;
  chainType: string;
  address: string;
  name: string;
  tags?: Record<string, string>;
  lastBlockNumber?: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Factory functions for IWallet
 */
export const Wallet = {
  /**
   * Converts a plain JSON object from the API to an IWallet
   */
  fromJson(json: any): IWallet {
    return fromJson<IWallet>(json);
  },

  /**
   * Converts an array of plain JSON objects from the API to IWallet objects
   */
  fromJsonArray(jsonArray: any[]): IWallet[] {
    return fromJsonArray<IWallet>(jsonArray);
  }
};

/**
 * Interface representing a token balance response
 */
export interface ITokenBalanceResponse {
  token: IToken;
  balance: string;
  updatedAt: string;
}

/**
 * Factory functions for ITokenBalanceResponse
 */
export const TokenBalanceResponse = {
  fromJson(json: any): ITokenBalanceResponse {
    const response = fromJson<ITokenBalanceResponse>(json);
    
    // Convert nested token
    if (json.token) {
      response.token = fromJson<IToken>(json.token);
    }
    
    return response;
  },

  fromJsonArray(jsonArray: any[]): ITokenBalanceResponse[] {
    return jsonArray.map(json => TokenBalanceResponse.fromJson(json));
  }
};

/**
 * Interface representing a request to create a wallet
 */
export interface ICreateWalletRequest {
  chainType: string;
  name: string;
  tags?: Record<string, string>;
}

/**
 * Factory functions for ICreateWalletRequest
 */
export const CreateWalletRequest = {
  create(chainType: string, name: string, tags?: Record<string, string>): ICreateWalletRequest {
    return { chainType, name, tags };
  }
};

/**
 * Interface representing a request to update a wallet
 */
export interface IUpdateWalletRequest {
  name: string;
  tags?: Record<string, string>;
}

/**
 * Factory functions for IUpdateWalletRequest
 */
export const UpdateWalletRequest = {
  create(name: string, tags?: Record<string, string>): IUpdateWalletRequest {
    return { name, tags };
  }
}; 