"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { WalletFrontend, CreateWalletRequest, UpdateWalletRequest } from "@/types/wallet";

// Schema for creating a new wallet
const createFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  chain_type: z.string().min(1, "Chain type is required"),
});

// Schema for updating an existing wallet
const updateFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

type CreateFormValues = z.infer<typeof createFormSchema>;
type UpdateFormValues = z.infer<typeof updateFormSchema>;

interface WalletFormProps {
  wallet?: WalletFrontend;
  onSubmit: (data: any) => Promise<void>; // Using any to work around TypeScript issues
  onCancel: () => void;
}

export default function WalletForm({ wallet, onSubmit, onCancel }: WalletFormProps) {
  const isEditing = !!wallet;
  
  // Use the appropriate form schema based on whether we're editing or creating
  const form = useForm({
    resolver: zodResolver(isEditing ? updateFormSchema : createFormSchema),
    defaultValues: {
      name: wallet?.name || "",
      ...(isEditing ? {} : { chain_type: "" }),
    },
  });

  async function handleSubmit(values: any) {
    try {
      await onSubmit(values);
      form.reset();
    } catch (error) {
      console.error("Form submission error:", error);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Wallet Name</FormLabel>
              <FormControl>
                <Input placeholder="My Wallet" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        {!isEditing && (
          <FormField
            control={form.control}
            name="chain_type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Chain Type</FormLabel>
                <Select 
                  onValueChange={field.onChange} 
                  defaultValue={field.value}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select chain type" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="ETHEREUM">Ethereum</SelectItem>
                    <SelectItem value="POLYGON">Polygon</SelectItem>
                    <SelectItem value="ARBITRUM">Arbitrum</SelectItem>
                    <SelectItem value="OPTIMISM">Optimism</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
        
        <div className="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            {isEditing ? "Update" : "Create"}
          </Button>
        </div>
      </form>
    </Form>
  );
} 