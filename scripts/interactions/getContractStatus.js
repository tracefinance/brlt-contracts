const { task } = require("hardhat/config");
// const { ethers } = require("hardhat"); // Moved to where hre is available or inside action
const fs = require('fs');
const path = require('path');
// yargs is not needed for Hardhat tasks

task("get-status", "Gets the status of the BRLT contract")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addOptionalParam("checkBalances", "Comma-separated list of accounts to check balances for")
  .addOptionalParam("checkBlacklist", "Comma-separated list of accounts to check blacklist status for")
  .addOptionalParam("checkAllowances", "Comma-separated list of owner:spender pairs to check allowances for")
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre; // Get ethers from hre
    const [deployer] = await ethers.getSigners();

    const { proxy: proxyAddress, checkBalances, checkBlacklist, checkAllowances } = taskArgs;

    const accountsToCheckBalances = checkBalances ? checkBalances.split(',').map(a => a.trim()) : [];
    const accountsToCheckBlacklist = checkBlacklist ? checkBlacklist.split(',').map(a => a.trim()) : [];
    const allowancePairsToCheck = checkAllowances ? checkAllowances.split(',').map(p => {
        const parts = p.trim().split(':');
        if (parts.length !== 2) {
            console.error(`Invalid allowance pair format: ${p}. Expected owner:spender`);
            return null; // Or throw error
        }
        return { owner: parts[0].trim(), spender: parts[1].trim() };
    }).filter(p => p !== null) : [];

    console.log(`Executing get-status task for BRLT at proxy: ${proxyAddress}`);
    console.log(`Using network: ${hre.network.name}`);
    console.log(`Signer (for read-only calls): ${deployer.address}\n`);

    let brlt;
    try {
        // Use hre.artifacts if available, or construct path as before
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json"); // Assuming structure
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, deployer);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e; // Re-throw to fail the task
    }

    try {
        console.log("--- Basic Contract Info ---");
        console.log(`Name: ${await brlt.name()}`);
        console.log(`Symbol: ${await brlt.symbol()}`);
        console.log(`Decimals: ${await brlt.decimals()}`);
        const totalSupply = await brlt.totalSupply();
        console.log(`Total Supply: ${ethers.formatUnits(totalSupply, 18)} BRLT`);
        console.log(`Proxy Address: ${proxyAddress}`);
        console.log(`Paused State: ${await brlt.paused()}\n`);

        console.log("--- Roles (Hashes & Default Admin) ---");
        const roles = ["MINTER_ROLE", "BURNER_ROLE", "PAUSER_ROLE", "BLACKLISTER_ROLE", "UPGRADER_ROLE", "DEFAULT_ADMIN_ROLE"];
        for (const roleName of roles) {
            const roleHash = await brlt[roleName]();
            console.log(`${roleName}: ${roleHash}`);
            if (roleName === "DEFAULT_ADMIN_ROLE") {
                const deployerHasAdmin = await brlt.hasRole(roleHash, deployer.address);
                console.log(`  (Script signer ${deployer.address} has DEFAULT_ADMIN_ROLE: ${deployerHasAdmin})`);
            }
        }
        console.log("(Note: To get all members of a role, AccessControlEnumerable would be needed in the contract, or use off-chain indexing.)\n");

        if (accountsToCheckBalances.length > 0) {
            console.log("--- Account Balances ---");
            for (const account of accountsToCheckBalances) {
                try {
                    const balance = await brlt.balanceOf(account);
                    console.log(`Balance of ${account}: ${ethers.formatUnits(balance, 18)} BRLT`);
                } catch (e) {
                    console.warn(`Could not fetch balance for ${account}: ${e.message}`);
                }
            }
            console.log("");
        }

        if (accountsToCheckBlacklist.length > 0) {
            console.log("--- Account Blacklist Status ---");
            for (const account of accountsToCheckBlacklist) {
                try {
                    const status = await brlt.isBlacklisted(account);
                    console.log(`Blacklist status of ${account}: ${status}`);
                } catch (e) {
                    console.warn(`Could not fetch blacklist status for ${account}: ${e.message}`);
                }
            }
            console.log("");
        }

        if (allowancePairsToCheck.length > 0) {
            console.log("--- Account Allowances ---");
            for (const pair of allowancePairsToCheck) {
                try {
                    const allowance = await brlt.allowance(pair.owner, pair.spender);
                    console.log(`Allowance for ${pair.spender} from ${pair.owner}: ${ethers.formatUnits(allowance, 18)} BRLT`);
                } catch (e) {
                    console.warn(`Could not fetch allowance for ${pair.owner} -> ${pair.spender}: ${e.message}`);
                }
            }
            console.log("");
        }

        console.log("Contract status query completed.");

    } catch (error) {
        console.error(`Error querying contract status:`, error);
        throw error; // Re-throw to fail the task
    }
  });

// No main() or process.exit needed for tasks 