const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("manage-roles", "Grants or revokes a role for an account on the BRLT contract")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("role", "The name of the role (e.g., MINTER_ROLE)")
  .addParam("account", "The address of the account to manage the role for")
  .addParam("action", "The action to perform: 'grant' or 'revoke'")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [deployer] = await ethers.getSigners(); // Signer needs to have admin/appropriate role

    const { proxy: proxyAddress, role: roleName, account, action } = taskArgs;

    if (action !== 'grant' && action !== 'revoke') {
      console.error("Error: Action must be 'grant' or 'revoke'.");
      throw new Error("Invalid action parameter");
    }
    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(account)) {
      console.error(`Error: Invalid account address provided: ${account}`);
      throw new Error("Invalid account address");
    }

    console.log(`Executing manage-roles task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Action: ${action}`);
    console.log(`Role: ${roleName}`);
    console.log(`Account: ${account}`);
    console.log(`Signer (admin): ${deployer.address}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, deployer);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    let roleHash;
    try {
        // Ensure roleName is a valid key on the contract instance for fetching role hash
        if (typeof brlt[roleName] !== 'function') {
            throw new Error(`Role ${roleName} is not a recognized function on the contract to get its hash.`);
        }
        roleHash = await brlt[roleName]();
        if (!roleHash || typeof roleHash !== 'string' || !roleHash.startsWith('0x')) {
            throw new Error(`Role ${roleName} not found on the contract or returned an invalid hash: ${roleHash}`);
        }
    } catch (e) {
        console.error(`Error fetching role hash for ${roleName}. Ensure it's a valid role name (e.g., MINTER_ROLE, DEFAULT_ADMIN_ROLE) callable as a contract function.`);
        console.error(e.message);
        throw e;
    }
    
    console.log(`Successfully fetched ${roleName} hash: ${roleHash}`);

    try {
        let tx;
        const hasRoleBefore = await brlt.hasRole(roleHash, account);
        console.log(`Account ${account} ${hasRoleBefore ? 'HAS' : 'DOES NOT HAVE'} ${roleName} before action.`);

        if (action === 'grant') {
            if (hasRoleBefore) {
                console.log(`Account ${account} already has ${roleName}. No action needed.`);
                return;
            }
            console.log(`Granting ${roleName} to ${account}...`);
            tx = await brlt.grantRole(roleHash, account);
        } else { // action === 'revoke'
            if (!hasRoleBefore) {
                console.log(`Account ${account} does not have ${roleName}. Cannot revoke.`);
                return;
            }
            // Special handling for DEFAULT_ADMIN_ROLE: cannot revoke from the only admin if it's the signer.
            // This check might be more complex depending on how many admins there are.
            // For simplicity, this example doesn't prevent revoking the last admin, which could be dangerous.
            // Consider adding AccessControlEnumerable to check role member count if this is critical.
            // console.log("Skipping revoke for DEFAULT_ADMIN_ROLE from self if potentially last admin.");
            // return;
        } // This block requires AccessControlEnumerable, which BRLT.sol does not use.
          // So, this specific check for last admin removal cannot be reliably done here without it.

        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const hasRoleAfter = await brlt.hasRole(roleHash, account);
        console.log(`Account ${account} now ${hasRoleAfter ? 'HAS' : 'DOES NOT HAVE'} ${roleName}.`);
        console.log(`${action.charAt(0).toUpperCase() + action.slice(1)} role ${roleName} for account ${account} successful.`);

    } catch (error) {
        console.error(`Error during ${action} role ${roleName}:`, error);
        throw error;
    }
  }); 