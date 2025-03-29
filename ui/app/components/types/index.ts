export type Wallet = {
    name: string;
    address: string;
    chainType: string;
};

export type Token = {
    address: string;
    chainType: string;
    symbol: string;
    decimals: number;
};

export type TokenBalance = {
    token: Token;
    balance: string;
    updatedAt: string;
};