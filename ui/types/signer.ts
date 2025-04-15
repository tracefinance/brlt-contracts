import { fromJson, fromJsonArray, toJson } from './model';
import type { IPagedResponse } from './model'; // Re-add IPagedResponse import as it's needed implicitly

/**
 * Defines the possible types for a signer.
 */
export type ISignerType = 'internal' | 'external';

/**
 * Interface representing a signer address.
 */
export interface IAddress {
  id: number;
  signerId: number;
  chainType: string;
  address: string;
  createdAt: string;
  updatedAt: string;
}

export const Address = {
  /**
   * Converts a plain JSON object (snake_case) from the API to an IAddress (camelCase).
   * @param json The plain JSON object.
   * @returns An IAddress object.
   */
  fromJson(json: any): IAddress {
    return fromJson<IAddress>(json);
  },

  /**
   * Converts an array of plain JSON objects from the API to IAddress objects.
   * @param jsonArray The array of plain JSON objects.
   * @returns An array of IAddress objects.
   */
  fromJsonArray(jsonArray: any[]): IAddress[] {
    return fromJsonArray<IAddress>(jsonArray);
  }
};

/**
 * Interface representing a signer.
 */
export interface ISigner {
  id: number;
  name: string;
  type: ISignerType;
  userId?: number;
  addresses?: IAddress[];
  createdAt: string;
  updatedAt: string;
}

export const Signer = {
  fromJson(json: any): ISigner {
    const signer = fromJson<ISigner>(json);
    if (json.addresses) {
      signer.addresses = Address.fromJsonArray(json.addresses);
    }
    return signer;
  },

  fromJsonArray(jsonArray: any[]): ISigner[] {
    return jsonArray.map(json => Signer.fromJson(json));
  }
};

export interface ICreateSignerRequest {
  name: string;
  type: ISignerType;
  userId?: number;
}

export const CreateSignerRequest = {
  create(name: string, type: ISignerType, userId?: number): ICreateSignerRequest {
    return { name, type, userId };
  },
  toJson(request: ICreateSignerRequest): any {
    return toJson(request);
  }
};

export interface IUpdateSignerRequest {
  name: string;
  type: ISignerType;
  userId?: number;
}

export const UpdateSignerRequest = {
  create(name: string, type: ISignerType, userId?: number): IUpdateSignerRequest {
    return { name, type, userId };
  },
  toJson(request: IUpdateSignerRequest): any {
    return toJson(request);
  }
};

export interface IAddAddressRequest {
  chainType: string;
  address: string;
}

export const AddAddressRequest = {
  create(chainType: string, address: string): IAddAddressRequest {
    return { chainType, address };
  },
  toJson(request: IAddAddressRequest): any {
    return toJson(request);
  }
};
