import { 
  plainToInstance, 
  instanceToPlain,
} from 'class-transformer';
import 'reflect-metadata';
import type { ClassConstructor } from 'class-transformer';

/**
 * Converts a plain JSON object to an instance of the specified class
 * @param cls The class constructor
 * @param json The plain JSON object
 * @returns An instance of the specified class
 */
export function fromJson<T>(cls: ClassConstructor<T>, json: any): T {
  return plainToInstance(cls, json, {
    excludeExtraneousValues: true
  });
}

/**
 * Converts an array of plain JSON objects to instances of the specified class
 * @param cls The class constructor
 * @param jsonArray The array of plain JSON objects
 * @returns An array of instances of the specified class
 */
export function fromJsonArray<T>(cls: ClassConstructor<T>, jsonArray: any[]): T[] {
  return jsonArray.map(json => fromJson(cls, json));
}

/**
 * Converts an instance to a plain JSON object
 * @param instance The instance to convert
 * @returns A plain JSON object
 */
export function toJson<T>(instance: T): any {
  return instanceToPlain(instance, {
    excludeExtraneousValues: true
  });
}

/**
 * Base model class that provides common JSON conversion methods
 */
export abstract class BaseModel {
  /**
   * Converts this instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return toJson(this);
  }

  /**
   * Creates a static fromJson method for the derived class
   * This is a helper factory method for model classes
   */
  static createFromJson<T extends BaseModel>(cls: ClassConstructor<T>) {
    return (json: any): T => fromJson(cls, json);
  }

  /**
   * Creates a static fromJsonArray method for the derived class
   * This is a helper factory method for model classes
   */
  static createFromJsonArray<T extends BaseModel>(cls: ClassConstructor<T>) {
    return (jsonArray: any[]): T[] => fromJsonArray(cls, jsonArray);
  }
} 