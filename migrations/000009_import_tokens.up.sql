-- Import initial tokens from configuration

-- Ethereum tokens
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x0000000000000000000000000000000000000000', 'ethereum', 'ETH', 18, 'native');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599', 'ethereum', 'WBTC', 8, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', 'ethereum', 'WETH', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', 'ethereum', 'USDC', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xdAC17F958D2ee523a2206206994597C13D831ec7', 'ethereum', 'USDT', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x6B175474E89094C44Da98b954EedeAC495271d0F', 'ethereum', 'DAI', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x514910771AF9Ca656af840dff83E8264EcF986CA', 'ethereum', 'LINK', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984', 'ethereum', 'UNI', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0', 'ethereum', 'MATIC', 18, 'erc20');

-- Polygon tokens
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x0000000000000000000000000000000000000000', 'polygon', 'MATIC', 18, 'native');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270', 'polygon', 'WMATIC', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x1BFD67037B42Cf73acF2047067bd4F2C47D9BfD6', 'polygon', 'WBTC', 8, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619', 'polygon', 'WETH', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174', 'polygon', 'USDC', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xc2132D05D31c914a87C6611C10748AEb04B58e8F', 'polygon', 'USDT', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063', 'polygon', 'DAI', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xb0897686c545045aFc77CF20eC7A532E3120E0F1', 'polygon', 'LINK', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xb33EaAd8d922B1083446DC23f610c2567fB5180f', 'polygon', 'UNI', 18, 'erc20');

-- Base tokens
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x0000000000000000000000000000000000000000', 'base', 'ETH', 18, 'native');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x4200000000000000000000000000000000000006', 'base', 'WETH', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913', 'base', 'USDC', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x50c5725949A6F0c72E6C4a641F24049A917DB0Cb', 'base', 'DAI', 18, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0xfde4C96c8593536E31F229EA8f37b2ADa2699bb2', 'base', 'USDT', 6, 'erc20');
INSERT INTO tokens (address, chain_type, symbol, decimals, type) VALUES ('0x0555E30da8f98308EdB960aa94C0Db47230d2B9c', 'base', 'WBTC', 8, 'erc20');
