const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("manage-blacklist", "Blacklists or unblacklists an account on the BRLT contract")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("account", "The address of the account to manage blacklist status for")
  .addParam("action", "The action to perform: 'blacklist' or 'unblacklist'")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [blacklisterSigner] = await ethers.getSigners(); // Signer must have BLACKLISTER_ROLE

    const { proxy: proxyAddress, account, action } = taskArgs;

    if (action !== 'blacklist' && action !== 'unblacklist') {
      console.error("Error: Action must be 'blacklist' or 'unblacklist'.");
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

    console.log(`Executing manage-blacklist task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Target Account: ${account}`);
    console.log(`Action: ${action}`);
    console.log(`Signer (potential blacklister): ${blacklisterSigner.address}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, blacklisterSigner);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        const blacklisterRole = await brlt.BLACKLISTER_ROLE();
        const hasBlacklisterRole = await brlt.hasRole(blacklisterRole, blacklisterSigner.address);
        if (!hasBlacklisterRole) {
            console.error(`Error: Signer ${blacklisterSigner.address} does not have the BLACKLISTER_ROLE.`);
            console.error(`Please grant BLACKLISTER_ROLE to this address using the manage-roles task.`);
            throw new Error(`Signer ${blacklisterSigner.address} does not have BLACKLISTER_ROLE`);
        }

        const currentBlacklistStatus = await brlt.isBlacklisted(account);
        console.log(`Account ${account} is currently ${currentBlacklistStatus ? 'BLACKLISTED' : 'NOT BLACKLISTED'}.`);

        let tx;
        if (action === 'blacklist') {
            if (currentBlacklistStatus) {
                console.log(`Account ${account} is already blacklisted. No action needed.`);
                return;
            }
            console.log(`Blacklisting ${account}...`);
            tx = await brlt.blacklistAddress(account);
        } else { // action === 'unblacklist'
            if (!currentBlacklistStatus) {
                console.log(`Account ${account} is not blacklisted. No action needed for unblacklist.`);
                return;
            }
            console.log(`Unblacklisting ${account}...`);
            tx = await brlt.unblacklistAddress(account);
        }

        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const newBlacklistStatus = await brlt.isBlacklisted(account);
        console.log(`Account ${account} is now ${newBlacklistStatus ? 'BLACKLISTED' : 'NOT BLACKLISTED'}.`);
        console.log(`Successfully performed action: ${action} for account ${account}.`);

    } catch (error) {
        console.error(`Error during manage blacklist action:`, error);
        throw error;
    }
  }); 