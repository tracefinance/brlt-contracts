const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("query-brlt", "Queries BRLT contract information for a given proxy and account")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("account", "Address of the account to query information for")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const { proxy: proxyAddress, account: accountToQuery } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(accountToQuery)) {
      console.error(`Error: Invalid account address to query provided: ${accountToQuery}`);
      throw new Error("Invalid account address");
    }

    console.log(`Querying BRLT contract at ${proxyAddress} on network ${network.name}`);
    console.log(`Querying details for address: ${accountToQuery}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    console.log("--- Contract Info ---");
    const name = await brlt.name();
    const symbol = await brlt.symbol();
    const decimals = await brlt.decimals();
    const totalSupply = await brlt.totalSupply();

    console.log(`Name: ${name}`);
    console.log(`Symbol: ${symbol}`);
    console.log(`Decimals: ${decimals}`);
    console.log(`Total Supply: ${ethers.formatUnits(totalSupply, decimals)} ${symbol}`);

    console.log("\n--- Address Info ---  ");
    const balance = await brlt.balanceOf(accountToQuery);
    console.log(`Balance of ${accountToQuery}: ${ethers.formatUnits(balance, decimals)} ${symbol}`);

    console.log("\n--- Roles Info for Account --- ");
    const DEFAULT_ADMIN_ROLE = await brlt.DEFAULT_ADMIN_ROLE();
    const MINTER_ROLE = await brlt.MINTER_ROLE();
    const BURNER_ROLE = await brlt.BURNER_ROLE();
    const PAUSER_ROLE = await brlt.PAUSER_ROLE();
    const BLACKLISTER_ROLE = await brlt.BLACKLISTER_ROLE();
    const UPGRADER_ROLE = await brlt.UPGRADER_ROLE();

    const roles = [
      { name: "DEFAULT_ADMIN_ROLE", hash: DEFAULT_ADMIN_ROLE },
      { name: "MINTER_ROLE", hash: MINTER_ROLE },
      { name: "BURNER_ROLE", hash: BURNER_ROLE },
      { name: "PAUSER_ROLE", hash: PAUSER_ROLE },
      { name: "BLACKLISTER_ROLE", hash: BLACKLISTER_ROLE },
      { name: "UPGRADER_ROLE", hash: UPGRADER_ROLE },
    ];

    for (const role of roles) {
      const hasRole = await brlt.hasRole(role.hash, accountToQuery);
      console.log(`Has ${role.name} (${role.hash.substring(0,10)}...): ${hasRole}`);
    }
    console.log("\nQuery complete.");
  }); 