// console.log("Attempting to load BASE_TESTNET_RPC_URL:", process.env.BASE_TESTNET_RPC_URL); // Debug line - Commented out
require("@nomicfoundation/hardhat-toolbox");
require('@openzeppelin/hardhat-upgrades');

const fs = require('fs'); // Import fs for file system check
const envPath = process.env.NODE_ENV === 'test' ? '.env.test' : '.env';
const envFileExists = fs.existsSync(envPath);
console.log(`Dotenv: Attempting to load environment variables from: ${envPath}`);
console.log(`Dotenv: Does ${envPath} exist? ${envFileExists}`);

const dotenvResult = require('dotenv').config({
  path: envPath
});

if (dotenvResult.error) {
  console.error("Dotenv: Error loading .env file:", dotenvResult.error);
} else {
  console.log("Dotenv: Successfully loaded .env file. Parsed content (first few keys):");
  // Log only a few keys to avoid exposing too much, especially PRIVATE_KEY
  const keysToShow = ["BASE_TESTNET_RPC_URL", "BRLT_INITIAL_ADMIN", "NODE_ENV"]; // Add other relevant non-sensitive keys if needed
  for (const key of keysToShow) {
    if (dotenvResult.parsed && dotenvResult.parsed.hasOwnProperty(key)) {
      console.log(`  - Parsed ${key}: ${dotenvResult.parsed[key]}`);
    }
  }
  if (!dotenvResult.parsed || !dotenvResult.parsed.hasOwnProperty("BASE_TESTNET_RPC_URL")) {
    console.log("  - Dotenv: BASE_TESTNET_RPC_URL was NOT found in parsed .env content.");
  }
}

console.log("Dotenv: BASE_TESTNET_RPC_URL in process.env after dotenv load:", process.env.BASE_TESTNET_RPC_URL);
console.log("Dotenv: PRIVATE_KEY in process.env after dotenv load is set:", !!process.env.PRIVATE_KEY); // Check if private key is set
console.log("Dotenv: BASESCAN_API_KEY in process.env after dotenv load is set:", !!process.env.BASESCAN_API_KEY);
console.log("Dotenv: SEPOLIA_RPC_URL in process.env after dotenv load:", process.env.SEPOLIA_RPC_URL);

require("./scripts/deploy-task.js");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.28",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  paths: {
    sources: "./solidity",
  },
  networks: {
    // Base Mainnet
    base: {
      url: process.env.BASE_RPC_URL || "https://mainnet.base.org",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 8453,
    },
    // Base Goerli (testnet)
    "base-testnet": {
      url: process.env.BASE_TESTNET_RPC_URL || "https://sepolia.base.org",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 84532,
      maxPriorityFeePerGas: 2_000_000_000,
      maxFeePerGas: 30_000_000_000,
    },
    // Polygon zkEVM Mainnet
    "polygon-zkevm": {
      url: process.env.POLYGON_ZKEVM_RPC_URL || "https://zkevm-rpc.com",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 1101,
    },
    // Renaming polygon-zkevm-testnet to polygon-amoy
    "polygon-amoy": {
      url: process.env.POLYGON_AMOY_RPC_URL || "https://rpc-amoy.polygon.technology/",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 80002,
    },
    // Sepolia Testnet
    "sepolia": {
      url: process.env.SEPOLIA_RPC_URL || "https://rpc.sepolia.org",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 11155111,
    }
  },
  etherscan: {
    apiKey: {
      base: process.env.BASESCAN_API_KEY,
      "base-testnet": process.env.BASESCAN_API_KEY,
      "polygon-zkevm": process.env.POLYGONSCAN_API_KEY,
      "polygon-amoy": process.env.POLYGONSCAN_API_KEY,
      "sepolia": process.env.ETHERSCAN_API_KEY, // Assuming general Etherscan key for Sepolia
    },
    customChains: [
      {
        network: "base",
        chainId: 8453,
        urls: {
          apiURL: "https://api.basescan.org/api",
          browserURL: "https://basescan.org"
        }
      },
      {
        network: "base-testnet",
        chainId: 84532,
        urls: {
          apiURL: "https://api-sepolia.basescan.org/api",
          browserURL: "https://sepolia.basescan.org"
        }
      },
      {
        network: "polygon-zkevm",
        chainId: 1101,
        urls: {
          apiURL: "https://api-zkevm.polygonscan.com/api",
          browserURL: "https://zkevm.polygonscan.com"
        }
      },
      {
        network: "polygon-amoy",
        chainId: 80002,
        urls: {
          apiURL: "https://www.oklink.com/api/v5/explorer/contract/verify-source-code-plugin/AMOY_TESTNET",
          browserURL: "https://www.oklink.com/amoy"
        }
      },
      {
        network: "sepolia",
        chainId: 11155111,
        urls: {
          apiURL: "https://api-sepolia.etherscan.io/api",
          browserURL: "https://sepolia.etherscan.io"
        }
      }
    ]
  }
};
