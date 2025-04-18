/* eslint-disable @typescript-eslint/no-explicit-any */
import { fromJson, fromJsonArray } from './model';

/**
 * Interface representing a user
 */
export interface IUser {
  id: string;
  email: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Factory functions for IUser
 */
export const User = {
  /**
   * Converts a plain JSON object from the API to an IUser
   */
  fromJson(json: any): IUser {
    return fromJson<IUser>(json);
  },

  /**
   * Converts an array of plain JSON objects from the API to IUser objects
   */
  fromJsonArray(jsonArray: any[]): IUser[] {
    return fromJsonArray<IUser>(jsonArray);
  }
};

/**
 * Interface representing a request to create a user
 */
export interface ICreateUserRequest {
  email: string;
  password: string;
}

/**
 * Factory functions for ICreateUserRequest
 */
export const CreateUserRequest = {
  create(email: string, password: string): ICreateUserRequest {
    return { email, password };
  }
};

/**
 * Interface representing a request to update a user
 */
export interface IUpdateUserRequest {
  email?: string;
  password?: string;
}

/**
 * Factory functions for IUpdateUserRequest
 */
export const UpdateUserRequest = {
  create(email?: string, password?: string): IUpdateUserRequest {
    return { email, password };
  }
}; 